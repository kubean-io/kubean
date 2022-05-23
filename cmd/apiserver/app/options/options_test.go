package options

import (
	"reflect"
	"testing"

	apiserverconfig "github.com/daocloud/kubean/pkg/apiserver/config"
	"github.com/daocloud/kubean/pkg/provider/jenkins"
	"github.com/daocloud/kubean/pkg/store/db/connect"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/pflag"
)

func TestAddFlags(t *testing.T) {
	fs := pflag.NewFlagSet("addflagstest", pflag.ContinueOnError)
	s := NewAPIServerRunOptions()
	for _, f := range s.Flags().FlagSets {
		fs.AddFlagSet(f)
	}

	args := []string{
		"--bind-address=localhost",
		"--insecure-port=8080",
		"--secure-port=8443",
		"--tls-cert-file=/a/b/c.pub",
		"--tls-private-key=/a/b/d.key",
		"--debug",
		"--jenkins-host=http://localhost:30007",
		"--jenkins-max-connections=10",
		"--jenkins-username=admin",
		"--jenkins-password=password",
		"--jenkins-skip-verify",
		"--mysql-database=db1",
		"--mysql-host=localhost",
		"--mysql-port=3036",
		"--mysql-username=admin",
		"--mysql-password=password",
		"--mysql-skip-verify",
	}
	if err := fs.Parse(args); err != nil {
		t.Errorf("Failed to parse args: %v", err)
	}

	// This is a snapshot of expected options parsed by args.
	expected := &Options{
		ServerRunOptions: &ServerRunOptions{
			BindAddress:   "localhost",
			InsecurePort:  8080,
			SecurePort:    8443,
			TLSCertFile:   "/a/b/c.pub",
			TLSPrivateKey: "/a/b/d.key",
		},
		ConfigFile: "",
		Debug:      true,
		Config: &apiserverconfig.Config{
			DBOptions: &connect.Options{
				Database:   "db1",
				Host:       "localhost",
				Port:       "3036",
				User:       "admin",
				Password:   "password",
				SkipVerify: true,
			},
			JenkinsOptions: &jenkins.Options{
				Host:           "http://localhost:30007",
				MaxConnections: 10,
				Username:       "admin",
				Password:       "password",
				Namespace:      "kubean-system",
				SkipVerify:     true,
			},
		},
	}

	if !reflect.DeepEqual(s, expected) {
		t.Errorf("Got different run options than expected.\nDifference detected on:\n%s", cmp.Diff(expected, s, cmpopts.IgnoreUnexported()))
	}
}
