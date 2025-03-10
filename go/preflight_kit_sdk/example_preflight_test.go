// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package preflight_kit_sdk

import (
	"context"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
)

type Call struct {
	Name string
	Args []interface{}
}

type ExamplePreflight struct {
	calls       chan<- Call
	statusError error
}
type ExampleConfig struct {
	Duration  string
	InputFile string
}

func NewExamplePreflight(calls chan<- Call) *ExamplePreflight {
	return &ExamplePreflight{calls: calls}
}

func (preflight *ExamplePreflight) Describe() preflight_kit_api.PreflightDescription {
	return preflight_kit_api.PreflightDescription{
		Id:          "ExamplePreflightId",
		Description: "This is an Example Preflight",
		Start:       preflight_kit_api.MutatingEndpointReference{},
		Status: preflight_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1s"),
		},
		Cancel: &preflight_kit_api.MutatingEndpointReference{},
	}
}

func (preflight *ExamplePreflight) Start(_ context.Context, request preflight_kit_api.StartPreflightRequestBody) (*preflight_kit_api.StartResult, error) {
	preflight.calls <- Call{"Start", []interface{}{request}}
	return &preflight_kit_api.StartResult{}, nil
}

func (preflight *ExamplePreflight) Status(_ context.Context, request preflight_kit_api.PreflightStatusRequestBody) (*preflight_kit_api.StatusResult, error) {
	preflight.calls <- Call{"Status", []interface{}{request}}
	return &preflight_kit_api.StatusResult{}, nil
}

func (preflight *ExamplePreflight) Cancel(_ context.Context, request preflight_kit_api.CancelPreflightRequestBody) (*preflight_kit_api.CancelResult, error) {
	preflight.calls <- Call{"Cancel", []interface{}{request}}
	return &preflight_kit_api.CancelResult{}, nil
}
