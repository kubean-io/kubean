package entrypoint

import (
	"fmt"
	"strings"
	"text/template"
)

// Generate kubespray job entrypoint script

const (
	PBAction = "playbook"
	SHAction = "shell"

	ResetPB          = "reset.yml"
	ScalePB          = "scale.yml"
	ClusterPB        = "cluster.yml"
	RemoveNodePB     = "remove-node.yml"
	UpgradeClusterPB = "upgrade-cluster.yml"

	PingPB        = "ping.yml"
	RepoPB        = "enable-repo.yml"
	FirewallPB    = "disable-firewalld.yml"
	ClusterInfoPB = "cluster-info.yml"

	KubeconfInstall = `
set -x
# postback cluster kubeconfig
inventory_file="/conf/hosts.yml"
first_master=` + "`" + `yq e '.all.children.kube_control_plane.hosts' $inventory_file -o y | head -n 1 | cut -f1 -d":"` + "`" + `
first_master_ip=` + "`" + `yq e '.all.hosts.'$first_master'.ip' $inventory_file -o y` + "`" + `
kubeconfig_name="$CLUSTER_NAME-kubeconf"
fetch_src="/root/.kube/config"
fetch_dest="/root/$kubeconfig_name"
ansible -i $inventory_file $first_master -m fetch -a "src=$fetch_src dest=$fetch_dest" %s
sed -i "s/127.0.0.1:.*/"$first_master_ip":6443/" $fetch_dest/$first_master$fetch_src
set +e
kubeconf_count=` + "`" + `kubectl -n kubean-system get configmap | grep $kubeconfig_name | wc -l | sed 's/ //g'` + "`" + `
if [ "${kubeconf_count}" -gt 0 ]; then
    kubectl -n kubean-system delete configmap $kubeconfig_name
fi
set -e
kubectl -n kubean-system create configmap $kubeconfig_name --from-file=$fetch_dest/$first_master$fetch_src
kubectl patch --type=merge kubeancluster $CLUSTER_NAME -p '{"spec": {"kubeconfRef": {"name": "'$kubeconfig_name'", "namespace": "kubean-system"}}}'
`
	KubeconfReset = `
set -x
# reset cluster kubeconfig
kubeconfig_name="$CLUSTER_NAME-kubeconf"
set +e
kubeconf_count=` + "`" + `kubectl -n kubean-system get configmap | grep $kubeconfig_name | wc -l | sed 's/ //g'` + "`" + `
if [ "${kubeconf_count}" -gt 0 ]; then
    kubectl -n kubean-system delete configmap $kubeconfig_name
fi
set -e
kubectl patch --type=merge kubeancluster $CLUSTER_NAME -p '{"spec": {"kubeconfRef": null}}'
`
	EntrypointTemplate = `
#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# preinstall
{{ range $preCMD := .PreHookCMDs }}
{{- $preCMD }}
{{ end }}

# run kubespray
{{ .SprayCMD }}

# postinstall
{{ range $postCMD := .PostHookCMDs }}
{{- $postCMD }}
{{ end }}
`
)

type void struct{}

var member void

type Playbooks struct {
	List []string
	Dict map[string]void
}

type Actions struct {
	Types     []string
	Playbooks *Playbooks
}

func NewActions() *Actions {
	actions := &Actions{}
	actions.Types = []string{PBAction, SHAction}
	actions.Playbooks = &Playbooks{}
	actions.Playbooks.List = []string{
		ResetPB, ScalePB, ClusterPB, RemoveNodePB, UpgradeClusterPB,
		PingPB, RepoPB, FirewallPB, ClusterInfoPB,
	}
	actions.Playbooks.Dict = map[string]void{}
	for _, pbItem := range actions.Playbooks.List {
		actions.Playbooks.Dict[pbItem] = member
	}
	return actions
}

type ArgsError struct {
	msg string
}

func (argsError ArgsError) Error() string {
	return argsError.msg
}

type EntryPoint struct {
	PreHookCMDs  []string
	SprayCMD     string
	PostHookCMDs []string
	Actions      *Actions
}

func NewEntryPoint() *EntryPoint {
	ep := &EntryPoint{}
	ep.Actions = NewActions()
	return ep
}

