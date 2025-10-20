// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2024 Steadybit GmbH

package preflight_kit_sdk

import (
	"context"

	"github.com/steadybit/extension-kit/extconversion"
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

type ExampleState struct {
	Foo      string
	TestStep string
}

func NewExamplePreflight(calls chan<- Call) *ExamplePreflight {
	return &ExamplePreflight{calls: calls}
}

// Make sure our ExamplePreflight implements all the interfaces we need
var _ Preflight[ExampleState] = (*ExamplePreflight)(nil)
var _ PreflightWithCancel[ExampleState] = (*ExamplePreflight)(nil)

func toExampleState(state preflight_kit_api.PreflightState) *ExampleState {
	result := ExampleState{}
	err := extconversion.Convert(state, &result)
	if err != nil {
		panic(err)
	}
	return &result
}

func (preflight *ExamplePreflight) NewEmptyState() ExampleState {
	return ExampleState{}
}

func (preflight *ExamplePreflight) Describe() preflight_kit_api.PreflightDescription {
	return preflight_kit_api.PreflightDescription{
		Id:                      "ExamplePreflightId",
		Description:             "This is an Example Preflight",
		Start:                   preflight_kit_api.MutatingEndpointReference{},
		TargetAttributeIncludes: []string{"target.attribute.to.include", "target.attribute.to.include.2"},
		Status: preflight_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1s"),
		},
		Cancel: &preflight_kit_api.MutatingEndpointReference{},
	}
}

func (preflight *ExamplePreflight) Start(_ context.Context, state *ExampleState) (*preflight_kit_api.StartResult, error) {
	preflight.calls <- Call{"Start", []interface{}{state}}
	state.TestStep = "Prepare"
	return &preflight_kit_api.StartResult{}, nil
}

func (preflight *ExamplePreflight) Status(_ context.Context, state *ExampleState) (*preflight_kit_api.StatusResult, error) {
	preflight.calls <- Call{"Status", []interface{}{state}}
	if preflight.statusError != nil {
		return nil, preflight.statusError
	}
	state.TestStep = "Status"
	return &preflight_kit_api.StatusResult{}, nil
}

func (preflight *ExamplePreflight) Cancel(_ context.Context, state *ExampleState) (*preflight_kit_api.CancelResult, error) {
	preflight.calls <- Call{"Cancel", []interface{}{state}}
	return &preflight_kit_api.CancelResult{}, nil
}
