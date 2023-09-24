package main

import (
	"io"

	"commit-insights/internal/models"

	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	Config struct {
		AccID            string   `json:"accID"`
		OrgID            string   `json:"orgID"`
		ProjectID        string   `json:"projectID"`
		PipelineID       string   `json:"pipelineID"`
		StageID          string   `json:"stageID"`
		StatusList       []string `json:"statusList"`
		RepoName         string   `json:"repoName"`
		Branch           string   `json:"branch"`
		BuildType        string   `json:"buildType"`
		IngestionType    string   `json:"ingestionType"`
		CommitID         string   `json:"commitID"`
		HarnessSecret    string   `json:"harnessSecret"`
		PipeExecutionURL string   `json:"harnessPipeExecutionURL"`
	}

	Plugin struct {
		Config Config
	}
)

var plugin Plugin

const lineBreak = "|---------------------------------------------"

func getLastSuccessfulExecution(accID string, orgID string, projectID string, pipelineID string, stageID string, statusList []string, repoName string, branch string, buildType string) (string, string, models.Pipeline, error) {

	url := "https://app.harness.io/pipeline/api/pipelines/execution/summary?page=0&size=1&accountIdentifier=" + accID + "&orgIdentifier=" + orgID + "&projectIdentifier=" + projectID + "&pipelineIdentifier=" + pipelineID + ""
	method := "POST"

	var statusListJson string
	statusListJsonBytes, err := json.Marshal(statusList)
	if err != nil {
		return "", "", models.Pipeline{}, err
	}
	statusListJson = string(statusListJsonBytes)

	if buildType == "push" {
		buildType = "branch"
	} else if buildType == "pull_request" {
		buildType = "PR"
	}
	payload := strings.NewReader(fmt.Sprintf(`{"status":%s,"moduleProperties":{"ci":{"buildType":"%s","branch":"%s","repoName":"%s"}},"filterType":"PipelineExecution"}`, statusListJson, buildType, branch, repoName))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return "", "", models.Pipeline{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", plugin.Config.HarnessSecret)
	fmt.Println(url)
	fmt.Println(payload)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error finding last successful execution")
		fmt.Println(url)
		fmt.Println(payload)
		fmt.Println(err)
		fmt.Println("Status: ", req.Response.Status)
		return "", "", models.Pipeline{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", models.Pipeline{}, err
	}
	fmt.Println(string(body))

	defer res.Body.Close()

	var response models.Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", "", models.Pipeline{}, errors.New("error parsing JSON")
	}

	if len(response.Data.Content) == 0 {
		return "", "", models.Pipeline{}, errors.New("no successful execution found")
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", "", models.Pipeline{}, err
	}
	pipelineParsed, err := parsePipeline(responseBytes)
	if err != nil {
		return "", "", models.Pipeline{}, err
	}
	fmt.Println(pipelineParsed)

	var pipeline models.Pipeline
	for _, content := range response.Data.Content {
		fmt.Printf("| Found execution with status:\033[0m \033[1;32m%s\033[0m\n", content.Status)
		fmt.Println(lineBreak)
		fmt.Printf("| \033[1;36mPlan Execution ID:\033[0m \033[1;32m%s\033[0m\n", content.PlanExecutionId)
		fmt.Printf("| \033[1;36mPipeline Name:\033[0m \033[1;32m%s\033[0m\n", content.Name)
		fmt.Printf("| \033[1;36mPipe Status:\033[0m \033[1;32m%s\033[0m\n", content.Status)
		fmt.Println(lineBreak)

		pipeline = models.Pipeline{
			Name:        content.Name,
			Status:      content.Status,
			StartedTime: time.Unix(int64(content.StartTs/1000), 0).String(),
			Duration:    time.Unix(int64(content.EndTs/1000), 0).Sub(time.Unix(int64(content.StartTs/1000), 0)).String(),
			StageCount:  0,
			StepCount:   0,
			Message:     "",
		}

		// content := response.Data.Content[0]
		// foundSuccessfulExecution := false
		for _, nodeInfo := range content.LayoutNodeMap {
			fmt.Println(lineBreak)
			fmt.Printf("| \033[1;36mNode Type:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeType)
			fmt.Printf("| \033[1;36mNode Group:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeGroup)
			fmt.Printf("| \033[1;36mNode Identifier:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeIdentifier)
			fmt.Printf("| \033[1;36mName:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Name)
			fmt.Printf("| \033[1;36mNode UUID:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.NodeUuid)
			fmt.Printf("| \033[1;36mStatus:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Status)
			fmt.Printf("| \033[1;36mModule:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.Module)
			fmt.Printf("| \033[1;36mStart TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(nodeInfo.StartTs/1000), 0))
			fmt.Printf("| \033[1;36mEnd TS:\033[0m \033[1;32m%s\033[0m\n", time.Unix(int64(nodeInfo.EndTs/1000), 0))
			fmt.Printf("| \033[1;36mFailure Info:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.FailureInfo.Message)
			fmt.Printf("| \033[1;36mEdge Layout List:\033[0m \033[1;32m%s\033[0m\n", nodeInfo.EdgeLayoutList)
			fmt.Println(lineBreak)
		}

		// fmt.Printf("| \033[1;36mStage ID:\033[0m \033[1;32m%s\033[0m\n", stageID)

		var commiters []string

		for _, commit := range content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits {
			commiters = append(commiters, commit.ID)
		}
		// fmt.Printf("Commits: %s\n", strings.Join(commiters, ", "))
		fmt.Printf("| \033[1;36mNumber of commits:\033[0m \033[1;32m%d\033[0m\n", len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits))
		fmt.Println(lineBreak)
		// fmt.Printf("Last Commit SHA: %s\n", content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits)-1].ID)

		if len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits) > 0 {
			// fmt.Printf("First Commit SHA: %s\n", content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[0].ID)
		} else {
			fmt.Println("No commits found")
		}
		firstCommit := content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[0].ID
		lastCommit := content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits[len(content.ModuleInfo.CI.CIExecutionInfoDTO.Branch.Commits)-1].ID

		return firstCommit, lastCommit, pipeline, nil
	}

	return "", "", models.Pipeline{}, errors.New("no successful execution found")
}

func (p *Plugin) Exec() error {

	plugin = *p

	// var accID string = "6_vVHzo9Qeu9fXvj-AcbC"
	// var orgID string = "default"
	// var projectID string = "GIT_FLOW_DEMO"
	// var pipelineID string = "FAST_CISTO_SonarQube_Quality_Gate_Plugin"
	// var stageID string = "Build_Golang"
	// var statusList []string = []string{"Success", "Aborted"}
	// var repoName string = "sonarqube-scanner"
	// var branch string = "main"
	// var buildType string = "branch"
	var accID string = p.Config.AccID
	var orgID string = p.Config.OrgID
	var projectID string = p.Config.ProjectID
	var pipelineID string = p.Config.PipelineID
	var stageID string = p.Config.StageID
	var statusList []string = p.Config.StatusList
	var repoName string = p.Config.RepoName
	var branch string = p.Config.Branch
	var buildType string = p.Config.BuildType
	var commitID string = p.Config.CommitID

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mHarness Commit Insights\033[0m")
	fmt.Println(lineBreak)
	fmt.Printf("| \033[1;36mAccount ID:\033[0m \033[1;32m%s\033[0m\n", accID)
	fmt.Printf("| \033[1;36mOrg ID:\033[0m \033[1;32m%s\033[0m\n", orgID)
	fmt.Printf("| \033[1;36mProject ID:\033[0m \033[1;32m%s\033[0m\n", projectID)
	fmt.Printf("| \033[1;36mPipeline ID:\033[0m \033[1;32m%s\033[0m\n", pipelineID)
	fmt.Printf("| \033[1;36mStage ID:\033[0m \033[1;32m%s\033[0m\n", stageID)
	fmt.Printf("| \033[1;36mStatus List:\033[0m \033[1;32m%s\033[0m\n", statusList)
	fmt.Println(lineBreak)
	fmt.Printf("| \033[1;36mRepo Name:\033[0m \033[1;32m%s\033[0m\n", repoName)
	fmt.Printf("| \033[1;36mBranch:\033[0m \033[1;32m%s\033[0m\n", branch)
	fmt.Printf("| \033[1;36mBuild Type:\033[0m \033[1;32m%s\033[0m\n", buildType)
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mSearching for last successful execution...\033[0m")
	fmt.Println(lineBreak)

	// You can include a parameter here to select the source (payload or pipeline) for the commit hashes
	// source := "payload" // or "pipeline"
	source := p.Config.IngestionType

	var oldCommitHash, newCommitHash, branchName, repoNamePayload string
	var err error
	var isPrivate bool
	var pipeline models.Pipeline

	if source == "payload" {
		fmt.Println(lineBreak)
		fmt.Println("| \033[1;36mParsing payload...\033[0m")
		fmt.Println(lineBreak)
		// Read json fro file
		payload, err := os.ReadFile("bitbucket/payload-bitbucket-harness.json")
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}

		// Get the old and new commit hashes from the payload
		// payload := `...` // Your JSON payload here

		oldCommitHash, newCommitHash, branchName, repoNamePayload, isPrivate, err = ParsePayload(payload)
		if err != nil {
			log.Fatalf("Error parsing payload: %v", err)
		}
		fmt.Println(lineBreak)
		fmt.Println("| Old Commit Hash: ", oldCommitHash)
		fmt.Println("| New Commit Hash: ", newCommitHash)
		fmt.Println("| Branch Name: ", branchName)
		fmt.Println("| Repository Name: ", repoNamePayload)
		fmt.Println("| Is Private: ", isPrivate)
		fmt.Println(lineBreak)

	} else if source == "pipeline" {
		branchName = branch
		fmt.Println(lineBreak)
		fmt.Println("| \033[1;36mGetting last successful execution...\033[0m")
		fmt.Println(lineBreak)
		fmt.Println("| Branch Name: ", branchName)
		fmt.Println("| Repo Name: ", repoName)
		fmt.Println("| Build Type: ", buildType)
		fmt.Println(lineBreak)
		// Get the old and new commit hashes from the pipeline

		oldCommitHash, newCommitHash, pipeline, err = getLastSuccessfulExecution(accID, orgID, projectID, pipelineID, stageID, statusList, repoName, branch, buildType)
		if err != nil {
			fmt.Println("Error getting last successful execution")
			fmt.Println(err)
			oldCommitHash = commitID
			newCommitHash = commitID
			// return err

		}
		fmt.Println(lineBreak)
		if pipeline.Status == "Success" {
			fmt.Println("| \033[1;32mLast successful execution found\033[0m")
			fmt.Println(lineBreak)
			fmt.Printf("| \033[1;36mPipeline Name:\033[0m \033[1;32m%s\033[0m\n", pipeline.Name)
			fmt.Printf("| \033[1;36mPipeline Status:\033[0m \033[1;32m%s\033[0m\n", pipeline.Status)
			fmt.Printf("| \033[1;36mPipeline Started Time:\033[0m \033[1;32m%s\033[0m\n", pipeline.StartedTime)
			fmt.Printf("| \033[1;36mPipeline Duration:\033[0m \033[1;32m%s\033[0m\n", pipeline.Duration)
			fmt.Printf("| \033[1;36mPipeline Stage Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StageCount)
			fmt.Printf("| \033[1;36mPipeline Step Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StepCount)
			fmt.Printf("| \033[1;36mPipeline Message:\033[0m \033[1;32m%s\033[0m\n", pipeline.Message)
			fmt.Println(lineBreak)
		} else {
			fmt.Println("| \033[1;31mLast successful execution not found\033[0m")
			fmt.Println(lineBreak)
		}

	}

	// commitInfo, err := GetCommitInfo(oldCommitHash, newCommitHash)
	// if err != nil {
	// 	log.Fatalf("Error getting git info: %v", err)
	// }

	// fmt.Println("Commit Info: ", commitInfo)

	fmt.Printf("| \033[1;33mFirst Commit SHA:\033[0m %s\n", oldCommitHash)
	fmt.Println(lineBreak)
	fmt.Printf("| \033[1;33mLast Commit SHA:\033[0m %s\n", newCommitHash)
	fmt.Println(lineBreak)
	// git()
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mSearching for commit info...\033[0m")
	fmt.Println(lineBreak)

	// result, err := GetCommitInfo("e7c79ef9dcaa60c41c88ea5417b977bffe0bdb9f", "HEAD")
	result, err := GetCommitInfo(oldCommitHash, "HEAD")
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mGit Commit Info\033[0m")
	fmt.Println(lineBreak)
	for _, fileInfo := range result {
		fmt.Printf("| \033[1;36mFile Name:\033[0m \033[1;32m%s\033[0m\n", fileInfo.FileName)
		for _, commitInfo := range fileInfo.CommitDetails {
			fmt.Printf("| \033[1;36mCommit Hash:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Hash)
			fmt.Printf("| \033[1;36mCommit Name:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Name)
			fmt.Printf("| \033[1;36mCommit Email:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Email)
			fmt.Printf("| \033[1;36mCommit Username:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Username)
			fmt.Printf("| \033[1;36mCommit Author Time:\033[0m \033[1;32m%s\033[0m\n", commitInfo.AuthorTime)
			fmt.Printf("| \033[1;36mCommit Committer Name:\033[0m \033[1;32m%s\033[0m\n", commitInfo.CommitterName)
			fmt.Printf("| \033[1;36mCommit Committer Email:\033[0m \033[1;32m%s\033[0m\n", commitInfo.CommitterEmail)
			fmt.Printf("| \033[1;36mCommit Ref Names:\033[0m \033[1;32m%s\033[0m\n", commitInfo.RefNames)
			fmt.Printf("| \033[1;36mCommit Title:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Title)
			fmt.Printf("| \033[1;36mCommit Body:\033[0m \033[1;32m%s\033[0m\n", commitInfo.Body)
			fmt.Printf("| \033[1;36mCommit Parent Hashes:\033[0m \033[1;32m%s\033[0m\n", commitInfo.ParentHashes)
			fmt.Println("| \033[1;36mFile Changes:\033[0m")
			for _, change := range commitInfo.Changes {
				fmt.Printf("| \033[1;36mFile Name:\033[0m \033[1;32m%s\033[0m | \033[1;36mStatus:\033[0m \033[1;32m%s\033[0m\n", change.FileName, change.Status)
			}
			fmt.Println(lineBreak)
		}
	}

	// Prepare maps to store committers and participants (to avoid duplicates)
	var committers map[string]struct{} = make(map[string]struct{})
	var committersName map[string]struct{} = make(map[string]struct{})
	var participants map[string]struct{} = make(map[string]struct{})
	var participantsName map[string]struct{} = make(map[string]struct{})

	// Prepare a slice to store file changes
	var fileChanges []struct {
		FileName   template.HTML
		Status     string
		Committer  string
		Reviewer   string // Assume you have this data; if not, adjust accordingly
		CommitHash string
		Title      string
		Time       string
	}

	// Inside the loop where you are iterating over commitInfo, gather the necessary data:
	for _, fileInfo := range result {
		for _, commitInfo := range fileInfo.CommitDetails {

			// Add committers and participants to the maps
			committers[commitInfo.Email] = struct{}{}
			committersName[commitInfo.Name] = struct{}{}
			participants[commitInfo.Name] = struct{}{}
			participantsName[commitInfo.CommitterName] = struct{}{}

			// Add file changes to the slice
			for _, change := range commitInfo.Changes {
				fileChanges = append(fileChanges, struct {
					FileName   template.HTML
					Status     string
					Committer  string
					Reviewer   string
					CommitHash string
					Title      string
					Time       string
				}{
					FileName:   template.HTML(change.FileName),
					Status:     change.Status,
					Committer:  commitInfo.Name, // Adjust as per your data structure
					Reviewer:   "",              // Adjust to include reviewer information, if available
					CommitHash: commitInfo.Hash,
					Title:      commitInfo.Title,
					Time:       commitInfo.AuthorTime,
				})
			}
		}
	}

	// Convert maps to slices for committers and participants
	var committersList []string
	for committer := range committers {
		committersList = append(committersList, committer)
	}
	_ = committersList // Fix SA4010 by assigning the result of append to a variable

	var committersNameList []string
	for committerName := range committersName {
		committersNameList = append(committersNameList, committerName)
	}

	var participantsNameList []string
	for participantName := range participantsName {
		participantsNameList = append(participantsNameList, participantName)
	}
	_ = participantsNameList // Fix SA4010 by assigning the result of append to a variable

	var participantsList []string
	for participant := range participants {
		participantsList = append(participantsList, participant)
	}

	// convert 1695494200 to Date and Time format MM/DD/YYYY HH:MM:SS AM/PM GMT e.g. 09/22/2021 12:00:00 AM -3:00

	createdStr := os.Getenv("CI_BUILD_CREATED")

	created, err := strconv.ParseInt(createdStr, 10, 64)
	if err != nil {
		fmt.Println("\033[33m| No CI_BUILD_CREATED env variable found\033[0m")
		created = time.Now().Unix()
	}

	t := time.Unix(created, 0).UTC() // Make sure it's in UTC

	// add current timezone offset to the time
	_, offsetSeconds := time.Now().Zone()

	// Convert offset in seconds to time.Duration
	offset := time.Duration(offsetSeconds) * time.Second

	t = t.Add(offset)

	currentTimezone := time.Now().Format("-0700")
	// fmt.Println(currentTimezone)

	fmt.Println("| Current Pipeline Build Created Date/Time: " + t.Format("01/02/2006 03:04:05 PM "+currentTimezone))

	createdStr = t.Format("01/02/2006 03:04:05 PM " + currentTimezone)

	// fmt.Println("Pipe URL: " + p.Config.PipeExecutionURL)
	// Call the GenerateReport function
	report, err := GenerateReport(repoName, branchName, buildType, committersNameList, pipeline.Name, p.Config.PipeExecutionURL, fileChanges, createdStr)
	if err != nil {
		return err
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mGit Commit Report\033[0m")
	fmt.Println(lineBreak)
	// fmt.Println(report)
	// fmt.Println(lineBreak)

	//save to a html file
	err = os.WriteFile("report.html", []byte(report), 0644)
	if err != nil {
		return err
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mGit Commit Report saved to report.html\033[0m")
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mCommit Insights completed successfully\033[0m")
	fmt.Println(lineBreak)

	return nil
}

func parsePipeline(jsonData []byte) (*models.Pipeline, error) {
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mParsing pipeline...\033[0m")
	fmt.Println(lineBreak)

	var response models.Response
	err := json.Unmarshal(jsonData, &response)
	if err != nil {
		return nil, err
	}
	duration := time.Duration(response.Data.Content[0].EndTs-response.Data.Content[0].StartTs) * time.Millisecond
	var durationStr string
	if duration < time.Minute {
		durationStr = fmt.Sprintf("%.0f seconds", duration.Seconds())
	} else if duration < time.Hour {
		durationStr = fmt.Sprintf("%.0f minutes", duration.Minutes())
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		durationStr = fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}

	pipeline := models.Pipeline{
		Name:        response.Data.Content[0].Name,
		Status:      response.Data.Content[0].Status,
		StartedTime: time.Unix(int64(response.Data.Content[0].StartTs/1000), 0).String(),
		Duration:    durationStr,
		StageCount:  response.Data.Content[0].TotalStagesCount,
		StepCount:   response.Data.Content[0].SuccessfulStagesCount + response.Data.Content[0].FailedStagesCount,
		Message:     "",
	}

	layoutNodeMap := response.Data.Content[0].LayoutNodeMap

	for _, nodeInfo := range layoutNodeMap {
		if nodeInfo.NodeGroup == "STAGE" {
			stage := models.Stage{
				Name:   nodeInfo.Name,
				Module: nodeInfo.Module,
			}

			for _, childId := range nodeInfo.EdgeLayoutList.CurrentNodeChildren {
				childNodeInfo := layoutNodeMap[childId]
				step := models.Step{
					Name:   childNodeInfo.Name,
					Status: childNodeInfo.Status,
				}
				stage.Steps = append(stage.Steps, step)
			}

			pipeline.Stages = append(pipeline.Stages, stage)
		}
	}

	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mPipeline Info\033[0m")
	fmt.Println(lineBreak)
	fmt.Printf("| \033[1;36mPipeline Name:\033[0m \033[1;32m%s\033[0m\n", pipeline.Name)
	fmt.Printf("| \033[1;36mPipeline Status:\033[0m \033[1;32m%s\033[0m\n", pipeline.Status)
	fmt.Printf("| \033[1;36mPipeline Started Time:\033[0m \033[1;32m%s\033[0m\n", pipeline.StartedTime)
	fmt.Printf("| \033[1;36mPipeline Duration:\033[0m \033[1;32m%s\033[0m\n", pipeline.Duration)
	fmt.Printf("| \033[1;36mPipeline Stage Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StageCount)
	fmt.Printf("| \033[1;36mPipeline Step Count:\033[0m \033[1;32m%d\033[0m\n", pipeline.StepCount)
	fmt.Printf("| \033[1;36mPipeline Message:\033[0m \033[1;32m%s\033[0m\n", pipeline.Message)
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mPipeline Stages\033[0m")
	fmt.Println(lineBreak)
	for _, stage := range pipeline.Stages {
		fmt.Printf("| \033[1;36mStage Name:\033[0m \033[1;32m%s\033[0m\n", stage.Name)
		fmt.Printf("| \033[1;36mStage Module:\033[0m \033[1;32m%s\033[0m\n", stage.Module)

		// fmt.Printf("| \033[1;36mStage Step Count:\033[0m \033[1;32m%d\033[0m\n", len(stage.Steps))
		fmt.Println(lineBreak)
		// fmt.Println("| \033[1;36mStage Steps\033[0m")
		// fmt.Println(lineBreak)
		// for _, step := range stage.Steps {
		// 	fmt.Printf("| \033[1;36mStep Name:\033[0m \033[1;32m%s\033[0m\n", step.Name)
		// 	fmt.Printf("| \033[1;36mStep Status:\033[0m \033[1;32m%s\033[0m\n", step.Status)
		// 	fmt.Printf("| \033[1;36mStep Message:\033[0m \033[1;32m%s\033[0m\n", step.Message)
		// 	fmt.Println(lineBreak)
		// }
	}

	fmt.Println(" Json: ", pipeline)

	return &pipeline, nil
}

// func Parse models.PayloadSteps(jsonData []byte) (* models.PayloadSteps, error) {
func ParsePayloadSteps() (*models.PayloadSteps, error) {
	fmt.Println(lineBreak)
	fmt.Println("| \033[1;36mParsing payload...\033[0m")
	fmt.Println(lineBreak)

	file, err := os.Open("pipegraph.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	// Read the file content
	data, err := os.ReadFile("pipegraph.json")
	// data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	var payloadSteps models.PayloadSteps
	//convert data to strings
	// fmt.Println("Data: ", string(data))
	// fmt.Println("Data: ", data)
	err = json.Unmarshal(data, &payloadSteps)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return nil, err
	}

	for _, node := range payloadSteps.Data.ExecutionGraph.NodeMap {
		fmt.Printf("Step: %s (%s)\n", node.Name, node.Identifier)
		fmt.Printf("Status: %s\n", node.Status)
		if node.Status == "Failed" {
			fmt.Printf("Failure Message: %s\n", node.FailureInfo.Message)
			fmt.Printf("Failure Types: %v\n", node.FailureInfo.FailureTypeList)
		}
		fmt.Printf("Start Time: %s\n", time.Unix(0, node.StartTs*int64(time.Millisecond)).UTC())
		fmt.Printf("End Time: %s\n", time.Unix(0, node.EndTs*int64(time.Millisecond)).UTC())
		fmt.Println("---")
	}

	return &payloadSteps, nil
}
