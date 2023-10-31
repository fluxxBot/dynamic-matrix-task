package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jfrog/jfrog-pipelines-tasks-sdk-go/tasks"
	"golang.org/x/exp/slices"
)

type AddDynamicSteplets struct {
	inputs Input
}

type Input struct {
	step_names           []string
	environment_variable map[string]interface{}
	nodePools            []string
	runtimes             map[string]interface{}
}

// Read implements io.Reader.
func (Input) Read(p []byte) (n int, err error) {
	panic("unimplemented")
}

var (
	readInput = tasks.GetInput
)

func haltExecution(errMessage string) {
	tasks.Error(errMessage)
	os.Exit(1)
}

func main() {
	tasks.Info("Starting task ...")
	status := "success"
	// Set greeting message as task output

	r := new(AddDynamicSteplets)
	r.readInputs()
	r.createSteplets()

	tasks.SetOutput("status", status)
}

func (r *AddDynamicSteplets) readInputs() {
	// Fetch inputs
	i := Input{}
	inputStepNames := readInput("step_names")
	if len(inputStepNames) == 0 {
		tasks.Error("Enter step names for which steplets are to be created")
		return
	}
	i.step_names = strings.Split(inputStepNames, ",")

	inputEnvironmentVariables := readInput("environment_variables")
	if len(inputEnvironmentVariables) == 0 {
		return
	}
	err := json.Unmarshal([]byte(inputEnvironmentVariables), &i.environment_variable)
	if err != nil {
		tasks.Error("Failed to parse Environment Variables input")
		return
	}

	inputNodePoolsVariables := readInput("nodePools")
	if len(inputEnvironmentVariables) == 0 {
		return
	}
	err = json.Unmarshal([]byte(inputNodePoolsVariables), &i.nodePools)
	if err != nil {
		tasks.Error("Failed to parse nodePools input")
		return
	}

	inputRuntimes := readInput("runtimes")
	if len(inputRuntimes) == 0 {
		return
	}
	err = json.Unmarshal([]byte(inputRuntimes), &i.runtimes)
	if err != nil {
		tasks.Error("Failed to parse runtimes input")
		return
	}

	r.inputs = i

	tasks.Debug(fmt.Sprintf("Received inputs are [%+v]", i))
}

func (r *AddDynamicSteplets) createSteplets() error {
	runId := getValue("run_id")
	step_name := getValue("step_name")
	apiToken := getValue("builder_api_token")
	pipelinesURL := getValue("pipelines_api_url")

	inputStepName := r.inputs.step_names
	if slices.Contains(inputStepName, step_name) {
		return errors.New("cannot create steplets for same step")
	}

	tasks.Info("RunId :- "+runId, "apiToken:-"+apiToken)
	req, err := http.NewRequest("Body", pipelinesURL+"/steps/"+step_name+"/"+runId+"/add_matrix_steplets", r.inputs)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "apiToken "+apiToken)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	tasks.Info(content)
	return nil
}

// getValue is a wrapper for tasks.GetVariable by handling error in case variable is not available
func getValue(varName string) string {
	value, err := tasks.GetVariable(varName)
	if err != nil {
		haltExecution("Failed to fetch " + varName)
	}
	return value
}
