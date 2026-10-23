package tools

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
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

func TestNodePingCommand(t *testing.T) {
	shell, err := exec.LookPath("sh")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		ip       string
		tools    map[string]int
		want     string
		exitCode int
	}{
		{
			name:  "IPv4 keeps ping",
			ip:    "192.0.2.1",
			tools: map[string]int{"ping": 0, "ping6": 0},
			want:  "executed ping <-c> <1> <192.0.2.1>",
		},
		{
			name:  "IPv6 prefers ping6",
			ip:    "fd00::1",
			tools: map[string]int{"ping": 0, "ping6": 0},
			want:  "executed ping6 <-c> <1> <fd00::1>",
		},
		{
			name:  "IPv6 supports standalone ping6",
			ip:    "fd00::1",
			tools: map[string]int{"ping6": 0},
			want:  "executed ping6 <-c> <1> <fd00::1>",
		},
		{
			name:  "IPv6 supports unified ping",
			ip:    "fd00::1",
			tools: map[string]int{"ping": 0},
			want:  "executed ping <-6> <-c> <1> <fd00::1>",
		},
		{
			name:     "missing IPv6 tools fail explicitly",
			ip:       "fd00::1",
			want:     "IPv6 ping requires ping6 or ping on the target node",
			exitCode: 127,
		},
		{
			name:     "ping6 failure does not fall back to ping",
			ip:       "fd00::1",
			tools:    map[string]int{"ping": 0, "ping6": 1},
			want:     "executed ping6 <-c> <1> <fd00::1>",
			exitCode: 1,
		},
		{
			name:     "unified ping failure is preserved",
			ip:       "fd00::1",
			tools:    map[string]int{"ping": 2},
			want:     "executed ping <-6> <-c> <1> <fd00::1>",
			exitCode: 2,
		},
		{
			name:     "IPv4 failure is preserved",
			ip:       "192.0.2.1",
			tools:    map[string]int{"ping": 1, "ping6": 0},
			want:     "executed ping <-c> <1> <192.0.2.1>",
			exitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			for name, exitCode := range tt.tools {
				script := fmt.Sprintf("#!/bin/sh\nprintf 'executed %s'\nprintf ' <%%s>' \"$@\"\nprintf '\\n'\necho '%s diagnostic' >&2\nexit %d\n", name, name, exitCode)
				if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			command, err := nodePingCommand(tt.ip)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(shell, "-c", "cd \"$PING_TEST_BIN\" || exit 1\nPATH=\"$PWD\"\nexport PATH\n"+command)
			cmd.Env = append(os.Environ(), "PING_TEST_BIN="+filepath.ToSlash(binDir))
			output, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				exitCode = exitErr.ExitCode()
			}
			if exitCode != tt.exitCode {
				t.Fatalf("exit code = %d, want %d; output:\n%s", exitCode, tt.exitCode, output)
			}
			if !strings.Contains(string(output), tt.want) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.want, output)
			}
			wantExecutions := 1
			if len(tt.tools) == 0 {
				wantExecutions = 0
			}
			if got := strings.Count(string(output), "executed "); got != wantExecutions {
				t.Errorf("executed %d ping commands, want %d; output:\n%s", got, wantExecutions, output)
			}
		})
	}
}

