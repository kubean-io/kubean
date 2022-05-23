package options

import (
	"fmt"
	"os"

	utilnet "github.com/daocloud/kubean/pkg/util/net"
)

// Validate validates server run options, to find
// options' misconfiguration.
func (s *Options) Validate() []error {
	var errors []error

	errors = append(errors, s.ServerRunOptions.Validate()...)
	errors = append(errors, s.JenkinsOptions.Validate()...)
	errors = append(errors, s.DBOptions.Validate()...)

	return errors
}

func (s *ServerRunOptions) Validate() []error {
	var errs []error

	if s.SecurePort == 0 && s.InsecurePort == 0 {
		errs = append(errs, fmt.Errorf("insecure and secure port can not be disabled at the same time"))
	}
	if utilnet.IsValidPort(s.SecurePort) {
		if s.TLSCertFile == "" {
			errs = append(errs, fmt.Errorf("tls cert file is empty while secure serving"))
		}
		if _, err := os.Stat(s.TLSCertFile); err != nil {
			errs = append(errs, err)
		}
		if s.TLSPrivateKey == "" {
			errs = append(errs, fmt.Errorf("tls private key file is empty while secure serving"))
		}
		if _, err := os.Stat(s.TLSPrivateKey); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
