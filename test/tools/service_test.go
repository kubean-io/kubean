package tools

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestAppendContainerStatuses(t *testing.T) {
	var diagnostics strings.Builder
	appendContainerStatuses(&diagnostics, "container", []corev1.ContainerStatus{
		{
			Name: "spray-job",
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					Reason:   "Error",
					ExitCode: 1,
					Message:  "kubeadm init failed",
				},
			},
		},
	})

	got := diagnostics.String()
	for _, expected := range []string{
		"container spray-job terminated",
		"reason=Error",
		"exitCode=1",
		`message="kubeadm init failed"`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected diagnostics to contain %q, got %q", expected, got)
		}
	}
}

func TestNormalizeKubeanJobLog(t *testing.T) {
	got := normalizeKubeanJobLog(`fatal: failed\n\t[ERROR SystemVerification]`)
	want := "fatal: failed\n\t[ERROR SystemVerification]"
	if got != want {
		t.Fatalf("normalizeKubeanJobLog() = %q, want %q", got, want)
	}
}

func TestBoundedKubeanJobLogPreservesBeginningAndEnd(t *testing.T) {
	got := boundedKubeanJobLog("first-cause--retry-failure", 20)
	for _, expected := range []string{"first-caus", "ry-failure", "6 log bytes omitted"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected bounded log to contain %q, got %q", expected, got)
		}
	}
}

func TestPingCommandForIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{name: "IPv4", ip: "192.168.136.6", want: "ping"},
		{name: "IPv6", ip: "fd89:ee78:d8a6:8608::1:6806", want: "ping6"},
		{name: "invalid", ip: "not-an-ip", want: "ping"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pingCommandForIP(tt.ip); got != tt.want {
				t.Fatalf("pingCommandForIP(%q) = %q, want %q", tt.ip, got, tt.want)
			}
		})
	}
}
