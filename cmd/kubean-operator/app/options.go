// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	componentbaseconfig "k8s.io/component-base/config"

	"github.com/kubean-io/kubean/pkg/util"
)

const (
	defaultBindAddress = "0.0.0.0"
	defaultPort        = 20358
)

type Options struct {
	LeaderElection componentbaseconfig.LeaderElectionConfiguration
	// KubeAPIQPS is the QPS to use while talking with kubean-apiserver.
	BindAddress string
	// SecurePort is the port that the server serves at.
	SecurePort int

	KubeAPIQPS float32
	// KubeAPIBurst is the burst to allow while talking with kubean-apiserver.
	KubeAPIBurst int
}

func NewOptions() *Options {
	return &Options{
		LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
			LeaderElect:       true,
			ResourceLock:      resourcelock.LeasesResourceLock,
			ResourceNamespace: util.GetCurrentNSOrDefault(),
			ResourceName:      "kubean-controller",
		},
		KubeAPIQPS:   100.0,
		KubeAPIBurst: 100,
	}
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.BindAddress, "bind-address", defaultBindAddress,
		"The IP address on which to listen for the --secure-port port.")
	flags.IntVar(&o.SecurePort, "secure-port", defaultPort,
		"The secure port on which to serve HTTPS.")
	flags.BoolVar(&o.LeaderElection.LeaderElect, "leader-elect", true, "Start a leader election client and gain leadership before executing the main loop. Enable this when running replicated components for high availability.")
	flags.StringVar(&o.LeaderElection.ResourceNamespace, "leader-elect-resource-namespace", "default", "The namespace of resource object that is used for locking during leader election.")
	flags.Float32Var(&o.KubeAPIQPS, "kube-api-qps", 100.0, "QPS to use while talking with kubean-apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
	flags.IntVar(&o.KubeAPIBurst, "kube-api-burst", 100, "Burst to use while talking with kubean-apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
}

func (o *Options) Validate() field.ErrorList {
	errs := field.ErrorList{}
	newPath := field.NewPath("Options")
	if o.SecurePort < 0 || o.SecurePort > 65535 {
		errs = append(errs, field.Invalid(newPath.Child("SecurePort"), o.SecurePort, "must be between 0 and 65535 inclusive"))
	}
	return errs
}
