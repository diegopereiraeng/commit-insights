package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/vanng822/go-premailer/premailer"
)

const htmlHeader = `
<html>
<head>
    <style>
`

const htmlStyle = `
		body {
			font-family: Arial, sans-serif;
			margin: 0;
			background-color: #f0f0f0;
			display: flex;
			flex-direction: column;
			align-items: center;
		}
		.header {
			background-color: #00ABE3;
			color: white;
			padding: 20px;
			text-align: center;
			font-size: 20px;
			border-radius: 10px 10px 0 0;
			margin: -20px -20px 20px -20px;
		}
		.super-container {
			background-color: #fff;
			border-radius: 10px;
			box-shadow: 0 0 10px rgba(0,0,0,0.1);
			overflow-x: auto;
			overflow-y: auto;
			max-width: 90%;
			max-height: 90vh;
			margin: 10px 20px;
			padding: 20px;
			height: fit-content;
		}
		.section {
			background-color: #fff;
			border-radius: 10px;
			box-shadow: 0 0 10px rgba(0,0,0,0.1);
			overflow-x: auto;
			overflow-y: auto;
			max-width: 90%;
			max-height: 100vh;
			margin: 10px 20px;
			padding: 20px;
		}
		.green {
			background-color: rgba(40, 167, 69, 0.3);
		}
		.red {
			background-color: rgba(203, 36, 49, 0.3);
		}
		.orange {
			background-color: rgba(255, 165, 0, 0.3);
		}

		table {
			width: 100%;
			border-collapse: collapse;
			margin-bottom: 20px;
			border: 1px solid #ccc;
			
		}
		th, td {
			border: 1px solid #ccc;
			padding: 8px;
			text-align: left;
		}
		th {
			background-color: #f8f8f8;
		}

`

const htmlPreBody = `
	</style>
</head>
<body>
`
const htmlBody = `
<div class="super-container">
	<div class="header">
		Commit Insights Report
	</div>
	<div class="section">
		<strong>Repository Name:</strong> {{.RepoName}}<br>
		<strong>Branch Name:</strong> {{.BranchName}}<br>
		<strong>Trigger Type:</strong> {{.TriggerType}}
	</div>
	<div class="section">
		<strong>Committers:</strong> {{.Committers}}
	</div>
	<div class="section">
		<strong>Pipeline Name:</strong> {{.PipeName}}<br>
		<strong>Pipeline Build Started:</strong> {{.PipeBuildCreated}}<br>
		<strong>Pipeline URL:</strong> <a href={{.PipeURL}}>Harness Execution Link</a>
	</div>
	<div class="section">
		<strong>File Changes:</strong><p>
		<table>
			<tr>
				<th>Committer/Reviewer</th>
				<th>Status</th>
				<th>File Name</th>
				<th>Commit Hash</th>
				<th>Title</th>
				<th>Date</th>
			</tr>
			{{range .FileChanges}}
			<tr class="{{.StatusClass}}">
				<td>{{.Committer}}{{if .Reviewer}} / {{.Reviewer}}{{end}}</td>
				<td>{{.Status}}</td>
				<td>{{.FileName}}</td>
				<td>{{.CommitHash}}</td>
				<td>{{.Title}}</td>
				<td>{{.Time}}</td>
			</tr>
			{{end}}
		</table>
	</div>
</div>
`
const htmlPostBody = `
</body>
</html>
`

const htmlTemplate = htmlHeader + htmlStyle + htmlPreBody + htmlBody + htmlPostBody

type reportData struct {
	RepoName         string
	BranchName       string
	TriggerType      string
	Committers       string
	CommittersEmail  string
	PipeName         string
	PipeURL          string
	PipeBuildCreated string
	FileChanges      []struct {
		FileName    string
		Status      string
		StatusClass string
		Committer   string
		Reviewer    string
		CommitHash  string
		Title       string
		Time        string
	}
}

