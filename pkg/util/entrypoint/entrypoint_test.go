package entrypoint

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

var testData = `
- message: "Check ansible passwordless mode"
  input:
    actionType: playbook
    action: cluster.yml
    isPrivateKey: true
    prehook:
      actionType: shell
      action: |
        ansible -i host.yml -m shell -a "echo 'hello'|base64"
    posthook:
      actionType: shell
      action: |
        systemctl status kubelet
  matchString: "--private-key"
  output: true

- message: "Check ansible password mode"
  input:
    actionType: playbook
    action: cluster.yml
    isPrivateKey: false
    prehook:
      actionType: shell
      action: |
        ansible -i host.yml -m shell -a "echo 'hello'|base64"
    posthook:
      actionType: shell
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
      actionType: shell
      action: |
        ansible -i host.yml -m shell -a "echo 'hello'|base64"
    posthook:
      actionType: shell
      action: |
        systemctl status kubelet
  matchString: "reset_confirmation=yes"
  output: true

- message: "Check entrypoint script prehook part"
  input:
    actionType: playbook
    action: cluster.yml
    prehook:
      actionType: shell
      action: |
        ansible -i host.yml -m ping
    posthook:
      actionType: shell
      action: |
        systemctl status docker
        systemctl status kubelet
  matchString: "ansible -i host.yml -m ping"
  output: true

- message: "Check entrypoint script posthook part"
  input:
    actionType: playbook
    action: cluster.yml
    prehook:
      actionType: shell
      action: |
        ansible -i host.yml -m shell -a "echo 'hello'|base64"
        systemctl status docker
    posthook:
      actionType: shell
      action: |
        kubectl get cs
  matchString: "kubectl get cs"
  output: true
`

type SubAction struct {
	ActionType string `yaml:"actionType"`
	Action     string `yaml:"action"`
}

type ActionData struct {
	ActionType   string    `yaml:"actionType"`
	Action       string    `yaml:"action"`
	PreHook      SubAction `yaml:"prehook"`
	PostHook     SubAction `yaml:"posthook"`
	IsPrivateKey bool      `yaml:"isPrivateKey"`
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
			ep := &EntryPoint{}
			err = ep.PreHookRunPart(item.Input.PreHook.ActionType, item.Input.PreHook.Action)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			err = ep.SprayRunPart(item.Input.ActionType, item.Input.Action, item.Input.IsPrivateKey)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			err = ep.PostHookRunPart(item.Input.PostHook.ActionType, item.Input.PostHook.Action)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
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
