// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package preflight_kit_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"github.com/steadybit/preflight-kit/go/preflight_kit_sdk/state_persister"
)

const (
	defaultCallInterval  = "1s"
	minHeartbeatInterval = 5 * time.Second
)

type preflightHttpAdapter[T any] struct {
	description preflight_kit_api.PreflightDescription
	preflight   Preflight[T]
	rootPath    string
}

func newPreflightHttpAdapter[T any](preflight Preflight[T]) *preflightHttpAdapter[T] {
	description := getDescriptionWithDefaults(preflight)
	adapter := &preflightHttpAdapter[T]{
		description: description,
		preflight:   preflight,
		rootPath:    fmt.Sprintf("/%s", description.Id),
	}
	return adapter
}

func (a *preflightHttpAdapter[T]) handleGetDescription(w http.ResponseWriter, _ *http.Request, _ []byte) {
	exthttp.WriteBody(w, a.description)
}

func (a *preflightHttpAdapter[T]) handleStart(w http.ResponseWriter, r *http.Request, body []byte) {
	parsedBody, err, done := parseStartRequest(w, body)
	if done {
		return
	}
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body.", err))
		return
	}

	state := a.preflight.NewEmptyState()
	result, err := a.preflight.Start(r.Context(), &state)
	if result == nil {
		result = &preflight_kit_api.StartResult{}
	}
	if err != nil {
		if extensionError, ok := err.(extension_kit.ExtensionError); ok {
			exthttp.WriteError(w, extensionError)
		} else {
			exthttp.WriteError(w, extension_kit.ToError("Failed to start preflight.", err))
		}
		return
	}

	var convertedState preflight_kit_api.PreflightState
	err = extconversion.Convert(state, &convertedState)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode action state.", err))
		return
	}
	result.State = convertedState

	if a.description.Cancel != nil {
		err = statePersister.PersistState(r.Context(), &state_persister.PersistedState{PreflightActionExecutionId: parsedBody.PreflightActionExecutionId, PreflightActionId: a.description.Id, State: convertedState})
		if err != nil {
			exthttp.WriteError(w, extension_kit.ToError("Failed to persist preflightAction state.", err))
			return
		}
		if a.description.Status.CallInterval != nil {
			interval, err := time.ParseDuration(*a.description.Status.CallInterval)
			if interval < minHeartbeatInterval {
				interval = minHeartbeatInterval
			}
			if err == nil {
				monitorHeartbeat(parsedBody.PreflightActionExecutionId, interval, interval*4)
			}
		}

	}
	exthttp.WriteBody(w, result)
}

func parseStartRequest(w http.ResponseWriter, body []byte) (preflight_kit_api.StartPreflightRequestBody, error, bool) {
	var parsedBody preflight_kit_api.StartPreflightRequestBody
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body.", err))
		return preflight_kit_api.StartPreflightRequestBody{}, nil, true
	}
	return parsedBody, err, false
}

func (a *preflightHttpAdapter[T]) handleStatus(w http.ResponseWriter, r *http.Request, body []byte) {
	var parsedBody preflight_kit_api.StatusPreflightRequestBody
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body.", err))
		return
	}

	recordHeartbeat(parsedBody.PreflightActionExecutionId)

	if stopEvent := getStopEvent(parsedBody.PreflightActionExecutionId); stopEvent != nil {
		exthttp.WriteBody(w, preflight_kit_api.StatusResult{
			Completed: true,
			Error: &preflight_kit_api.PreflightKitError{
				Title:  fmt.Sprintf("Preflight was stopped by extension: %s", stopEvent.reason),
				Status: extutil.Ptr(preflight_kit_api.Errored),
			},
		})
		return
	}

	preflight := a.preflight

	state := preflight.NewEmptyState()
	err = extconversion.Convert(parsedBody.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse state.", err))
		return
	}

	result, err := preflight.Status(r.Context(), &state)
	if result == nil {
		result = &preflight_kit_api.StatusResult{}
	}
	if err != nil {
		var extErr *extension_kit.ExtensionError
		if errors.As(err, &extErr) {
			exthttp.WriteError(w, *extErr)
		} else {
			exthttp.WriteError(w, extension_kit.ToError("Failed to read status.", err))
		}
		return
	}

	if result.State != nil {
		exthttp.WriteError(w, extension_kit.ToError("Please modify the state using the given state pointer.", err))
	}

	var convertedState preflight_kit_api.PreflightState
	err = extconversion.Convert(state, &convertedState)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode preflight state.", err))
		return
	}
	result.State = &convertedState

	if a.description.Cancel != nil {
		err = statePersister.PersistState(r.Context(), &state_persister.PersistedState{PreflightActionExecutionId: parsedBody.PreflightActionExecutionId, PreflightActionId: a.description.Id, State: convertedState})
		if err != nil {
			exthttp.WriteError(w, extension_kit.ToError("Failed to persist preflight state.", err))
			return
		}
	}
	exthttp.WriteBody(w, result)
}