func TestPingPodFromNode(t *testing.T) {
	networkErr := errors.New("exit status 1")
	sshErr := errors.New("exit status 255")
	toolErr := errors.New("exit status 127")
	const received = "1 packets transmitted, 1 received, 0% packet loss"
	tests := []struct {
		name       string
		ip         string
		output     string
		runErr     error
		failures   int
		wantCalls  int
		wantTime   time.Duration
		wantError  bool
		wantOutput string
	}{
		{name: "success", ip: "fd00::1", wantCalls: 1},
		{
			name: "transient failure recovers", ip: "fd00::1",
			output: "using ping6 for IPv6\nnetwork unreachable", runErr: networkErr, failures: 1,
			wantCalls: 2, wantTime: 5 * time.Second,
		},
		{
			name: "network failure is preserved", ip: "fd00::1",
			output: "using ping6 for IPv6\nnetwork unreachable", runErr: networkErr, failures: 4,
			wantCalls: 4, wantTime: 15 * time.Second, wantError: true, wantOutput: "network unreachable",
		},
		{
			name: "SSH failure is preserved", ip: "fd00::1",
			output: "ssh: connection refused", runErr: sshErr, failures: 4,
			wantCalls: 4, wantTime: 15 * time.Second, wantError: true, wantOutput: "ssh: connection refused",
		},
		{
			name: "missing tools are reported", ip: "fd00::1",
			output: "IPv6 ping requires ping6 or ping on the target node", runErr: toolErr, failures: 4,
			wantCalls: 4, wantTime: 15 * time.Second, wantError: true, wantOutput: "IPv6 ping requires ping6 or ping",
		},
		{
			name: "zero received packets fail", ip: "fd00::1",
			output: "1 packets transmitted, 0 received", failures: 1,
			wantCalls: 1, wantError: true, wantOutput: "did not report a received packet",
		},
		{
			name: "invalid address is rejected before SSH", ip: "not-an-ip",
			wantError: true, wantOutput: "invalid pod IP address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				start := time.Now()
				err := pingPodFromNode("root@192.0.2.10", tt.ip, func(string) ([]byte, error) {
					calls++
					if calls <= tt.failures {
						return []byte(tt.output), tt.runErr
					}
					return []byte(received), nil
				})
				if calls != tt.wantCalls || time.Since(start) != tt.wantTime {
					t.Fatalf("calls=%d elapsed=%v, want calls=%d elapsed=%v", calls, time.Since(start), tt.wantCalls, tt.wantTime)
				}
				if !tt.wantError {
					if err != nil {
						t.Fatal(err)
					}
					return
				}
				if err == nil {
					t.Fatal("expected an error")
				}
				if tt.runErr != nil && !errors.Is(err, tt.runErr) {
					t.Errorf("expected original command error %v, got %v", tt.runErr, err)
				}
				for _, want := range []string{"root@192.0.2.10", tt.ip, tt.wantOutput} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("expected error to contain %q, got %v", want, err)
					}
				}
			})
		})
	}
}

func TestNodePingPodByPasswdReportsInvalidIP(t *testing.T) {
	previousGomega := gomega.Default
	gomega.Default = gomega.NewWithT(t)
	defer func() { gomega.Default = previousGomega }()

	failures := gomega.InterceptGomegaFailures(func() {
		NodePingPodByPasswd("unused", "root@192.0.2.10", "not-an-ip")
	})
	if len(failures) != 1 {
		t.Fatalf("got %d assertion failures, want one: %v", len(failures), failures)
	}
	if !strings.Contains(failures[0], "invalid pod IP address") {
		t.Fatalf("expected invalid IP diagnostics, got %s", failures[0])
	}
}

