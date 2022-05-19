package options

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/daocloud/kubean/pkg/apiserver"
	apiserverconfig "github.com/daocloud/kubean/pkg/apiserver/config"
	"github.com/daocloud/kubean/pkg/provider/jenkins"
	"github.com/daocloud/kubean/pkg/store/db/service"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type ServerRunOptions struct {
	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// tls cert file
	TLSCertFile string

	// tls private key file
	TLSPrivateKey string
}

func NewServerRunOptions() *ServerRunOptions {
	// create default server run options
	s := ServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  8000,
		SecurePort:    0,
		TLSCertFile:   "",
		TLSPrivateKey: "",
	}

	return &s
}

type Options struct {
	ConfigFile       string
	ServerRunOptions *ServerRunOptions
	*apiserverconfig.Config

	// Debug indicates kpanda apiserver mode is debug.
	Debug bool
}

func NewAPIServerRunOptions() *Options {
	s := &Options{
		ServerRunOptions: NewServerRunOptions(),
		Config:           apiserverconfig.New(),
	}
	return s
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet, c *ServerRunOptions) {
	fs.StringVar(&s.BindAddress, "bind-address", c.BindAddress, "server bind address")
	fs.IntVar(&s.InsecurePort, "insecure-port", c.InsecurePort, "insecure port number")
	fs.IntVar(&s.SecurePort, "secure-port", s.SecurePort, "secure port number")
	fs.StringVar(&s.TLSCertFile, "tls-cert-file", c.TLSCertFile, "tls cert file")
	fs.StringVar(&s.TLSPrivateKey, "tls-private-key", c.TLSPrivateKey, "tls private key")
}

func (s *Options) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.Debug, "debug", false, "apiserver server mode")
	s.ServerRunOptions.AddFlags(fs, s.ServerRunOptions)
	s.DBOptions.AddFlags(fss.FlagSet("mysql"), s.DBOptions)
	s.JenkinsOptions.AddFlags(fss.FlagSet("jenkins"), s.JenkinsOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

// NewAPIServer creates an APIServer instance using given options.
func (s *Options) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Debug:  s.Debug,
		Config: s.Config,
	}

	// Create the main listener.
	address := fmt.Sprintf("%s:%d", s.ServerRunOptions.BindAddress, s.ServerRunOptions.InsecurePort)

	// Create your protocol servers.
	apiServer.HttpServer = &http.Server{
		Addr: address,
	}

	if s.DBOptions.Host != "" {
		services, err := service.NewServices(s.DBOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to db: %v", err)
		}
		apiServer.DBRepository = services.PipelineService
	}

	if !s.JenkinsOptions.SkipVerify && s.JenkinsOptions.Host != "" {
		jenkinsProvider, err := jenkins.NewJenkinsProvider(s.JenkinsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to jenkins, please check jenkins status, error: %v", err)
		}
		apiServer.JenkinsProvider = jenkinsProvider
	}

	return apiServer, nil
}
