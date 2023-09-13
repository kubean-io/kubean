// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Options struct {
	KubeAPIQPS float32
	// KubeAPIBurst is the burst to allow while talking with kubean-apiserver.
	KubeAPIBurst int
}

func NewOptions() *Options {
	return &Options{
		KubeAPIQPS:   100.0,
		KubeAPIBurst: 100,
	}
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
	flags.Float32Var(&o.KubeAPIQPS, "kube-api-qps", 100.0, "QPS to use while talking with kubean-apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
	flags.IntVar(&o.KubeAPIBurst, "kube-api-burst", 100, "Burst to use while talking with kubean-apiserver. Doesn't cover events and node heartbeat apis which rate limiting is controlled by a different set of flags.")
}

func (o *Options) Validate() field.ErrorList {
	errs := field.ErrorList{}
	return errs
}
