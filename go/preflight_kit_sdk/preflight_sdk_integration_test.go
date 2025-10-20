// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package preflight_kit_sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extsignals"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

var (
	ANY_ARG = struct{}{}
)

type PreflightOperations struct {
	executionId uuid.UUID
	basePath    string
	description preflight_kit_api.PreflightDescription
	calls       <-chan Call
	preflight   *ExamplePreflight
}

type TestCase struct {
	Name string
	Fn   func(t *testing.T, op PreflightOperations)
}

func Test_SDK(t *testing.T) {
	testCases := []TestCase{
		{
			Name: "should run a simple preflight",
			Fn:   testcaseSimple,
		},
		{
			Name: "should cancel preflights on heartbeat timeout",
			Fn:   testcaseHeartbeatTimeout,
		},
		{
			Name: "should return error from status",
			Fn:   testCaseStatusWithGenericError,
		},
		{
			Name: "should return extension error from status",
			Fn:   testCaseStatusWithExtensionKitError,
		},
	}
	calls := make(chan Call, 1024)
	defer close(calls)

	serverPort, err := freeport.GetFreePort()
	require.NoError(t, err)

	preflight := NewExamplePreflight(calls)
	go func(preflight *ExamplePreflight) {
		extlogging.InitZeroLog()
		RegisterPreflight(preflight)
		exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(GetPreflightList))
		extsignals.ActivateSignalHandlers()
		exthttp.Listen(exthttp.ListenOpts{Port: serverPort})
	}(preflight)
	time.Sleep(1 * time.Second)

	basePath := fmt.Sprintf("http://localhost:%d", serverPort)
	preflightPath := listExtension(t, basePath)
	description := describe(t, fmt.Sprintf("%s%s", basePath, preflightPath))

	for _, testCase := range testCases {
		op := PreflightOperations{
			basePath:    basePath,
			description: description,
			executionId: uuid.New(),
			calls:       calls,
			preflight:   preflight,
		}

		op.resetCalls()
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Fn(t, op)
		})
	}

	fmt.Println("Yes, IntelliJ, yes, the test is finished.")
}

func testcaseSimple(t *testing.T, op PreflightOperations) {
	result := op.start(t)
	assertPrepareResult(t, result)
	op.assertCall(t, "Start", ANY_ARG)
	state := result.State

	state = op.status(t, state)
	op.assertCall(t, "Status", toExampleState(state))

	op.cancel(t, state)
	op.assertCall(t, "Cancel", toExampleState(state))
}

func testcaseHeartbeatTimeout(t *testing.T, op PreflightOperations) {
	result := op.start(t)
	state := result.State
	op.resetCalls()

	time.Sleep(25 * time.Second)
	op.assertCall(t, "Cancel", ANY_ARG)

	statusResult, _ := op.statusResult(t, state)
	require.NotNil(t, statusResult.Error)
	assert.Equal(t, preflight_kit_api.Errored, *statusResult.Error.Status)
	assert.Equal(t, "Preflight was stopped by extension: heartbeat timeout", statusResult.Error.Title)
}

func testCaseStatusWithGenericError(t *testing.T, op PreflightOperations) {
	op.preflight.statusError = fmt.Errorf("this is a test error")
	statusResult, err := op.statusResult(t, preflight_kit_api.PreflightState{})
	assert.Equal(t, &preflight_kit_api.PreflightKitError{Title: "Failed to read preflight status.", Detail: extutil.Ptr("this is a test error")}, statusResult.Error)
	assert.Nil(t, err)
	op.assertCall(t, "Status", ANY_ARG)
}

func testCaseStatusWithExtensionKitError(t *testing.T, op PreflightOperations) {
	op.preflight.statusError = extutil.Ptr(extension_kit.ToError("this is a test error", errors.New("with some details")))
	statusResult, err := op.statusResult(t, preflight_kit_api.PreflightState{})
	assert.Equal(t, &preflight_kit_api.PreflightKitError{Title: "this is a test error", Detail: extutil.Ptr("with some details")}, statusResult.Error)
	assert.Nil(t, err)
	op.assertCall(t, "Status", ANY_ARG)
}

