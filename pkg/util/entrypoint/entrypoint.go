package entrypoint

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	klog "k8s.io/klog/v2"
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
	KubeconfigPB  = "kubeconfig.yml"
	ClusterInfoPB = "cluster-info.yml"
	UpdateHostsPB = "update-hosts.yml"
	RemovePkgsPB  = "remove-pkgs.yml"
	PreCheckPB    = "precheck.yml"
	RenewCertsPB  = "renew-certs.yml"
)

//go:embed entrypoint.sh.template
var entrypointTemplate string

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
		PingPB, RepoPB, FirewallPB, KubeconfigPB, ClusterInfoPB, UpdateHostsPB,
		RemovePkgsPB, PreCheckPB, RenewCertsPB,
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

func (ep *EntryPoint) buildPlaybookCmd(action, extraArgs string, isPrivateKey, builtinAction bool) (string, error) {
	if builtinAction {
		if _, ok := ep.Actions.Playbooks.Dict[action]; !ok {
			return "", ArgsError{fmt.Sprintf("unknown playbook type, the currently supported ranges include: %s", ep.Actions.Playbooks.List)}
		}
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

func (ep *EntryPoint) hookRunPart(actionType, action, extraArgs string, isPrivateKey, builtinAction bool) (string, error) {
	if !builtinAction {
		klog.Infof("use external action %s, type %s", action, actionType)
	}
	hookRunCmd := ""
	if actionType == PBAction {
		playbookCmd, err := ep.buildPlaybookCmd(action, extraArgs, isPrivateKey, builtinAction)
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

func (ep *EntryPoint) PreHookRunPart(actionType, action, extraArgs string, isPrivateKey, builtinAction bool) error {
	prehook, err := ep.hookRunPart(actionType, action, extraArgs, isPrivateKey, builtinAction)
	if err != nil {
		return ArgsError{fmt.Sprintf("prehook: %s", err)}
	}
	ep.PreHookCMDs = append(ep.PreHookCMDs, prehook)
	return nil
}

func (ep *EntryPoint) PostHookRunPart(actionType, action, extraArgs string, isPrivateKey, builtinAction bool) error {
	posthook, err := ep.hookRunPart(actionType, action, extraArgs, isPrivateKey, builtinAction)
	if err != nil {
		return ArgsError{fmt.Sprintf("posthook: %s", err)}
	}
	ep.PostHookCMDs = append(ep.PostHookCMDs, posthook)
	return nil
}

func (ep *EntryPoint) SprayRunPart(actionType, action, extraArgs string, isPrivateKey, builtinAction bool) error {
	if !builtinAction {
		klog.Infof("use external action %s, type %s", action, actionType)
	}
	if actionType == PBAction {
		playbookCmd, err := ep.buildPlaybookCmd(action, extraArgs, isPrivateKey, builtinAction)
		if err != nil {
			return ArgsError{fmt.Sprintf("buildPlaybookCmd: %s", err)}
		}
		ep.SprayCMD = playbookCmd
	} else if actionType == SHAction {
		ep.SprayCMD = action
	} else {
		return ArgsError{fmt.Sprintf("unknown action type, the currently supported ranges include: %s", ep.Actions.Types)}
	}
	return nil
}

func (ep *EntryPoint) Render() (string, error) {
	b := &strings.Builder{}
	tmpl := template.Must(template.New("entrypoint").Parse(entrypointTemplate))
	if err := tmpl.Execute(b, ep); err != nil {
		return "", err
	}
	return b.String(), nil
}
