package entrypoint

import (
	"fmt"
	"strings"
	"text/template"
)

const (
	PBAction = "playbook"
	SHAction = "shell"

	ResetPB          = "reset.yml"
	ScalePB          = "scale.yml"
	ClusterPB        = "cluster.yml"
	RemoveNodePB     = "remove-node.yml"
	UpgradeClusterPB = "upgrade-cluster.yml"
)

// Generate kubespray job entrypoint script

type EntryPoint struct {
	PreHookRun  string
	SprayRun    string
	PostHookRun string
}

func (ep *EntryPoint) hookRunPart(actionType, action string) (string, error) {
	hookRunCmd := ""
	if actionType == PBAction {
		// todo
		return "", fmt.Errorf("playbook is not currently supported")
	} else if actionType == SHAction {
		hookRunCmd = action
	} else {
		return "", fmt.Errorf("unknown action type: %s", actionType)
	}
	return hookRunCmd, nil
}

func (ep *EntryPoint) PreHookRunPart(actionType, action string) error {
	prehook, err := ep.hookRunPart(actionType, action)
	if err != nil {
		return fmt.Errorf("prehook: %w", err)
	}
	ep.PreHookRun = prehook
	return nil
}

func (ep *EntryPoint) PostHookRunPart(actionType, action string) error {
	posthook, err := ep.hookRunPart(actionType, action)
	if err != nil {
		return fmt.Errorf("posthook: %w", err)
	}
	ep.PostHookRun = posthook
	return nil
}

func (ep *EntryPoint) SprayRunPart(actionType, action string, isPrivateKey bool) error {
	sprayRunCmd := ""
	if actionType == PBAction {
		sprayRunCmd = "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\""
		if isPrivateKey {
			sprayRunCmd = fmt.Sprintf("%s --private-key /auth/ssh-privatekey", sprayRunCmd)
		}
		if action == ClusterPB {
			sprayRunCmd = fmt.Sprintf("%s /kubespray/cluster.yml", sprayRunCmd)
		} else if action == ResetPB {
			sprayRunCmd = fmt.Sprintf("%s -e \"reset_confirmation=yes\" /kubespray/reset.yml", sprayRunCmd)
		} else {
			return fmt.Errorf("unknown kubespray playbook: %s", action)
		}
	} else if actionType == SHAction {
		sprayRunCmd = action
	} else {
		return fmt.Errorf("unknown action type: %s", actionType)
	}
	ep.SprayRun = sprayRunCmd
	return nil
}

func (ep *EntryPoint) Render() (string, error) {
	b := &strings.Builder{}
	tmpl := template.Must(template.ParseFiles("entrypoint.sh.template"))
	if err := tmpl.Execute(b, ep); err != nil {
		return "", err
	}
	return b.String(), nil
}