func (ep *EntryPoint) buildPlaybookCmd(action, extraArgs string, isPrivateKey bool) (string, error) {
	if _, ok := ep.Actions.Playbooks.Dict[action]; !ok {
		return "", ArgsError{fmt.Sprintf("unknown playbook type, the currently supported ranges include: %s", ep.Actions.Playbooks.List)}
	}
	playbookCmd := "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\""
	if isPrivateKey {
		playbookCmd = fmt.Sprintf("%s --private-key /auth/ssh-privatekey", playbookCmd)
	}
	if action == ResetPB {
		playbookCmd = fmt.Sprintf("%s -e \"reset_confirmation=yes\"", playbookCmd)
	}
	if action == RemoveNodePB {
		playbookCmd = fmt.Sprintf("%s -e \"skip_confirmation=true\"", playbookCmd)
	}
	playbookCmd = fmt.Sprintf("%s /kubespray/%s", playbookCmd, action)
	if len(extraArgs) > 0 {
		playbookCmd = fmt.Sprintf("%s %s", playbookCmd, extraArgs)
	}
	return playbookCmd, nil
}

func (ep *EntryPoint) hookRunPart(actionType, action, extraArgs string, isPrivateKey bool) (string, error) {
	hookRunCmd := ""
	if actionType == PBAction {
		playbookCmd, err := ep.buildPlaybookCmd(action, extraArgs, isPrivateKey)
		if err != nil {
			return "", ArgsError{fmt.Sprintf("buildPlaybookCmd: %s", err)}
		}
		hookRunCmd = playbookCmd
	} else if actionType == SHAction {
		hookRunCmd = action
	} else {
		return "", ArgsError{fmt.Sprintf("unknown action type, the currently supported ranges include: %s", ep.Actions.Types)}
	}
	return hookRunCmd, nil
}

func (ep *EntryPoint) PreHookRunPart(actionType, action, extraArgs string, isPrivateKey bool) error {
	prehook, err := ep.hookRunPart(actionType, action, extraArgs, isPrivateKey)
	if err != nil {
		return ArgsError{fmt.Sprintf("prehook: %s", err)}
	}
	ep.PreHookCMDs = append(ep.PreHookCMDs, prehook)
	return nil
}

func (ep *EntryPoint) PostHookRunPart(actionType, action, extraArgs string, isPrivateKey bool) error {
	posthook, err := ep.hookRunPart(actionType, action, extraArgs, isPrivateKey)
	if err != nil {
		return ArgsError{fmt.Sprintf("posthook: %s", err)}
	}
	ep.PostHookCMDs = append(ep.PostHookCMDs, posthook)
	return nil
}

func (ep *EntryPoint) kubeconfPostbackPart(action string, isPrivateKey bool) error {
	if _, ok := ep.Actions.Playbooks.Dict[action]; !ok {
		return nil
	}
	script := ""
	if action == ClusterPB {
		script = KubeconfInstall
		if isPrivateKey {
			script = fmt.Sprintf(script, "--private-key /auth/ssh-privatekey")
		} else {
			script = fmt.Sprintf(script, "")
		}
	}
	if action == ResetPB {
		script = KubeconfReset
	}
	posthook, err := ep.hookRunPart(SHAction, script, "", false)
	if err != nil {
		return ArgsError{fmt.Sprintf("posthook: %s", err)}
	}
	ep.PostHookCMDs = append(ep.PostHookCMDs, posthook)
	return nil
}

func (ep *EntryPoint) SprayRunPart(actionType, action, extraArgs string, isPrivateKey bool) error {
	if actionType == PBAction {
		playbookCmd, err := ep.buildPlaybookCmd(action, extraArgs, isPrivateKey)
		if err != nil {
			return ArgsError{fmt.Sprintf("buildPlaybookCmd: %s", err)}
		}
		ep.SprayCMD = playbookCmd
	} else if actionType == SHAction {
		ep.SprayCMD = action
	} else {
		return ArgsError{fmt.Sprintf("unknown action type, the currently supported ranges include: %s", ep.Actions.Types)}
	}
	if err := ep.kubeconfPostbackPart(action, isPrivateKey); err != nil {
		return ArgsError{fmt.Sprintf("failed to set kubeconfig postback: %s", err)}
	}
	return nil
}

func (ep *EntryPoint) Render() (string, error) {
	b := &strings.Builder{}
	tmpl := template.Must(template.New("entrypoint").Parse(EntrypointTemplate))
	if err := tmpl.Execute(b, ep); err != nil {
		return "", err
	}
	return b.String(), nil
}
