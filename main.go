package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var build = "1" // build number set at compile-time

func main() {
	app := cli.NewApp()
	app.Name = "Commit-Insights-Plugin"
	app.Usage = "CLI tool to ingest commit insights into Harness"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "acc_id",
			Usage:  "e.g: 2_gVHyo9Qiu4dXvj-AcbC",
			EnvVar: "HARNESS_ACCOUNT_ID, PLUGIN_ACC_ID",
		},
		cli.StringFlag{
			Name:   "orgID",
			Usage:  "default",
			EnvVar: "HARNESS_ORG_ID, PLUGIN_ORG_ID",
		},
		cli.StringFlag{
			Name:   "projectID",
			Usage:  "GIT_FLOW_DEMO",
			EnvVar: "HARNESS_PROJECT_ID, PLUGIN_PROJECT_ID",
		},
		cli.StringFlag{
			Name:   "pipelineID",
			Usage:  "FAST_CISTO_SonarQube_Quality_Gate_Plugin",
			EnvVar: "HARNESS_PIPELINE_ID, PLUGIN_PIPELINE_ID",
		},
		cli.StringFlag{
			Name:   "stageID",
			Usage:  "Build_Golang",
			EnvVar: "HARNESS_STAGE_ID, PLUGIN_STAGE_ID",
		},
		cli.StringSliceFlag{
			Name:   "statusList",
			Usage:  "Comma-separated list of statuses to filter by. E.g: Success,Aborted",
			Value:  &cli.StringSlice{"Success"},
			EnvVar: "PLUGIN_STATUS_LIST",
		},
		cli.StringFlag{
			Name:   "repoName",
			Usage:  "sonarqube-scanner",
			EnvVar: "DRONE_REPO_NAME, PLUGIN_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "branch",
			Usage:  "main",
			EnvVar: "CI_COMMIT_BRANCH, PLUGIN_BRANCH",
		},
		cli.StringFlag{
			Name:   "buildType",
			Usage:  "push or pull_request",
			EnvVar: "DRONE_BUILD_EVENT, PLUGIN_BUILD_TYPE",
		},
		cli.StringFlag{
			Name:   "ingestionType",
			Usage:  "payload or pipeline",
			Value:  "pipeline",
			EnvVar: "PLUGIN_INGESTION_TYPE",
		},
		cli.StringFlag{
			Name:   "commit_id",
			Usage:  "e7c79ef9dcaa60c41c88ea5417b977bffe0bdb9f",
			EnvVar: "CI_COMMIT_SHA, PLUGIN_COMMIT_ID",
		},
		cli.StringFlag{
			Name:   "harness_secret",
			Usage:  "Harness access token with visualization permissions",
			EnvVar: "PLUGIN_HARNESS_SECRET",
		},
		cli.StringFlag{
			Name:   "harness_pipe_execution_url",
			Usage:  "Provide a custom Harness execution URL, or it gonna take the current pipeline execution URL (Optional)",
			EnvVar: "CI_BUILD_LINK, PLUGIN_HARNESS_PIPE_EXECUTION_URL",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) {
	if c.String("json_file_name") != "" && c.String("json_content") != "" {
		fmt.Println("Error: Please specify either json_file_name or json_content, but not both.")
		os.Exit(1)
	}

	config := Config{
		AccID:            c.String("acc_id"),
		OrgID:            c.String("orgID"),
		ProjectID:        c.String("projectID"),
		PipelineID:       c.String("pipelineID"),
		StageID:          c.String("stageID"),
		StatusList:       c.StringSlice("statusList"),
		RepoName:         c.String("repoName"),
		Branch:           c.String("branch"),
		BuildType:        c.String("buildType"),
		IngestionType:    c.String("ingestionType"),
		CommitID:         c.String("commit_id"),
		HarnessSecret:    c.String("harness_secret"),
		PipeExecutionURL: c.String("harness_pipe_execution_url"),
	}

	plugin := Plugin{Config: config}
	if err := plugin.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
