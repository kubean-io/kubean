// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"flag"
	"net"
	"os"
	"strconv"
	"time"

	kubeanClusterClientSet "github.com/kubean-io/kubean-api/generated/cluster/clientset/versioned"
	kubeanClusterOperationClientSet "github.com/kubean-io/kubean-api/generated/clusteroperation/clientset/versioned"
	kubeanLocalArtifactSetClientSet "github.com/kubean-io/kubean-api/generated/localartifactset/clientset/versioned"
	kubeaninfomanifestClientSet "github.com/kubean-io/kubean-api/generated/manifest/clientset/versioned"

	"github.com/kubean-io/kubean/pkg/controllers/cluster"
	"github.com/kubean-io/kubean/pkg/controllers/clusterops"
	"github.com/kubean-io/kubean/pkg/controllers/infomanifest"
	"github.com/kubean-io/kubean/pkg/controllers/offlineversion"
	"github.com/kubean-io/kubean/pkg/util"
	"github.com/kubean-io/kubean/pkg/version"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func NewCommand(ctx context.Context) *cobra.Command {
	opts := NewOptions()
	cmd := &cobra.Command{
		Use:  "kubean-operator",
		Long: "run operator for Cluster and ClusterOperation",
		RunE: func(cmd *cobra.Command, args []string) error {
			if errs := opts.Validate(); len(errs) != 0 {
				return errs.ToAggregate()
			}
			return Run(ctx, opts)
		},
	}
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of controller manager",
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
	klog.Warningf("Start KuBean Controller")
	if err := StartManager(ctx, opt); err != nil {
		return err
	}
	return nil
}

// StartManager will block.
func StartManager(ctx context.Context, opt *Options) error {
	config, err := controllerruntime.GetConfig()
	if err != nil {
		return err
	}
	config.QPS, config.Burst = opt.KubeAPIQPS, opt.KubeAPIBurst
	controllerManager, err := controllerruntime.NewManager(config, controllerruntime.Options{
		Scheme:                     util.NewSchema(), // register schema
		LeaderElection:             opt.LeaderElection.LeaderElect,
		LeaderElectionID:           opt.LeaderElection.ResourceName,
		LeaderElectionNamespace:    opt.LeaderElection.ResourceNamespace,
		LeaderElectionResourceLock: opt.LeaderElection.ResourceLock,
		HealthProbeBindAddress:     net.JoinHostPort(opt.BindAddress, strconv.Itoa(opt.SecurePort)),
		LivenessEndpointName:       "/healthz",
	})
	if err != nil {
		klog.Errorf("Failed to build controllerManager ,%s", err)
		return err
	}
	if err := controllerManager.AddHealthzCheck("ping", healthz.Ping); err != nil {
		klog.Errorf("Failed to add health check endpoint: %s", err)
		return err
	}
	ClientSet, clusterClientSet, clusterClientOperationSet, infomanifestClientSet, localArtifactSetClientSet, err := prepareClient()
	if err != nil {
		klog.ErrorS(err, "create clientSet")
		return err
	}
	if err := setupManager(controllerManager, ClientSet, clusterClientSet, clusterClientOperationSet, infomanifestClientSet, localArtifactSetClientSet, opt, ctx.Done()); err != nil {
		klog.Errorf("setupManager %s", err)
		return err
	}
	go func() {
		for {
			if err := clusterops.RunWithLeaseLock(ctx, ClientSet, clusterClientOperationSet); err != nil {
				klog.ErrorS(err, "run webhook with LeaseLock")
			}
			time.Sleep(time.Second * 5)
		}
	}()
	if err := controllerManager.Start(ctx); err != nil {
		klog.Errorf("KubeanOperator ControllerManager exit ,%s", err)
		return err
	}
	return nil
}

func prepareClient() (kubernetes.Interface, kubeanClusterClientSet.Interface, kubeanClusterOperationClientSet.Interface, kubeaninfomanifestClientSet.Interface, kubeanLocalArtifactSetClientSet.Interface, error) {
	resetConfig, err := rest.InClusterConfig()
	if err != nil {
		resetConfig, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	ClientSet, err := kubernetes.NewForConfig(resetConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	clusterClientSet, err := kubeanClusterClientSet.NewForConfig(resetConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	clusterClientOperationSet, err := kubeanClusterOperationClientSet.NewForConfig(resetConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	infomanifestClientSet, err := kubeaninfomanifestClientSet.NewForConfig(resetConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	localArtifactSetClientSet, err := kubeanLocalArtifactSetClientSet.NewForConfig(resetConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return ClientSet, clusterClientSet, clusterClientOperationSet, infomanifestClientSet, localArtifactSetClientSet, nil
}

func setupManager(
	mgr controllerruntime.Manager,
	ClientSet kubernetes.Interface,
	clusterClientSet kubeanClusterClientSet.Interface,
	clusterClientOperationSet kubeanClusterOperationClientSet.Interface,
	infomanifestClientSet kubeaninfomanifestClientSet.Interface,
	localArtifactSetClientSet kubeanLocalArtifactSetClientSet.Interface,
	opt *Options, stopChan <-chan struct{},
) error {
	clusterController := &cluster.Controller{
		Client:              mgr.GetClient(),
		ClientSet:           ClientSet,
		KubeanClusterSet:    clusterClientSet,
		KubeanClusterOpsSet: clusterClientOperationSet,
	}
	// the message type
	if err := clusterController.SetupWithManager(mgr); err != nil {
		klog.Errorf("ControllerManager Cluster but %s", err)
		return err
	}
	clusterOpsController := &clusterops.Controller{
		Client:                mgr.GetClient(),
		ClientSet:             ClientSet,
		KubeanClusterSet:      clusterClientSet,
		KubeanClusterOpsSet:   clusterClientOperationSet,
		InfoManifestClientSet: infomanifestClientSet,
	}
	if err := clusterOpsController.SetupWithManager(mgr); err != nil {
		klog.Errorf("ControllerManager ClusterOps but %s", err)
		return err
	}

	offlineVersionController := &offlineversion.Controller{
		Client:                    mgr.GetClient(),
		ClientSet:                 ClientSet,
		InfoManifestClientSet:     infomanifestClientSet,
		LocalArtifactSetClientSet: localArtifactSetClientSet,
	}
	if err := offlineVersionController.SetupWithManager(mgr); err != nil {
		klog.Errorf("ControllerManager OfflineVersion but %s", err)
		return err
	}

	infomanifestController := &infomanifest.Controller{
		Client:                    mgr.GetClient(),
		InfoManifestClientSet:     infomanifestClientSet,
		ClientSet:                 ClientSet,
		LocalArtifactSetClientSet: localArtifactSetClientSet,
	}
	if err := infomanifestController.SetupWithManager(mgr); err != nil {
		klog.Errorf("ControllerManager Infomanifest but %s", err)
		return err
	}
	return nil
}
