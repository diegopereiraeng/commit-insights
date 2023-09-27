
# Commit-Insights

Commit-Insights is a tool that provides insights into your codebase by analyzing Git commit history and harnessing data. It's designed to work as a plugin in your CI/CD pipeline.

## Features

- Gather commit history between two Git SHAs
- Extract and process commit metadata
- Extract file change information from each commit
- Output data in a structured format for further analysis

## Requirements

- Go 1.16+
- Git

## Installation

1. Clone the repository

    ```bash
    git clone https://github.com/diegopereiraeng/commit-insights.git
    ```

2. Navigate to the project directory

    ```bash
    cd commit-insights
    ```

## Run Locally

1. Download the Go modules required for the project.

    ```bash
    go mod download
    ```

2. Build the project to generate the binary.

    ```bash
    go build -o commit-insights
    ```

### Sample Example

You can run the tool using the following command as an example:

```bash
./commit-insights --acc_id=<YOUR_ACCOUNT_ID> --orgID=default --projectID=GIT_FLOW_DEMO --pipelineID=Plugin_Factory --stageID=Build_Golang --statusList=Success --repoName=commit-insights --branch=main --buildType=branch --ingestionType=pipeline --harness_secret=<YOUR_HARNESS_READ_TOKEN>
```

## Usage in Pipeline

Here is a sample example of how you can use the Commit-Insights in a pipeline:

\`\`\`yaml
- step:
    type: Plugin
    name: Commit-Insights
    identifier: CommitInsights
    spec:
      connectorRef: account.DockerHubDiego
      image: diegokoala/commit-insights:latest
      settings:
        harness_secret: <+secrets.getValue("harnesssatokenplugin")>
      imagePullPolicy: Always
\`\`\`

## Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a pull request
