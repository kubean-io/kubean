// Copyright 2023 Authors of kubean-io
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestWriteFile(t *testing.T) {
	type args struct {
		filepath string
		bytes    []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create and write tmp file",
			args: args{
				filepath: "/tmp/tmp.txt",
				bytes:    []byte("test"),
			},
			wantErr: false,
		},
		{
			name: "create and write file on read-only filesystem",
			args: args{
				filepath: "/proc/tmp.txt",
				bytes:    []byte("test"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WriteFile(tt.args.filepath, tt.args.bytes); (err != nil) != tt.wantErr {
				t.Errorf("WriteFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsExist(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "file exists",
			args: args{
				filepath: "/etc/hosts",
			},
			want: true,
		},
		{
			name: "file does not exist",
			args: args{
				filepath: "/tmp/tmp1.txt",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExist(tt.args.filepath); tt.want != got {
				t.Errorf("IsExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
