// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package preflight_kit_sdk

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"github.com/steadybit/preflight-kit/go/preflight_kit_sdk/heartbeat"
	"github.com/steadybit/preflight-kit/go/preflight_kit_sdk/state_persister"
	"net/http"
	"reflect"
	"runtime/coverage"
	"sync"
	"time"
)

var (
	registeredPreflights = make(map[string]interface{})
	statePersister       = state_persister.NewInmemoryStatePersister()
	stopEvents           = make([]stopEvent, 0, 10)
	heartbeatMonitors    = sync.Map{}
)

type stopEvent struct {
	timestamp                  time.Time
	reason                     string
	preflightActionExecutionId uuid.UUID
}

type Preflight interface {
	// Describe returns the preflight description.
	Describe() preflight_kit_api.PreflightDescription
	// Start is called when the preflight should actually happen.
	// [Details](https://github.com/steadybit/preflight-kit/blob/main/docs/preflight-api.md#start)
	Start(ctx context.Context, request preflight_kit_api.StartPreflightRequestBody) (*preflight_kit_api.StartResult, error)
	// Status is used to observe the current status of the preflight. This is called periodically by the preflight-kit if time control [preflight_kit_api.TimeControlInternal] or [preflight_kit_api.TimeControlExternal] is used.
	// [Details](https://github.com/steadybit/preflight-kit/blob/main/docs/preflight-api.md#status)
	Status(ctx context.Context, request preflight_kit_api.StatusPreflightRequestBody) (*preflight_kit_api.StatusResult, error)
}
type PreflightWithCancel interface {
	Preflight
	// Cancel is used to clean up any leftovers. This method is optional.
	// [Details](https://github.com/steadybit/preflight-kit/blob/main/docs/preflight-api.md#cancel)
	Cancel(ctx context.Context, request preflight_kit_api.CancelPreflightRequestBody) (*preflight_kit_api.CancelResult, error)
}

func CancelAllActivePreflights(reason string) {
	ctx := context.Background()
	preflightActionExecutionIds, err := statePersister.GetExecutionIds(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load active preflights")
	}
	if len(preflightActionExecutionIds) > 0 {
		log.Warn().Str("reason", reason).Msg("canceling active preflights")
	}
	for _, preflightActionExecutionId := range preflightActionExecutionIds {
		CancelPreflight(ctx, preflightActionExecutionId, reason)
	}
}

func CancelPreflight(ctx context.Context, preflightActionExecutionId uuid.UUID, reason string) {
	persistedState, err := statePersister.GetState(ctx, preflightActionExecutionId)
	if err != nil {
		log.Error().
			Err(err).
			Str("preflightActionExecutionId", preflightActionExecutionId.String()).
			Str("reason", reason).
			Msgf("state cannot be loaded, cannot cancel active preflight")
		return
	}

	action, ok := registeredPreflights[persistedState.PreflightActionId]
	if !ok {
		log.Error().
			Str("preflightActionId", persistedState.PreflightActionId).
			Str("preflightActionExecutionId", persistedState.PreflightActionExecutionId.String()).
			Str("reason", reason).
			Msgf("preflight is not registered, cannot cancel active preflight")
		return
	}

	preflightType := reflect.ValueOf(action)
	if cancelMethod := preflightType.MethodByName("Cancel"); !cancelMethod.IsNil() {
		log.Info().
			Str("preflightActionId", persistedState.PreflightActionId).
			Str("preflightActionExecutionId", preflightActionExecutionId.String()).
			Str("reason", reason).
			Msg("cancelling active preflight")

		markAsStopped(preflightActionExecutionId, reason)

		if err := cancelMethod.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(preflight_kit_api.CancelPreflightRequestBody{PreflightActionExecutionId: preflightActionExecutionId})})[1].Interface(); err != nil {
			log.Warn().
				Str("preflightActionId", persistedState.PreflightActionId).
				Str("preflightActionExecutionId", preflightActionExecutionId.String()).
				Str("reason", reason).
				Err(err.(error)).
				Msg("failed cancelling active preflight")
			return
		}

		stopMonitorHeartbeat(persistedState.PreflightActionExecutionId)
		if err := statePersister.DeleteState(ctx, persistedState.PreflightActionExecutionId); err != nil {
			log.Debug().
				Str("preflightActionId", persistedState.PreflightActionId).
				Str("preflightActionExecutionId", persistedState.PreflightActionExecutionId.String()).
				Str("reason", reason).
				Err(err).
				Msg("failed deleting persisted state")
		}
	}
}