func GenerateReport(repoName string, branchName string, triggerType string, committers []string, commitersEmail []string, pipeName string, pipeURL string, fileChanges []struct {
	FileName   template.HTML
	Status     string
	Committer  string
	Reviewer   string
	CommitHash string
	Title      string
	Time       string
}, buildCreated string) (string, error) {
	var committersStr string
	if len(committers) > 0 {
		committersStr = strings.Join(committers, ", ")
	}
	var committersEmailStr string
	if len(commitersEmail) > 0 {
		committersEmailStr = strings.Join(commitersEmail, ", ")
	}

	var fileChangesData []struct {
		FileName    template.HTML
		Status      string
		StatusClass string
		Committer   string
		Reviewer    string
		CommitHash  string
		Title       string
		Time        time.Time
	}
	for _, change := range fileChanges {
		var statusText, statusClass string
		switch change.Status {
		case "A":
			statusText = "Added"
			statusClass = "green"
		case "M":
			statusText = "Modified"
			statusClass = "orange"
		case "D":
			statusText = "Deleted"
			statusClass = "red"
		default:
			statusText = change.Status
		}
		//convert epoch 1694299436 format to human readable format
		// timeInt, err := strconv.Atoi(change.Time)
		// if err != nil {
		// 	return "", err
		// }
		// currentTimezone := time.Now().Format("-0700")
		// // fmt.Println(currentTimezone)

		// date := time.Unix(int64(timeInt), 0).Format("2006-01-02 15:04:05") + " " + currentTimezone
		// _ = date // fix "declared and not used" error
		timeInt, err := strconv.Atoi(change.Time)
		if err != nil {
			return "", err
		}
		date := time.Unix(int64(timeInt), 0)
		_, offsetSeconds := time.Now().Zone()
		offset := time.Duration(offsetSeconds) * time.Second
		date = date.Add(offset)

		fileChangesData = append(fileChangesData, struct {
			FileName    template.HTML
			Status      string
			StatusClass string
			Committer   string
			Reviewer    string
			CommitHash  string
			Title       string
			Time        time.Time
		}{
			FileName:    change.FileName,
			Status:      statusText,
			StatusClass: statusClass,
			Committer:   change.Committer,
			Reviewer:    change.Reviewer,
			CommitHash:  change.CommitHash,
			Title:       change.Title,
			Time:        date,
		})
	}

	sort.Slice(fileChangesData, func(i, j int) bool {
		return fileChangesData[i].Time.After(fileChangesData[j].Time)
	})

	data := reportData{
		RepoName:         repoName,
		BranchName:       branchName,
		TriggerType:      triggerType,
		Committers:       committersStr,
		CommittersEmail:  committersEmailStr,
		PipeName:         pipeName,
		PipeURL:          pipeURL,
		PipeBuildCreated: buildCreated,
		FileChanges: func() []struct {
			FileName    string
			Status      string
			StatusClass string
			Committer   string
			Reviewer    string
			CommitHash  string
			Title       string
			Time        string
		} {
			var changes []struct {
				FileName    string
				Status      string
				StatusClass string
				Committer   string
				Reviewer    string
				CommitHash  string
				Title       string
				Time        string
			}
			for _, change := range fileChangesData {
				changes = append(changes, struct {
					FileName    string
					Status      string
					StatusClass string
					Committer   string
					Reviewer    string
					CommitHash  string
					Title       string
					Time        string
				}{
					FileName:    string(change.FileName),
					Status:      change.Status,
					StatusClass: change.StatusClass,
					Committer:   change.Committer,
					Reviewer:    change.Reviewer,
					CommitHash:  change.CommitHash,    // Added this line
					Title:       change.Title,         // Added this line
					Time:        change.Time.String(), // Updated this line
				})
			}
			return changes
		}(),
	}

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var report strings.Builder
	if err := tmpl.Execute(&report, data); err != nil {
		return "", err
	}

	p, err := premailer.NewPremailerFromString(report.String(), premailer.NewOptions())
	if err != nil {
		return "", err
	}

	inlinedHtml, err := p.Transform()
	if err != nil {
		fmt.Println("Error inlining CSS:", err)
		return "", err
	}
	// inlinedHtml = strings.ReplaceAll(inlinedHtml, "\n", "")
	fmt.Println(inlinedHtml)

	vars := map[string]string{
		"HTML_TEMPLATE":      htmlTemplate,
		"HTML_HEADER":        htmlHeader,
		"HTML_STYLE":         htmlStyle,
		"HTML_PREBODY":       htmlPreBody,
		"HTML_BODY":          htmlBody,
		"HTML_POSTBODY":      htmlPostBody,
		"REPO_NAME":          repoName,
		"BRANCH_NAME":        branchName,
		"TRIGGER_TYPE":       triggerType,
		"COMMITTERS":         committersStr,
		"COMMITTERS_EMAIL":   committersEmailStr,
		"PIPE_NAME":          pipeName,
		"PIPE_URL":           pipeURL,
		"PIPE_BUILD_CREATED": buildCreated,
		"FILE_CHANGES":       fmt.Sprintf("%v", fileChanges),
		"REPORT":             strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(report, "\n", ""), "	", ""), "		", ""), "<!DOCTYPE html>", ""),
	}

	err = writeEnvFile(vars, os.Getenv("DRONE_OUTPUT"))

	if err != nil {
		fmt.Printf("| \033[33m[WARNING] - Failed to write to .env: %v\033[0m\n", err)
	}

	return inlinedHtml, nil
}

func writeEnvFile(vars map[string]string, outputPath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Println("Creating directory:", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Println("Error creating directory:", err)
			return err
		}
	}

	// Create the file if it doesn't exist
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fmt.Println("| Creating env file for Harness:", outputPath)
		if _, err := os.Create(outputPath); err != nil {
			fmt.Println("| \033[33m[WARNING] - Error creating file: ", err, "\033[0m")
			return err
		}
	}

	// Use godotenv.Write() to write the vars map to the specified file
	err := godotenv.Write(vars, outputPath)
	if err != nil {
		fmt.Println("[WARNING] \033[33m| Error writing to .env file: ", err, "\033[0m")
		return err
	}
	fmt.Println("Successfully wrote to .env file")

	// Read the file contents
	content, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Println("Error reading the .env file:", err)
		return err
	}

	// Print the file contents
	fmt.Println("File contents:")
	fmt.Println(string(content))

	return nil
}
