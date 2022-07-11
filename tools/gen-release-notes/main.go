package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	blackfriday "github.com/russross/blackfriday/v2"
	"sigs.k8s.io/yaml"
)

// golang flags don't accept arrays by default. This adds it.
type flagStrings []string

func (flagString *flagStrings) String() string {
	return strings.Join(*flagString, ",")
}

func (flagString *flagStrings) Set(value string) error {
	*flagString = append(*flagString, value)
	return nil
}

const kubeanIssueAddress = "https://gitlab.daocloud.cn/ndx/engineering/kubean/-/issues/"

func main() {
	var templatesDir, outDir, oldRelease, newRelease string
	var validateOnly bool
	var notesDirs flagStrings
	var issueAddress string

	flag.Var(&notesDirs, "notes", "the directory containing release notes. Repeat for multiple notes directories")
	flag.StringVar(&templatesDir, "templates", "./templates", "the directory containing release note templates")
	flag.StringVar(&outDir, "outDir", ".", "the directory containing release notes")
	flag.BoolVar(&validateOnly, "validate-only", false, "only for validation")
	flag.StringVar(&oldRelease, "oldRelease", "x.y.(z-1)", "old release")
	flag.StringVar(&newRelease, "newRelease", "x.y.z", "new release")
	flag.StringVar(&issueAddress, "issueAddress", "", "specify the issue link address")
	flag.Parse()

	if len(notesDirs) == 0 {
		notesDirs = []string{"."}
	}

	if issueAddress == "" {
		issueAddress = kubeanIssueAddress
	}

	var releaseNotes []Note
	for _, notesDir := range notesDirs {
		var releaseNoteFiles []string

		fmt.Printf("Looking for release notes in %s.\n", notesDir)

		releaseNotesDir := "release-notes/notes"
		if _, err := os.Stat(notesDir); os.IsNotExist(err) {
			fmt.Printf("Could not find repository -- directory %s does not exist.\n", notesDir)
			os.Exit(1)
		}

		if _, err := os.Stat(filepath.Join(notesDir, releaseNotesDir)); os.IsNotExist(err) {
			fmt.Printf("could not find release notes directory -- %s does not exist\n", filepath.Join(notesDir, releaseNotesDir))
			os.Exit(2)
		}

		var err error
		releaseNoteFiles, err = getNewFilesWithDiff(oldRelease, notesDir, releaseNotesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Found %d files.\n\n", len(releaseNoteFiles))

		fmt.Printf("Parsing release notes\n")
		releaseNotesEntries, err := parseReleaseNotesFiles(notesDir, releaseNoteFiles, issueAddress)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read release notes: %s\n", err.Error())
			os.Exit(1)
		}
		releaseNotes = append(releaseNotes, releaseNotesEntries...)
	}

	if len(releaseNotes) < 1 {
		fmt.Fprintf(os.Stderr, "failed to find any release notes.\n")
		// maps to EX_NOINPUT, but more importantly lets us differentiate between no files found and other errors
		os.Exit(66)
	}

	if validateOnly {
		return
	}

	fmt.Printf("\nLooking for markdown templates in %s.\n", templatesDir)
	templateFiles, err := getFilesWithExtension(templatesDir, "md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list files: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("Found %d files.\n\n", len(templateFiles))

	for _, filename := range templateFiles {
		output, err := populateTemplate(templatesDir, filename, releaseNotes, oldRelease, newRelease)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
			os.Exit(1)
		}

		filename = newRelease + "-" + filename

		if err := createDirIfNotExists(outDir); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create our dir: %s\n", err.Error())
		}
		if err := writeAsMarkdown(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write markdown: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s\n", filename)
		}

		if err := writeAsHTML(path.Join(outDir, filename), output); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write HTML: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote markdown to %s.html\n", filename)
		}
	}
}

func createDirIfNotExists(path string) error {
	err := os.MkdirAll(path, 0o755)
	if os.IsExist(err) {
		return nil
	}
	return err
}

// writeAsHTML generates HTML from markdown before writing it to a file.
func writeAsHTML(filename, markdown string) error {
	output := string(blackfriday.Run([]byte(markdown)))
	filename = strings.Replace(filename, ".md", ".html", 1)

	if err := ioutil.WriteFile(filename, []byte(output), 0o644); err != nil {
		return err
	}
	return nil
}

