// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"testing"
)

func TestUnifyVersions(t *testing.T) {
	type args struct {
		versions []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal with v prefix",
			args: args{
				versions: []string{"v1.0.0", "v1.0.1", "v2.0.0"},
			},
			want: []string{"1.0.0", "1.0.1", "2.0.0"},
		},
		{
			name: "mixed v and no prefix",
			args: args{
				versions: []string{"v1.0.0", "1.0.0", "v2.0.0"},
			},
			want: []string{"1.0.0", "2.0.0"},
		},
		{
			name: "no v prefix",
			args: args{
				versions: []string{"1.0.0", "2.0.0"},
			},
			want: []string{"1.0.0", "2.0.0"},
		},
		{
			name: "duplicates only",
			args: args{
				versions: []string{"v1.0.0", "v1.0.0", "1.0.0"},
			},
			want: []string{"1.0.0"},
		},
		{
			name: "empty input",
			args: args{
				versions: []string{},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnifyVersions(tt.args.versions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnifyVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