func TestWaitKubeSystemPodsReadyRefreshesPods(t *testing.T) {
	healthy := kubeSystemPodForTest("coredns", corev1.PodRunning, corev1.ConditionTrue)
	pending := kubeSystemPodForTest("nginx-proxy-node2", corev1.PodPending, corev1.ConditionFalse)
	pending.Spec.NodeName = "node2"
	replacement := kubeSystemPodForTest("coredns-replacement", corev1.PodRunning, corev1.ConditionFalse)
	readyReplacement := replacement.DeepCopy()
	readyReplacement.Status.Conditions[0].Status = corev1.ConditionTrue
	unready := healthy.DeepCopy()
	unready.Status.Conditions[0].Status = corev1.ConditionFalse
	terminating := pending.DeepCopy()
	terminating.Status.Phase = corev1.PodRunning
	terminating.Status.Conditions[0].Status = corev1.ConditionTrue
	deletedAt := metav1.NewTime(time.Unix(1, 0))
	terminating.DeletionTimestamp = &deletedAt

	tests := []struct {
		name      string
		snapshots [][]corev1.Pod
	}{
		{
			name:      "already healthy",
			snapshots: [][]corev1.Pod{{healthy}},
		},
		{
			name:      "removed worker pod disappears",
			snapshots: [][]corev1.Pod{{healthy, pending}, {healthy}},
		},
		{
			name:      "replacement must become ready",
			snapshots: [][]corev1.Pod{{healthy, pending}, {healthy, replacement}, {healthy, *readyReplacement}},
		},
		{
			name:      "previously healthy pod becomes unready",
			snapshots: [][]corev1.Pod{{healthy, pending}, {*unready}, {healthy}},
		},
		{
			name:      "new pod appears while waiting",
			snapshots: [][]corev1.Pod{{pending}, {healthy, replacement}, {healthy, *readyReplacement}},
		},
		{
			name:      "terminating pod must disappear",
			snapshots: [][]corev1.Pod{{healthy, *terminating}, {healthy}},
		},
		{
			name:      "empty list is not healthy",
			snapshots: [][]corev1.Pod{{}, {healthy}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				client := fake.NewClientset()
				calls := 0
				client.PrependReactor("list", "pods", func(action clienttesting.Action) (bool, runtime.Object, error) {
					snapshot := tt.snapshots[min(calls, len(tt.snapshots)-1)]
					calls++
					return true, &corev1.PodList{Items: snapshot}, nil
				})

				start := time.Now()
				if err := waitKubeSystemPodsReady(context.Background(), client, 10*time.Second, time.Second); err != nil {
					t.Fatal(err)
				}
				if calls != len(tt.snapshots) {
					t.Fatalf("listed pods %d times, want %d", calls, len(tt.snapshots))
				}
				if elapsed, want := time.Since(start), time.Duration(len(tt.snapshots)-1)*time.Second; elapsed != want {
					t.Fatalf("waited %v, want %v", elapsed, want)
				}
				for _, action := range client.Actions() {
					if action.GetVerb() != "list" || action.GetNamespace() != KubeSystemNamespace {
						t.Fatalf("unexpected API action: %#v", action)
					}
				}
			})
		})
	}
}

func TestWaitKubeSystemPodsReadyRejectsUnhealthyPods(t *testing.T) {
	pending := kubeSystemPodForTest("pending", corev1.PodPending, corev1.ConditionFalse)
	unready := kubeSystemPodForTest("unready", corev1.PodRunning, corev1.ConditionFalse)
	unready.Status.Conditions[0].Reason = "ContainersNotReady"
	unready.Status.Conditions[0].Message = "containers with unready status: [kube-proxy]"
	unready.Status.ContainerStatuses = []corev1.ContainerStatus{{
		Name: "kube-proxy",
		State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff", Message: "back-off restarting failed container"},
		},
	}}
	unknown := kubeSystemPodForTest("unknown", corev1.PodRunning, corev1.ConditionUnknown)
	unknown.Status.Conditions = nil
	terminating := kubeSystemPodForTest("terminating", corev1.PodRunning, corev1.ConditionTrue)
	deletedAt := metav1.NewTime(time.Unix(1, 0))
	terminating.DeletionTimestamp = &deletedAt

	tests := []struct {
		name string
		pods []corev1.Pod
		want []string
	}{
		{name: "empty list", want: []string{"no pods found in kube-system"}},
		{name: "pending", pods: []corev1.Pod{pending}, want: []string{"kube-system/pending", "phase=Pending"}},
		{name: "running but unready", pods: []corev1.Pod{unready}, want: []string{"Ready=False", "ContainersNotReady", "CrashLoopBackOff", "kube-proxy"}},
		{name: "readiness not reported", pods: []corev1.Pod{unknown}, want: []string{"Ready=Unknown"}},
		{name: "terminating", pods: []corev1.Pod{terminating}, want: []string{"terminating=true"}},
		{
			name: "failed",
			pods: []corev1.Pod{kubeSystemPodForTest("failed", corev1.PodFailed, corev1.ConditionFalse)},
			want: []string{"phase=Failed"},
		},
		{
			name: "succeeded is not running",
			pods: []corev1.Pod{kubeSystemPodForTest("succeeded", corev1.PodSucceeded, corev1.ConditionFalse)},
			want: []string{"phase=Succeeded"},
		},
		{name: "all unhealthy pods are reported", pods: []corev1.Pod{pending, unready}, want: []string{"kube-system/pending", "kube-system/unready"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				objects := make([]runtime.Object, len(tt.pods))
				for i := range tt.pods {
					objects[i] = tt.pods[i].DeepCopy()
				}
				client := fake.NewClientset(objects...)

				start := time.Now()
				err := waitKubeSystemPodsReady(context.Background(), client, 3*time.Second, time.Second)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("got error %v, want context deadline exceeded", err)
				}
				for _, want := range tt.want {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("expected error to contain %q, got %v", want, err)
					}
				}
				if elapsed := time.Since(start); elapsed != 3*time.Second {
					t.Fatalf("waited %v, want one 3s deadline for all pods", elapsed)
				}
			})
		})
	}
}