// writeAsMarkdown writes markdown to a file.
func writeAsMarkdown(filename, markdown string) error {
	if err := ioutil.WriteFile(filename, []byte(markdown), 0o644); err != nil {
		return err
	}
	return nil
}

func parseTemplateFormat(releaseNotes []Note, format string) ([]string, error) {
	template, err := ParseTemplate(format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %s", err.Error())
	}
	return getNotesForTemplateFormat(releaseNotes, template), nil
}

func getNotesForTemplateFormat(notes []Note, template Template) []string {
	parsedNotes := make([]string, 0)

	for _, note := range notes {
		if template.Type == "releaseNotes" {
			parsedNotes = append(parsedNotes, note.getReleaseNotes(template.Kind, template.Area, template.Action)...)
		} else if template.Type == "upgradeNotes" {
			parsedNotes = append(parsedNotes, note.getUpgradeNotes()...)
		} else if template.Type == "securityNotes" {
			parsedNotes = append(parsedNotes, note.getSecurityNotes()...)
		}
	}
	return parsedNotes
}

// getFilesWithExtension returns the files from filePath with extension extension.
func getFilesWithExtension(filePath, extension string) ([]string, error) {
	fmt.Println(os.Getwd())

	directory, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open directory: %s", err.Error())
	}
	defer directory.Close()

	var files []string
	files, err = directory.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("unable to list files for directory %s: %s", filePath, err.Error())
	}

	filesWithExtension := make([]string, 0)
	for _, fileName := range files {
		if strings.HasSuffix(fileName, extension) {
			filesWithExtension = append(filesWithExtension, fileName)
		}
	}

	return filesWithExtension, nil
}

func parseReleaseNotesFiles(filePath string, files []string, issueAddress string) ([]Note, error) {
	notes := make([]Note, 0)
	for _, file := range files {
		file = path.Join(filePath, file)
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to open file %s: %s", file, err.Error())
		}

		var note Note
		if err = yaml.Unmarshal(contents, &note); err != nil {
			return nil, fmt.Errorf("unable to parse release note %s:%s", file, err.Error())
		}
		note.File = file
		if issueAddress != "" {
			note.IssueAddress = issueAddress
		}
		notes = append(notes, note)
		fmt.Printf("found %d upgrade notes, %d release notes, and %d security notes in %s\n", len(note.UpgradeNotes),
			len(note.ReleaseNotes), len(note.SecurityNotes), note.File)
	}
	return notes, nil
}

func populateTemplate(filepath, filename string, releaseNotes []Note, oldRelease, newRelease string) (string, error) {
	filename = path.Join(filepath, filename)
	fmt.Printf("Processing %s\n", filename)

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to open file %s: %s", filename, err.Error())
	}

	comment := regexp.MustCompile("<!--(.*)-->")
	output := string(contents)

	output = strings.Replace(output, "<!--oldRelease-->", oldRelease, -1)
	output = strings.Replace(output, "<!--newRelease-->", newRelease, -1)

	now := time.Now()
	output = strings.Replace(output, "<!--publishDate-->", now.Format("2006-01-02"), -1)

	results := comment.FindAllString(output, -1)

	for _, result := range results {
		contents, err := parseTemplateFormat(releaseNotes, result)
		if err != nil {
			return "", fmt.Errorf("unable to parse templates: %s", err.Error())
		}
		joinedContents := strings.Join(contents, "\n")
		output = strings.Replace(output, result, joinedContents, -1)
	}

	return output, nil
}

func getNewFilesWithDiff(oldRelease, path, notesSubpath string) ([]string, error) {
	cmd := fmt.Sprintf("cd %s; git -P diff -r --diff-filter=AMR --name-only --relative=%s '%s'", path, notesSubpath, oldRelease)
	fmt.Printf("Executing: %s\n", cmd)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, err
	}
	outFiles := strings.Split(string(out), "\n")

	// the getFilesFromGHPRView(path, pullRequest, notesSubpath) method returns file names which are relative to the repo path.
	// the git diff-tree is relative to the notesSupbpath, so we need to add the subpath back to the filenames.
	outFileswithPath := []string{}
	for _, f := range outFiles[:len(outFiles)-1] { // skip the last file which is empty
		outFileswithPath = append(outFileswithPath, filepath.Join(notesSubpath, f))
	}

	return outFileswithPath, nil
}
