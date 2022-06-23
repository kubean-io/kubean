package entrypoint

import (
	"strings"
	"testing"

	kubeanclusteropsv1alpha1 "kubean.io/api/apis/kubeanclusterops/v1alpha1"

	"gopkg.in/yaml.v2"
)

var testData = `
- message: "Check ansible passwordless mode"
  input:
    actionType: playbook
    action: cluster.yml
    isPrivateKey: true
    prehook:
      - actionType: shell
        action: |
          ansible -i host.yml -m shell -a "echo 'hello'|base64"
      - actionType: shell
        action: |
          ansible -i host.yml -m shell -a "docker info"
    posthook:
      - actionType: shell
        action: |
          ansible -i host.yml node1 -m shell -a "kubectl get node -o wide"
      - actionType: shell
        action: |
          ansible -i host.yml node1 -m shell -a "systemctl status kubelet"
  matchString: "--private-key"
  output: true

- message: "Check ansible password mode"
  input:
    actionType: playbook
    action: cluster.yml
    isPrivateKey: false
    prehook:
      - actionType: shell
        action: |
          ansible -i host.yml -m shell -a "echo 'hello'|base64"
    posthook:
      - actionType: shell
        action: |
          systemctl status kubelet
  matchString: "--private-key"
  output: false

- message: "Check kubespray reset parameter 'reset_confirmation'"
  input:
    actionType: playbook
    action: reset.yml
    isPrivateKey: false
    prehook:
      - actionType: shell
        action: |
          ansible -i host.yml -m shell -a "echo 'hello'|base64"
    posthook:
      - actionType: shell
        action: |
          systemctl status kubelet
  matchString: "reset_confirmation=yes"
  output: true

- message: "Check entrypoint script prehook part"
  input:
    actionType: playbook
    action: cluster.yml
    prehook:
      - actionType: shell
        action: |
          ansible -i host.yml -m ping
    posthook:
      - actionType: shell
        action: |
          systemctl status docker
      - actionType: shell
        action: |
          systemctl status kubelet
  matchString: "ansible -i host.yml -m ping"
  output: true

- message: "Check entrypoint script posthook part"
  input:
    actionType: playbook
    action: cluster.yml
    prehook:
      - actionType: shell
        action: |
          ansible -i host.yml -m shell -a "echo 'hello'|base64"
      - actionType: shell
        action: |
          systemctl status docker
    posthook:
      - actionType: shell
        action: |
          kubectl get cs
  matchString: "kubectl get cs"
  output: true

- message: "Check kubespray remove-node"
  input:
    actionType: playbook
    action: remove-node.yml
  matchString: "remove-node.yml"
  output: true

- message: "Check kubespray upgrade-cluster"
  input:
    actionType: playbook
    action: upgrade-cluster.yml
  matchString: "upgrade-cluster.yml"
  output: true

- message: "Check kubespray scale"
  input:
    actionType: playbook
    action: scale.yml
  matchString: "scale.yml"
  output: true

- message: "Check scale parameter 'extraArgs'"
  input:
    actionType: playbook
    action: scale.yml
    extraArgs: "--limit=node3,node4"
  matchString: "--limit=node3,node4"
  output: true

- message: "Check remove node parameter 'extraArgs'"
  input:
    actionType: playbook
    action: remove-node.yml
    extraArgs: "-e node=node3"
  matchString: "-e node=node3"
  output: true

- message: "Check postback kubeconf when the cluster is installed"
  input:
    actionType: playbook
    action: cluster.yml
    isPrivateKey: true
  matchString: "ansible -i $inventory_file $first_master -m fetch"
  output: true

- message: "Check clean up kubeconf after cluster reset"
  input:
    actionType: playbook
    action: reset.yml
  matchString: "{\"spec\": {\"kubeconfRef\": null}}"
  output: true

- message: "Check entrypoint script cmd concatenation"
  input:
    actionType: playbook
    action: cluster.yml
    prehook:
      - actionType: shell
        action: ansible -i host.yml -m ping
      - actionType: shell
        action: echo "test"
    posthook:
      - actionType: shell
        action: systemctl status docker
      - actionType: shell
        action: systemctl status kubelet
  matchString: "ansible -i host.yml -m ping\n"
  output: true

`

type SubAction struct {
	ActionType string `yaml:"actionType"`
	Action     string `yaml:"action"`
	ExtraArgs  string `yaml:"extraArgs"`
}

type ActionData struct {
	ActionType   string       `yaml:"actionType"`
	Action       string       `yaml:"action"`
	ExtraArgs    string       `yaml:"extraArgs"`
	PreHooks     []*SubAction `yaml:"prehook"`
	PostHooks    []*SubAction `yaml:"posthook"`
	IsPrivateKey bool         `yaml:"isPrivateKey"`
}

type UnitTestData struct {
	Message     string      `yaml:"message"`
	Input       *ActionData `yaml:"input"`
	MatchString string      `yaml:"matchString"`
	Output      bool        `yaml:"output"`
}

func TestEntrypoint(t *testing.T) {
	ad := []UnitTestData{}

	err := yaml.Unmarshal([]byte(testData), &ad)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	for _, item := range ad {
		t.Run(item.Message, func(t *testing.T) {
			ep := NewEntryPoint()
			// Prehook 命令处理
			for _, prehook := range item.Input.PreHooks {
				err = ep.PreHookRunPart(prehook.ActionType, prehook.Action, prehook.ExtraArgs, item.Input.IsPrivateKey)
				if err != nil {
					t.Fatalf("error: %v", err)
				}
			}
			// Kubespray 命令处理
			err = ep.SprayRunPart(item.Input.ActionType, item.Input.Action, item.Input.ExtraArgs, item.Input.IsPrivateKey)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			// Posthook 命令处理
			for _, posthook := range item.Input.PostHooks {
				err = ep.PostHookRunPart(posthook.ActionType, posthook.Action, posthook.ExtraArgs, item.Input.IsPrivateKey)
				if err != nil {
					t.Fatalf("error: %v", err)
				}
			}
			// 渲染 Entrypoint 脚本
			epScript, err := ep.Render()
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if strings.Contains(epScript, item.MatchString) != item.Output {
				t.Fatalf("entry point script render fail: \nentrypoint.sh: | \n%s\n", epScript)
			}

			t.Logf("entrypoint.sh: | \n%s\n", epScript)
		})
	}
}

func TestPBActionValue(t *testing.T) {
	if kubeanclusteropsv1alpha1.PlaybookActionType != PBAction {
		t.Fatal()
	}
	if kubeanclusteropsv1alpha1.ShellActionType != SHAction {
		t.Fatal()
	}
}