// RegisterCoverageEndpoints registers two endpoints which get called by preflight_kit_test to retrieve coverage data.
func RegisterCoverageEndpoints() {
	exthttp.RegisterHttpHandler("/coverage/meta", handleCoverageMeta)
	exthttp.RegisterHttpHandler("/coverage/counters", handleCoverageCounters)
}

func handleCoverageMeta(w http.ResponseWriter, _ *http.Request, _ []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(200)
	if err := coverage.WriteMeta(w); err != nil {
		log.Err(err).Msgf("Failed to write coverage meta data.")
	}
}

func handleCoverageCounters(w http.ResponseWriter, _ *http.Request, _ []byte) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(200)
	if err := coverage.WriteCounters(w); err != nil {
		log.Err(err).Msgf("Failed to write coverage counters data.")
	}
}

func RegisterPreflight(a Preflight) {
	adapter := newPreflightHttpAdapter(a)
	registeredPreflights[adapter.description.Id] = a
	adapter.registerHandlers()
}

// ClearRegisteredPreflights clears all registered preflights - used for testing. Warning: This will not remove the registered routes from the http server.
func ClearRegisteredPreflights() {
	registeredPreflights = make(map[string]interface{})
}

// GetPreflightList returns a list of all root endpoints of registered preflights.
func GetPreflightList() preflight_kit_api.PreflightList {
	var result []preflight_kit_api.DescribingEndpointReference
	for preflightId := range registeredPreflights {
		result = append(result, preflight_kit_api.DescribingEndpointReference{
			Method: preflight_kit_api.GET,
			Path:   fmt.Sprintf("/%s", preflightId),
		})
	}

	return preflight_kit_api.PreflightList{
		Preflights: result,
	}
}

func monitorHeartbeat(preflightActionExecutionId uuid.UUID, interval, timeout time.Duration) {
	monitorHeartbeatWithCallback(preflightActionExecutionId, interval, timeout, func() {
		CancelPreflight(context.Background(), preflightActionExecutionId, "heartbeat timeout")
	})
}

func monitorHeartbeatWithCallback(preflightActionExecutionId uuid.UUID, interval, timeout time.Duration, callback func()) {
	// Add some jitter to the interval to account for network latency and processing time,
	// as we observed heartbeats always narrowly missing the specified interval.
	extendedInterval := interval + min(interval/100*5, 500*time.Millisecond)
	ch := make(chan time.Time, 1)
	monitor := heartbeat.Notify(ch, extendedInterval, timeout)
	heartbeatMonitors.Store(preflightActionExecutionId, monitor)
	go func() {
		for range ch {
			callback()
		}
	}()
}

func recordHeartbeat(preflightActionExecutionId uuid.UUID) {
	monitor, _ := heartbeatMonitors.Load(preflightActionExecutionId)
	if monitor != nil {
		monitor.(*heartbeat.Monitor).RecordHeartbeat()
	}
}

func stopMonitorHeartbeat(preflightActionExecutionId uuid.UUID) {
	monitor, _ := heartbeatMonitors.Load(preflightActionExecutionId)
	if monitor != nil {
		monitor.(*heartbeat.Monitor).Stop()
		heartbeatMonitors.Delete(preflightActionExecutionId)
	}
}

func markAsStopped(preflightActionExecutionId uuid.UUID, reason string) {
	if len(stopEvents) > 100 {
		stopEvents = stopEvents[1:]
	}
	stopEvents = append(stopEvents, stopEvent{
		preflightActionExecutionId: preflightActionExecutionId,
		reason:                     reason,
		timestamp:                  time.Now(),
	})
}

func getStopEvent(preflightActionExecutionId uuid.UUID) *stopEvent {
	for _, event := range stopEvents {
		if event.preflightActionExecutionId == preflightActionExecutionId {
			return &event
		}
	}
	return nil
}