func (a *preflightHttpAdapter[T]) hasCancel() bool {
	_, ok := a.preflight.(PreflightWithCancel[T])
	return ok
}

func (a *preflightHttpAdapter[T]) handleCancel(w http.ResponseWriter, r *http.Request, body []byte) {
	preflight := a.preflight.(PreflightWithCancel[T])

	var parsedBody preflight_kit_api.CancelPreflightRequestBody
	err := json.Unmarshal(body, &parsedBody)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse request body.", err))
		return
	}

	stopMonitorHeartbeat(parsedBody.PreflightActionExecutionId)

	if stopEvent := getStopEvent(parsedBody.PreflightActionExecutionId); stopEvent != nil {
		exthttp.WriteBody(w, preflight_kit_api.CancelResult{
			Error: &preflight_kit_api.PreflightKitError{
				Title: fmt.Sprintf("Preflight was stopped by extension %s", stopEvent.reason),
			},
		})
		return
	}

	state := preflight.NewEmptyState()
	err = extconversion.Convert(parsedBody.State, &state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to parse state.", err))
		return
	}

	result, err := preflight.Cancel(r.Context(), &state)
	if result == nil {
		result = &preflight_kit_api.CancelResult{}
	}
	if err != nil {
		extensionError, isExtensionError := err.(extension_kit.ExtensionError)
		if isExtensionError {
			exthttp.WriteError(w, extensionError)
		} else {
			exthttp.WriteError(w, extension_kit.ToError("Failed to cancel preflight.", err))
		}
		return
	}

	folder := fmt.Sprintf("/tmp/steadybit/%v", parsedBody.PreflightActionExecutionId)
	_, err = os.Stat(folder)
	if !os.IsNotExist(err) {
		err = os.RemoveAll(folder)
		if err != nil {
			log.Error().Msgf("Could not remove directory '%s'", folder)
		} else {
			log.Debug().Msgf("Directory '%s' removed successfully", folder)
		}
	}

	err = statePersister.DeleteState(r.Context(), parsedBody.PreflightActionExecutionId)
	if err != nil {
		log.Warn().
			Err(err).
			Str("preflightActionId", a.description.Id).
			Str("preflightActionExecutionId", parsedBody.PreflightActionExecutionId.String()).
			Msg("Failed to delete action state.")
		return
	}
	exthttp.WriteBody(w, result)
}

func (a *preflightHttpAdapter[T]) registerHandlers() {

	exthttp.RegisterHttpHandler(a.rootPath, a.handleGetDescription)
	exthttp.RegisterHttpHandler(a.description.Start.Path, a.handleStart)
	exthttp.RegisterHttpHandler(a.description.Status.Path, a.handleStatus)
	if a.hasCancel() {
		exthttp.RegisterHttpHandler(a.description.Cancel.Path, a.handleCancel)
	}
}

// getDescriptionWithDefaults wraps the preflight description and adds default paths and methods for prepare, start, status, cancel and metrics.
func getDescriptionWithDefaults[T any](preflight Preflight[T]) preflight_kit_api.PreflightDescription {
	description := preflight.Describe()
	if description.Start.Path == "" {
		description.Start.Path = fmt.Sprintf("/%s/start", description.Id)
	}
	if description.Start.Method == "" {
		description.Start.Method = preflight_kit_api.POST
	}
	if _, ok := preflight.(PreflightWithCancel[T]); ok && description.Cancel == nil {
		description.Cancel = &preflight_kit_api.MutatingEndpointReference{}
	}

	if description.Cancel != nil {
		if description.Cancel.Path == "" {
			description.Cancel.Path = fmt.Sprintf("/%s/cancel", description.Id)
		}
		if description.Cancel.Method == "" {
			description.Cancel.Method = preflight_kit_api.POST
		}
	}

	if description.Status.Path == "" {
		description.Status.Path = fmt.Sprintf("/%s/status", description.Id)
	}
	if description.Status.Method == "" {
		description.Status.Method = preflight_kit_api.POST
	}
	if description.Status.CallInterval == nil || *description.Status.CallInterval == "" {
		description.Status.CallInterval = extutil.Ptr(defaultCallInterval)
	}
	return description
}
