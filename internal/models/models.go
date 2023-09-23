// models/models.go
package models

// HTML GENERATOR

// Pipeline represents a pipeline with its stages and steps.
type Pipeline struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	StartedTime string  `json:"startedTime"`
	Duration    string  `json:"duration"`
	StageCount  int     `json:"stageCount"`
	StepCount   int     `json:"stepCount"`
	Message     string  `json:"message"`
	Stages      []Stage `json:"stages"`
}

// Stage represents a stage in a pipeline with its steps.
type Stage struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Module   string `json:"module"`
	StartTs  string `json:"startTs"`
	EndTs    string `json:"endTs"`
	Duration string `json:"duration"`
	Steps    []Step `json:"steps"`
}

// Step represents a step in a stage.
type Step struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	StartTs  string `json:"startTs"`
	EndTs    string `json:"endTs"`
	Duration string `json:"duration"`
}

// steps parsing
type PayloadSteps struct {
	Status string `json:"status"`
	Data   struct {
		ExecutionGraph struct {
			NodeMap map[string]Node `json:"nodeMap"`
		} `json:"executionGraph"`
	} `json:"data"`
}

type Node struct {
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	StartTs     int64  `json:"startTs"`
	EndTs       int64  `json:"endTs"`
	Status      string `json:"status"`
	StepType    string `json:"stepType"`
	FailureInfo struct {
		Message         string   `json:"message"`
		FailureTypeList []string `json:"failureTypeList"`
	} `json:"failureInfo"`
}

// PLUGIN CORE

// define Config and Plugin structure like this

type Commit struct {
	Recast string `json:"__recast"`
	ID     string `json:"id"`
	// Include other fields if needed
}

type BranchInfo struct {
	Recast  string   `json:"__recast"`
	Commits []Commit `json:"commits"`
	// Include other fields if needed
}

type CIExecutionInfoDTO struct {
	Branch BranchInfo `json:"branch"`
	// Include other fields if needed
}

type Branch struct {
	Recast  string   `json:"__recast"`
	Commits []Commit `json:"commits"`
}

type CI struct {
	CIExecutionInfoDTO CIExecutionInfoDTO `json:"ciExecutionInfoDTO"`
	// Branch             BranchInfo         `json:"branch"`
	// Include other fields if needed
}

type ModuleInfo struct {
	CI CI `json:"ci"`
	// Include other fields if needed
}

type Content struct {
	ModuleInfo            ModuleInfo    `json:"moduleInfo"`
	LayoutNodeMap         LayoutNodeMap `json:"layoutNodeMap"`
	PlanExecutionId       string        `json:"planExecutionId"`
	Status                string        `json:"status"`
	Name                  string        `json:"name"`
	StartTs               int           `json:"startTs"`
	EndTs                 int           `json:"endTs"`
	SuccessfulStagesCount int           `json:"successfulStagesCount"`
	FailedStagesCount     int           `json:"failedStagesCount"`
	TotalStagesCount      int           `json:"totalStagesCount"`
	StartingNodeId        string        `json:"startingNodeId"`
	ExecutionTriggerInfo  struct {
		TriggerType string `json:"triggerType"`
		TriggeredBy struct {
			Identifier string `json:"identifier"`
			ExtraInfo  struct {
				Email string `json:"email"`
			}
		}
		IsRerun bool `json:"isRerun"`
	} `json:"executionTriggerInfo"`
	// Include other fields if needed
}

type Data struct {
	Content []Content `json:"content"`
	// Include other fields if needed
}

type Response struct {
	Data   Data   `json:"data"`
	Status string `json:"status"`
	// Include other fields if needed
}

type NodeInfo struct {
	NodeType       string     `json:"nodeType"`
	NodeGroup      string     `json:"nodeGroup"`
	NodeIdentifier string     `json:"nodeIdentifier"`
	Name           string     `json:"name"`
	NodeUuid       string     `json:"nodeUuid"`
	Status         string     `json:"status"`
	Module         string     `json:"module"`
	ModuleInfo     ModuleInfo `json:"moduleInfo"`
	StartTs        int        `json:"startTs"`
	EndTs          int        `json:"endTs"`
	FailureInfo    struct {
		Message string `json:"message"`
	} `json:"failureInfo"`
	EdgeLayoutList EdgeLayout `json:"edgeLayoutList"`
	// NodeExecutionId string `json:"nodeExecutionId"`
	// Include other fields if needed
}

type EdgeLayout struct {
	CurrentNodeChildren []string `json:"currentNodeChildren"`
	NextIds             []string `json:"nextIds"`
	// ... (other fields)
}

type LayoutNodeMap map[string]NodeInfo
