package entrypoint

import (
	"strings"
	"testing"

	clusteroperationv1alpha1 "kubean.io/api/apis/clusteroperation/v1alpha1"

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
  matchString: "-e \"skip_confirmation=true\" /kubespray/remove-node.yml"
  output: true

- message: "Check kubespray upgrade-cluster"
  input:
    actionType: playbook
    action: upgrade-cluster.yml
  matchString: "/kubespray/upgrade-cluster.yml"
  output: true

- message: "Check kubespray scale"
  input:
    actionType: playbook
    action: scale.yml
  matchString: "/kubespray/scale.yml"
  output: true

- message: "Check scale parameter 'extraArgs'"
  input:
    actionType: playbook
    action: scale.yml
    extraArgs: "--limit=node3,node4"
  matchString: "/kubespray/scale.yml --limit=node3,node4"
  output: true

- message: "Check remove node parameter 'extraArgs'"
  input:
    actionType: playbook
    action: remove-node.yml
    extraArgs: "-e node=node3"
  matchString: "-e \"skip_confirmation=true\" /kubespray/remove-node.yml -e node=node3"
  output: true

- message: "Check postback kubeconf when the cluster is installed"
  input:
    actionType: playbook
    action: kubeconfig.yml
    isPrivateKey: true
  matchString: "--private-key /auth/ssh-privatekey /kubespray/kubeconfig.yml"
  output: true

- message: "Check clean up kubeconf after cluster reset"
  input:
    actionType: playbook
    action: kubeconfig.yml
    extraArgs: "-e undo=true"
  matchString: "/kubespray/kubeconfig.yml -e undo=true"
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
				err = ep.PreHookRunPart(prehook.ActionType, prehook.Action, prehook.ExtraArgs, item.Input.IsPrivateKey, true)
				if err != nil {
					t.Fatalf("error: %v", err)
				}
			}
			// Kubespray 命令处理
			err = ep.SprayRunPart(item.Input.ActionType, item.Input.Action, item.Input.ExtraArgs, item.Input.IsPrivateKey, true)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			// Posthook 命令处理
			for _, posthook := range item.Input.PostHooks {
				err = ep.PostHookRunPart(posthook.ActionType, posthook.Action, posthook.ExtraArgs, item.Input.IsPrivateKey, true)
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

func Test_SprayRunPart(t *testing.T) {
	tests := []struct {
		name string
		args func() bool
		want bool
	}{
		{
			name: "right builtin action",
			args: func() bool {
				ep := NewEntryPoint()
				return ep.SprayRunPart(PBAction, ResetPB, "-vvv", false, true) == nil
			},
			want: true,
		},
		{
			name: "right builtin action",
			args: func() bool {
				ep := NewEntryPoint()
				return ep.SprayRunPart(PBAction, ResetPB, "-vvv", false, true) == nil
			},
			want: true,
		},
		{
			name: "wrong builtin action",
			args: func() bool {
				ep := NewEntryPoint()
				return ep.SprayRunPart(PBAction, "abc.yml", "-vvv", false, true) == nil
			},
			want: false,
		},
		{
			name: "shell action",
			args: func() bool {
				ep := NewEntryPoint()
				return ep.SprayRunPart(SHAction, "sleep 10", "", false, true) == nil
			},
			want: true,
		},
		{
			name: "other wrong action",
			args: func() bool {
				ep := NewEntryPoint()
				return ep.SprayRunPart("OtherAction", "sleep 10", "", false, true) == nil
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.args() != test.want {
				t.Fatal()
			}
		})
	}
}

func TestPBActionValue(t *testing.T) {
	if clusteroperationv1alpha1.PlaybookActionType != PBAction {
		t.Fatal()
	}
	if clusteroperationv1alpha1.ShellActionType != SHAction {
		t.Fatal()
	}
}

func TestEntryPoint_buildPlaybookCmd(t *testing.T) {
	type fields struct {
		PreHookCMDs  []string
		SprayCMD     string
		PostHookCMDs []string
		Actions      *Actions
	}
	type args struct {
		action       string
		extraArgs    string
		isPrivateKey bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test action not found in playbook.dict",
			fields: fields{
				Actions: &Actions{
					Playbooks: &Playbooks{
						Dict: map[string]void{},
					},
				},
			},
			args: args{
				action: ResetPB,
			},
			wantErr: true,
		},
		{
			name:    "test private key case",
			wantErr: false,
			fields: fields{
				Actions: &Actions{
					Playbooks: &Playbooks{
						Dict: map[string]void{
							ResetPB: {},
						},
					},
				},
			},
			args: args{
				action:       ResetPB,
				isPrivateKey: true,
			},
			want: "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\" --private-key /auth/ssh-privatekey -e \"reset_confirmation=yes\" /kubespray/reset.yml",
		},
		{
			name:    "test extra args case",
			wantErr: false,
			fields: fields{
				Actions: &Actions{
					Playbooks: &Playbooks{
						Dict: map[string]void{
							ResetPB: {},
						},
					},
				},
			},
			args: args{
				isPrivateKey: true,
				action:       ResetPB,
				extraArgs:    "-e \"reset_confirmation=yes\"",
			},
			want: "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\" --private-key /auth/ssh-privatekey -e \"reset_confirmation=yes\" /kubespray/reset.yml -e \"reset_confirmation=yes\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &EntryPoint{
				PreHookCMDs:  tt.fields.PreHookCMDs,
				SprayCMD:     tt.fields.SprayCMD,
				PostHookCMDs: tt.fields.PostHookCMDs,
				Actions:      tt.fields.Actions,
			}
			got, err := ep.buildPlaybookCmd(tt.args.action, tt.args.extraArgs, tt.args.isPrivateKey, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildPlaybookCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildPlaybookCmd() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_entryPoint_hookRunPart(t *testing.T) {
	type fields struct {
		PreHookCMDs  []string
		SprayCMD     string
		PostHookCMDs []string
		Actions      *Actions
	}
	type args struct {
		actionType   string
		action       string
		extraArgs    string
		isPrivateKey bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test build playbook cmd failed",
			fields: fields{
				Actions: &Actions{
					Playbooks: &Playbooks{
						Dict: map[string]void{},
					},
				},
			},
			args: args{
				actionType: PBAction,
				action:     ResetPB,
			},
			wantErr: true,
		},
		{
			name: "test build playbook cmd success",
			fields: fields{
				Actions: &Actions{
					Playbooks: &Playbooks{
						Dict: map[string]void{
							ResetPB: {},
						},
					},
				},
			},
			args: args{
				actionType:   PBAction,
				action:       ResetPB,
				isPrivateKey: true,
			},
			wantErr: false,
			want:    "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\" --private-key /auth/ssh-privatekey -e \"reset_confirmation=yes\" /kubespray/reset.yml",
		},
		{
			name: "test shell action case",
			args: args{
				actionType: SHAction,
				action:     ResetPB,
			},
			want:    "reset.yml",
			wantErr: false,
		},
		{
			name: "test unsupported action",
			fields: fields{
				Actions: &Actions{},
			},
			args: args{
				actionType: "unknown",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &EntryPoint{
				PreHookCMDs:  tt.fields.PreHookCMDs,
				SprayCMD:     tt.fields.SprayCMD,
				PostHookCMDs: tt.fields.PostHookCMDs,
				Actions:      tt.fields.Actions,
			}
			got, err := ep.hookRunPart(tt.args.actionType, tt.args.action, tt.args.extraArgs, tt.args.isPrivateKey, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("hookRunPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("hookRunPart() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntryPoint_Render(t *testing.T) {
	type fields struct {
		PreHookCMDs  []string
		SprayCMD     string
		PostHookCMDs []string
		Actions      *Actions
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "test PreHookCMDs not empty case",
			fields: fields{
				PreHookCMDs: []string{"cmd1", "cmd2", "cmd3"},
			},
			wantErr: false,
			want:    "#!/bin/bash\n\nset -o errexit\nset -o nounset\nset -o pipefail\n\n# preinstall\ncmd1\ncmd2\ncmd3\n\n\n# run kubespray\n\n\n# postinstall\n\n",
		},
		{
			name: "test SprayCMD not empty case",
			fields: fields{
				PreHookCMDs: []string{"cmd1", "cmd2", "cmd3"},
				SprayCMD:    "echo $TEST",
			},
			want:    "#!/bin/bash\n\nset -o errexit\nset -o nounset\nset -o pipefail\n\n# preinstall\ncmd1\ncmd2\ncmd3\n\n\n# run kubespray\necho $TEST\n\n# postinstall\n\n",
			wantErr: false,
		},
		{
			name: "test PostHookCMDs not empty case",
			fields: fields{
				PreHookCMDs:  []string{"cmd1", "cmd2", "cmd3"},
				SprayCMD:     "echo $TEST",
				PostHookCMDs: []string{"cmd4"},
			},
			wantErr: false,
			want:    "#!/bin/bash\n\nset -o errexit\nset -o nounset\nset -o pipefail\n\n# preinstall\ncmd1\ncmd2\ncmd3\n\n\n# run kubespray\necho $TEST\n\n# postinstall\ncmd4\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := &EntryPoint{
				PreHookCMDs:  tt.fields.PreHookCMDs,
				SprayCMD:     tt.fields.SprayCMD,
				PostHookCMDs: tt.fields.PostHookCMDs,
				Actions:      tt.fields.Actions,
			}
			got, err := ep.Render()
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() got = %v, want %v", got, tt.want)
			}
		})
	}
}
