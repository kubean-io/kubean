// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"flag"
	"fmt"

	kubeanClusterOperationClientSet "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned"
	"github.com/kubean-io/kubean/pkg/version"
	clusteropswebhook "github.com/kubean-io/kubean/pkg/webhooks/clusterops"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	klog "k8s.io/klog/v2"
)

func NewCommand(ctx context.Context) *cobra.Command {
	opts := NewOptions()
	cmd := &cobra.Command{
		Use:  "kubean-admission",
		Long: "run admission controller for ClusterOperation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := opts.Validate(); len(errs) != 0 {
				return errs.ToAggregate()
			}
			return Run(ctx, opts)
		},
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of admission controller",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(version.Get())
		},
	}
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	cmd.AddCommand(versionCmd)
	opts.AddFlags(cmd.Flags())
	return cmd
}

func Run(ctx context.Context, opt *Options) error {
	klog.Warningf("Start Admission Controller")
	resetConfig, err := rest.InClusterConfig()
	if err != nil {
		// resetConfig, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			klog.ErrorS(err, "Failed to build kubernetes config")
			return err
		}
	}
	resetConfig.QPS = opt.KubeAPIQPS
	resetConfig.Burst = opt.KubeAPIBurst
	ClientSet, err := kubernetes.NewForConfig(resetConfig)
	if err != nil {
		return err
	}
	clusterClientOperationSet, err := kubeanClusterOperationClientSet.NewForConfig(resetConfig)
	if err != nil {
		return err
	}
	go func() {
		clusteropswebhook.CreateHTTPSCASecretWithLock(ctx, ClientSet)
	}()
	CASecret := clusteropswebhook.WaitForCASecretExist(ClientSet)
	if err := clusteropswebhook.CreateHTTPSCAFilesFromSecret(CASecret); err != nil {
		return err
	}
	clusteropswebhook.StartWebHookHTTPSServer(clusteropswebhook.PrepareWebHookHTTPSServer(clusterClientOperationSet))
	return fmt.Errorf("admission has exited")
}
