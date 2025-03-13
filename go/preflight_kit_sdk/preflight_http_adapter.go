// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package preflight_kit_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"github.com/steadybit/preflight-kit/go/preflight_kit_sdk/state_persister"
	"net/http"
	"os"
	"time"
)

const (
	defaultCallInterval  = "1s"
	minHeartbeatInterval = 5 * time.Second
)

type preflightHttpAdapter struct {
	description preflight_kit_api.PreflightDescription
	preflight   Preflight
	rootPath    string
}

func newPreflightHttpAdapter(preflight Preflight) *preflightHttpAdapter {
	description := getDescriptionWithDefaults(preflight)
	adapter := &preflightHttpAdapter{
		description: description,
		preflight:   preflight,
		rootPath:    fmt.Sprintf("/%s", description.Id),
	}
	return adapter
}

func (a *preflightHttpAdapter) handleGetDescription(w http.ResponseWriter, _ *http.Request, _ []byte) {
	exthttp.WriteBody(w, a.description)
}

func (a *preflightHttpAdapter) handleStart(w http.ResponseWriter, r *http.Request, body []byte) {
	parsedBody, err, done := parseStartRequest(w, body)
	if done {
		return
	}

	result, err := a.preflight.Start(r.Context(), parsedBody)
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

	if a.description.Cancel != nil {
		err = statePersister.PersistState(r.Context(), &state_persister.PersistedState{PreflightActionExecutionId: parsedBody.PreflightActionExecutionId, PreflightActionId: a.description.Id})
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

func (a *preflightHttpAdapter) handleStatus(w http.ResponseWriter, r *http.Request, body []byte) {
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
	result, err := preflight.Status(r.Context(), parsedBody)
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
	exthttp.WriteBody(w, result)
}

func (a *preflightHttpAdapter) hasCancel() bool {
	_, ok := a.preflight.(PreflightWithCancel)
	return ok
}

func (a *preflightHttpAdapter) handleCancel(w http.ResponseWriter, r *http.Request, body []byte) {
	preflight := a.preflight.(PreflightWithCancel)

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

	result, err := preflight.Cancel(r.Context(), parsedBody)
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

func (a *preflightHttpAdapter) registerHandlers() {

	exthttp.RegisterHttpHandler(a.rootPath, a.handleGetDescription)
	exthttp.RegisterHttpHandler(a.description.Start.Path, a.handleStart)
	exthttp.RegisterHttpHandler(a.description.Status.Path, a.handleStatus)
	if a.hasCancel() {
		exthttp.RegisterHttpHandler(a.description.Cancel.Path, a.handleCancel)
	}
}

// getDescriptionWithDefaults wraps the preflight description and adds default paths and methods for prepare, start, status, cancel and metrics.
func getDescriptionWithDefaults(preflight Preflight) preflight_kit_api.PreflightDescription {
	description := preflight.Describe()
	if description.Start.Path == "" {
		description.Start.Path = fmt.Sprintf("/%s/start", description.Id)
	}
	if description.Start.Method == "" {
		description.Start.Method = preflight_kit_api.POST
	}
	if _, ok := preflight.(PreflightWithCancel); ok && description.Cancel == nil {
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