func assertPrepareResult(t *testing.T, response preflight_kit_api.StartResult) {
	assert.Equal(t, "Prepare", response.State["TestStep"])

	executionIds, err := statePersister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	assert.Len(t, executionIds, 1)

	state, err := statePersister.GetState(context.Background(), executionIds[0])
	require.NoError(t, err)
	assert.Equal(t, "Prepare", (*state).State["TestStep"])
}

func listExtension(t *testing.T, path string) string {
	res, err := http.Get(path)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	response := preflight_kit_api.PreflightList{}
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Preflights)
	return response.Preflights[0].Path
}

func describe(t *testing.T, preflightPath string) preflight_kit_api.PreflightDescription {
	res, err := http.Get(preflightPath)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	var response preflight_kit_api.PreflightDescription
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Equal(t, "ExamplePreflightId", response.Id)
	assert.NotNil(t, response.Start)
	assert.NotNil(t, response.Status)
	assert.NotNil(t, response.Cancel)
	return response
}

func (op *PreflightOperations) start(t *testing.T) preflight_kit_api.StartResult {
	startBody := preflight_kit_api.StartPreflightRequestBody{PreflightActionExecutionId: op.executionId, ExperimentExecution: preflight_kit_api.ExperimentExecutionAO{}}
	jsonBody, err := json.Marshal(startBody)
	require.NoError(t, err)
	bodyReader := bytes.NewReader(jsonBody)
	res, err := http.Post(fmt.Sprintf("%s%s", op.basePath, op.description.Start.Path), "application/json", bodyReader)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	var response preflight_kit_api.StartResult
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Nil(t, response.Error)

	return response
}

func (op *PreflightOperations) status(t *testing.T, state preflight_kit_api.PreflightState) preflight_kit_api.PreflightState {
	response, _ := op.statusResult(t, state)

	assert.Equal(t, "Status", (*response.State)["TestStep"])

	executionIds, err := statePersister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	assert.Len(t, executionIds, 1)

	pState, err := statePersister.GetState(context.Background(), executionIds[0])
	require.NoError(t, err)
	assert.Equal(t, "Status", (*pState).State["TestStep"])

	return *response.State
}

func (op *PreflightOperations) statusResult(t *testing.T, state preflight_kit_api.PreflightState) (*preflight_kit_api.StatusResult, *preflight_kit_api.PreflightKitError) {
	statusBody := preflight_kit_api.StatusPreflightRequestBody{State: state, PreflightActionExecutionId: op.executionId}
	jsonBody, err := json.Marshal(statusBody)
	require.NoError(t, err)
	bodyReader := bytes.NewReader(jsonBody)
	res, err := http.Post(fmt.Sprintf("%s%s", op.basePath, op.description.Status.Path), "application/json", bodyReader)
	if res.StatusCode != http.StatusOK {
		var response preflight_kit_api.PreflightKitError
		err = json.NewDecoder(res.Body).Decode(&response)
		require.NoError(t, err)
		return nil, &response
	}

	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	var response preflight_kit_api.StatusResult
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	return &response, nil
}

func (op *PreflightOperations) cancel(t *testing.T, state preflight_kit_api.PreflightState) {
	statusBody := preflight_kit_api.StatusPreflightRequestBody{State: state, PreflightActionExecutionId: op.executionId}
	jsonBody, err := json.Marshal(statusBody)
	require.NoError(t, err)
	bodyReader := bytes.NewReader(jsonBody)
	res, err := http.Post(fmt.Sprintf("%s%s", op.basePath, op.description.Cancel.Path), "application/json", bodyReader)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	var response preflight_kit_api.CancelResult
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	executionIds, err := statePersister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	assert.Len(t, executionIds, 0)
}

func (op *PreflightOperations) resetCalls() {
	for len(op.calls) > 0 {
		<-op.calls
	}
}

func (op *PreflightOperations) assertCall(t *testing.T, name string, args ...interface{}) {
	select {
	case call := <-op.calls:
		assert.Equal(t, name, call.Name)
		assert.Equal(t, len(args), len(call.Args), "Arguments differ in length")
		for i, expected := range args {
			if expected == ANY_ARG {
				continue
			}
			actual := call.Args[i]
			fmt.Printf("Expected: %v, Actual: %v", &expected, actual)
			assert.EqualValues(t, expected, actual)
		}
	case <-time.After(1 * time.Second):
		assert.Fail(t, "No call to received", "Expected call to %s", name)
	}
}
