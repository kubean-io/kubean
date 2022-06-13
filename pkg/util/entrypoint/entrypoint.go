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
)

const EntrypointTemplate = `
 #!/bin/bash

# preinstall
{{ range $preCMD := .PreHookCMDs }}
{{- $preCMD -}}
{{ end }}

# run kubespray
{{ .SprayCMD }}

# postinstall
{{ range $postCMD := .PostHookCMDs }}
{{- $postCMD -}}
{{ end }}

`

type void struct{}

var member void

type EntryPoint struct {
	PreHookCMDs  []string
	SprayCMD     string
	PostHookCMDs []string
	Playbooks    map[string]void
}

func NewEntryPoint() *EntryPoint {
	ep := &EntryPoint{}
	ep.Playbooks = map[string]void{
		ResetPB:          member,
		ScalePB:          member,
		ClusterPB:        member,
		RemoveNodePB:     member,
		UpgradeClusterPB: member,
	}
	return ep
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
	ep.PreHookCMDs = append(ep.PreHookCMDs, prehook)
	return nil
}

func (ep *EntryPoint) PostHookRunPart(actionType, action string) error {
	posthook, err := ep.hookRunPart(actionType, action)
	if err != nil {
		return fmt.Errorf("posthook: %w", err)
	}
	ep.PostHookCMDs = append(ep.PostHookCMDs, posthook)
	return nil
}

func (ep *EntryPoint) SprayRunPart(actionType, action, extraArgs string, isPrivateKey bool) error {
	if _, ok := ep.Playbooks[action]; !ok {
		return fmt.Errorf("unknown kubespray playbook: %s", action)
	}

	if actionType == PBAction {
		ep.SprayCMD = "ansible-playbook -i /conf/hosts.yml -b --become-user root -e \"@/conf/group_vars.yml\""
		if isPrivateKey {
			ep.SprayCMD = fmt.Sprintf("%s --private-key /auth/ssh-privatekey", ep.SprayCMD)
		}
		if action == ResetPB {
			ep.SprayCMD = fmt.Sprintf("%s -e \"reset_confirmation=yes\"", ep.SprayCMD)
		}
		if action == RemoveNodePB {
			ep.SprayCMD = fmt.Sprintf("%s -e \"skip_confirmation=true\"", ep.SprayCMD)
		}
		ep.SprayCMD = fmt.Sprintf("%s /kubespray/%s", ep.SprayCMD, action)
		if len(extraArgs) > 0 {
			ep.SprayCMD = fmt.Sprintf("%s %s", ep.SprayCMD, extraArgs)
		}
	} else if actionType == SHAction {
		ep.SprayCMD = action
	} else {
		return fmt.Errorf("unknown action type: %s", actionType)
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