func TestWaitKubeSystemPodsReadyRetriesListErrors(t *testing.T) {
	for _, recoverAfterError := range []bool{false, true} {
		name := "persistent error"
		if recoverAfterError {
			name = "transient error"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				client := fake.NewClientset()
				calls := 0
				client.PrependReactor("list", "pods", func(action clienttesting.Action) (bool, runtime.Object, error) {
					calls++
					if recoverAfterError && calls > 1 {
						return true, &corev1.PodList{Items: []corev1.Pod{
							kubeSystemPodForTest("coredns", corev1.PodRunning, corev1.ConditionTrue),
						}}, nil
					}
					return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "pods"}, "", errors.New("list denied"))
				})

				start := time.Now()
				err := waitKubeSystemPodsReady(context.Background(), client, 3*time.Second, time.Second)
				if recoverAfterError {
					if err != nil || calls != 2 || time.Since(start) != time.Second {
						t.Fatalf("transient error did not recover after one poll: error=%v calls=%d elapsed=%v", err, calls, time.Since(start))
					}
				} else {
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("got error %v, want context deadline exceeded", err)
					}
					for _, want := range []string{"list pods in kube-system", "forbidden", "list denied"} {
						if !strings.Contains(err.Error(), want) {
							t.Errorf("expected error to contain %q, got %v", want, err)
						}
					}
					if elapsed := time.Since(start); elapsed != 3*time.Second {
						t.Fatalf("waited %v, want 3s", elapsed)
					}
				}
			})
		})
	}
}

func TestWaitPodSInKubeSystemBeRunningBoundsAPIRequests(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		client, err := kubernetes.NewForConfig(&restclient.Config{
			Host: "https://kube-system.test",
			QPS:  -1,
			Transport: kubeSystemRoundTripper(func(request *http.Request) (*http.Response, error) {
				calls++
				if request.URL.Path != "/api/v1/namespaces/kube-system/pods" {
					t.Errorf("unexpected request path: %s", request.URL.Path)
				}
				<-request.Context().Done()
				return nil, request.Context().Err()
			}),
		})
		if err != nil {
			t.Fatal(err)
		}

		previousGomega := gomega.Default
		gomega.Default = gomega.NewWithT(t)
		defer func() { gomega.Default = previousGomega }()

		start := time.Now()
		failures := gomega.InterceptGomegaFailures(func() {
			WaitPodSInKubeSystemBeRunning(client, 2, 60)
		})
		if len(failures) != 1 {
			t.Fatalf("got %d assertion failures, want one: %v", len(failures), failures)
		}
		for _, want := range []string{"context deadline exceeded", "list pods in kube-system"} {
			if !strings.Contains(failures[0], want) {
				t.Errorf("expected failure to contain %q, got %s", want, failures[0])
			}
		}
		if calls != 1 || time.Since(start) != 2*time.Second {
			t.Fatalf("API request exceeded the total deadline: calls=%d elapsed=%v", calls, time.Since(start))
		}
	})
}

func kubeSystemPodForTest(name string, phase corev1.PodPhase, ready corev1.ConditionStatus) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: KubeSystemNamespace},
		Spec:       corev1.PodSpec{NodeName: "node1"},
		Status: corev1.PodStatus{
			Phase:      phase,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: ready}},
		},
	}
}

type kubeSystemRoundTripper func(*http.Request) (*http.Response, error)

func (f kubeSystemRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
