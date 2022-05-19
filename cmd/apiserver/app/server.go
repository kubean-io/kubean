package app

import (
	"context"

	"github.com/daocloud/kubean"
	"github.com/daocloud/kubean/cmd/apiserver/app/options"
	apiserverconfig "github.com/daocloud/kubean/pkg/apiserver/config"
	"github.com/daocloud/kubean/pkg/version"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewAPIServerCommand(ctx context.Context) *cobra.Command {
	s := options.NewAPIServerRunOptions()

	// Load configuration from file
	conf, err := apiserverconfig.TryLoadFromDisk()
	if err == nil {
		s = &options.Options{
			ServerRunOptions: s.ServerRunOptions,
			Config:           conf,
		}
	} else {
		if _, ok := err.(kubean.ImproperlyConfiguredError); ok {
			klog.Infof("No configuration file found. Using default configuration")
		} else {
			klog.Fatal("Failed to load configuration from disk", err)
		}
	}

	cmd := &cobra.Command{
		Use: "kubean-apiserver",
		Long: `The kubean API server validates and configures data for the API objects. 
The API Server services REST operations and provides the frontend to the
cluster's shared state through which all other components interact.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := s.Validate(); len(errs) != 0 {
				return utilerrors.NewAggregate(errs)
			}

			return Run(signals.SetupSignalHandler(), s)
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of kpanda-apiserver",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func Run(ctx context.Context, s *options.Options) error {
	apiserver, err := s.NewAPIServer(ctx.Done())
	if err != nil {
		return err
	}

	err = apiserver.PrepareRun(ctx)
	if err != nil {
		return err
	}

	return apiserver.Run(ctx)
}
