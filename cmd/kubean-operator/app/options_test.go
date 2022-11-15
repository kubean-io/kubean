package app

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/component-base/config"
)

func TestOptions_Validate(t *testing.T) {
	type fields struct {
		LeaderElection config.LeaderElectionConfiguration
		BindAddress    string
		SecurePort     int
		KubeAPIQPS     float32
		KubeAPIBurst   int
	}
	tests := []struct {
		name   string
		fields fields
		want   field.ErrorList
	}{
		{
			name: "test invalid secure port less than 0",
			fields: fields{
				SecurePort: -1,
			},
			want: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "Options.SecurePort",
					BadValue: -1,
					Detail:   "must be between 0 and 65535 inclusive",
				},
			},
		},
		{
			name: "test invalid secure port more than 65535",
			fields: fields{
				SecurePort: 65536,
			},
			want: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "Options.SecurePort",
					BadValue: 65536,
					Detail:   "must be between 0 and 65535 inclusive",
				},
			},
		},
		{
			name: "test valid pass",
			fields: fields{
				SecurePort: 65535,
			},
			want: field.ErrorList{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Options{
				LeaderElection: tt.fields.LeaderElection,
				BindAddress:    tt.fields.BindAddress,
				SecurePort:     tt.fields.SecurePort,
				KubeAPIQPS:     tt.fields.KubeAPIQPS,
				KubeAPIBurst:   tt.fields.KubeAPIBurst,
			}
			if got := o.Validate(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
