package main

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
)

type Note struct {
	Kind          string         `json:"kind"`
	Area          string         `json:"area"`
	Issues        []string       `json:"issues,omitempty"`
	Jiras         []string       `json:"jiras,omitempty"`
	ReleaseNotes  []releaseNote  `json:"releaseNotes"`
	UpgradeNotes  []upgradeNote  `json:"upgradeNotes"`
	SecurityNotes []securityNote `json:"securityNotes"`
	File          string
	IssueAddress  string
}

func (note Note) getIssues() string {
	issueString := ""
	for _, issue := range note.Issues {
		if issueString != "" {
			issueString += " , "
		}
		if strings.Contains(issue, "gitlab.com") {
			issueNumber := path.Base(issue)
			issueString += fmt.Sprintf("([Issue #%s](%s)) ", issueNumber, issue)
		} else {
			issueString += fmt.Sprintf("( [Issue #%s](%s/%s) ) ",
				issue, strings.TrimRight(note.IssueAddress, "/"), issue)
		}
	}
	return issueString
}

func (note Note) getJiras() string {
	jiraString := ""
	for _, jira := range note.Jiras {
		if jiraString != "" {
			jiraString += ","
		}
		if strings.Contains(jira, "jira.daocloud.io/") {
			jiraString += fmt.Sprintf("([Jira #%s](%s)) ", jira, jira)
		} else {
			jiraString += fmt.Sprintf("([Jira #%s](https://jira.daocloud.io/browse/%s)) ", jira, jira)
		}
	}
	return jiraString
}

func filterNote(templateFilter, noteFilter string) bool {
	if templateFilter == "" {
		return true
	} else if templateFilter == noteFilter {
		return true
	} else if templateFilter[0] == '!' && templateFilter[1:] != noteFilter {
		return true
	}
	return false
}

func (note Note) getReleaseNotes(kind, area, action string) []string {
	notes := make([]string, 0)

	for _, releaseNote := range note.ReleaseNotes {
		if filterNote(kind, note.Kind) &&
			filterNote(area, note.Area) &&
			filterNote(action, releaseNote.Action) {
			noteEntry := fmt.Sprintf("%s %s %s\n", releaseNote, note.getIssues(), note.getJiras())
			if noteEntry != "" {
				notes = append(notes, noteEntry)
			}
		}
	}
	return notes
}

func (note Note) getSecurityNotes() []string {
	notes := make([]string, 0)
	for _, securityNote := range note.SecurityNotes {
		notes = append(notes, securityNote.String())
	}
	return notes
}

func (note Note) getUpgradeNotes() []string {
	notes := make([]string, 0)
	for _, upgradeNote := range note.UpgradeNotes {
		notes = append(notes, upgradeNote.String())
	}
	return notes
}

type upgradeNote struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (note *upgradeNote) UnmarshalJSON(data []byte) error {
	type noteIntType upgradeNote
	var noteInt noteIntType
	if err := json.Unmarshal(data, &noteInt); err != nil {
		return err
	}

	if noteInt.Title == "" {
		return fmt.Errorf("upgrade note title cannot be empty")
	}
	note.Title = noteInt.Title

	if noteInt.Content == "" {
		return fmt.Errorf("upgrade note body cannot be empty")
	}
	note.Content = noteInt.Content
	return nil
}

func (note upgradeNote) String() string {
	return fmt.Sprintf("## %s\n%s", note.Title, note.Content)
}

type releaseNote struct {
	Value  string
	Action string
	Issues []string
}

var actionList = map[string]int{
	"新增": 1,
	"弃用": 1,
	"修复": 1,
	"优化": 1,
	"改进": 1,
	"移除": 1,
	"升级": 1,
}

func (note *releaseNote) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &note.Value); err != nil {
		return err
	}
	if note.Value == "" {
		return fmt.Errorf("value missing for note: %s", note.Value)
	}

	note.Action = note.getAction(note.Value)
	if note.Action == "" {
		return fmt.Errorf("unable to determine action for note: %s; notes must start with an action and be of the form"+
			"**Action** {text} with an action listed here: "+
			"https://gitlab.daocloud.cn/ndx/mspider/-/blob/main/releasenotes/template.yaml", note.Value)
	}

	if _, ok := actionList[note.Action]; !ok {
		return fmt.Errorf("action %s is not allowed, refer to "+
			"https://gitlab.daocloud.cn/ndx/mspider/-/blob/main/releasenotes/template.yaml "+
			"for a list of allowed actions", note.Action)
	}
	return nil
}

func (note releaseNote) getAction(line string) string {
	action := ""
	actionRegexp := regexp.MustCompile(`\*\*\p{Han}*\*\*`)
	if match := actionRegexp.FindString(line); match != "" {
		action = match[2 : len(match)-2]
	}
	return action
}

func (note releaseNote) String() string {
	return fmt.Sprintf("- %s", note.Value)
}

type securityNote string

func (note securityNote) String() string {
	return fmt.Sprintf("- %s", string(note))
}
