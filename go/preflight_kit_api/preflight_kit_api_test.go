package preflight_kit_api

import (
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"testing"
	"time"
)

// markAsUsed checks that the provided value is not nil.
func markAsUsed(t *testing.T, v any) {
	if v == nil {
		t.Fail()
	}
}

func TestPreflightKitObjects(t *testing.T) {
	// AbstractExperimentExecutionStepAO
	t.Run("AbstractExperimentExecutionStepAO", func(t *testing.T) {
		now := time.Now()
		idVal := openapi_types.UUID(uuid.New())
		customLabel := "test label"
		ignoreFailure := true
		state := "running"
		ae := AbstractExperimentExecutionStepAO{
			CustomLabel:   &customLabel,
			Ended:         &now,
			Id:            &idVal,
			IgnoreFailure: &ignoreFailure,
			Parameters:    &map[string]map[string]interface{}{"key": {"subkey": "value"}},
			PredecessorId: &idVal,
			Reason:        &customLabel,
			Started:       &now,
			State:         &state,
		}
		markAsUsed(t, ae)
	})

	// AttributeAO
	t.Run("AttributeAO", func(t *testing.T) {
		attr := AttributeAO{
			Key:   "example-key",
			Value: "example-value",
		}
		markAsUsed(t, attr)
	})

	// BlastRadiusAO
	t.Run("BlastRadiusAO", func(t *testing.T) {
		maximum := int32(10)
		percentage := int32(50)
		targetType := "server"
		predicate := TargetPredicate{"region": "us-east"}
		br := BlastRadiusAO{
			Maximum:    &maximum,
			Percentage: &percentage,
			Predicate:  &predicate,
			TargetType: &targetType,
		}
		markAsUsed(t, br)
	})

	// CancelResult
	t.Run("CancelResult", func(t *testing.T) {
		detail := "error detail"
		title := "cancel error"
		status := Errored
		errObj := PreflightKitError{
			Detail: &detail,
			Title:  title,
			Status: &status,
		}
		cr := CancelResult{
			Error: &errObj,
		}
		markAsUsed(t, cr)
	})

	// DescribingEndpointReference
	t.Run("DescribingEndpointReference", func(t *testing.T) {
		der := DescribingEndpointReference{
			Method: GET,
			Path:   "/describe",
		}
		markAsUsed(t, der)
	})

	// ExperimentExecutionAO
	t.Run("ExperimentExecutionAO", func(t *testing.T) {
		now := time.Now()
		key := "exp-key-123"
		name := "Example Experiment"
		ee := ExperimentExecutionAO{
			Created: &now,
			Key:     &key,
			Name:    &name,
		}
		markAsUsed(t, ee)
	})

	// ExperimentExecutionStepActionAO
	t.Run("ExperimentExecutionStepActionAO", func(t *testing.T) {
		actionID := "action-123"
		customLabel := "step action label"
		now := time.Now()
		// Reuse BlastRadiusAO from above (with zero value is acceptable for testing)
		br := BlastRadiusAO{}
		esa := ExperimentExecutionStepActionAO{
			ActionId:         &actionID,
			ActionKind:       nil, // or set to one of the constants, e.g., ATTACK
			CustomLabel:      &customLabel,
			Ended:            &now,
			Id:               nil, // could be set as needed
			IgnoreFailure:    nil,
			Parameters:       nil,
			PredecessorId:    nil,
			Radius:           &br,
			Reason:           nil,
			Started:          &now,
			State:            nil,
			TargetExecutions: nil,
			TotalTargetCount: nil,
		}
		markAsUsed(t, esa)
	})

	// ExperimentExecutionStepWaitAO
	t.Run("ExperimentExecutionStepWaitAO", func(t *testing.T) {
		customLabel := "wait step"
		now := time.Now()
		esw := ExperimentExecutionStepWaitAO{
			CustomLabel:   &customLabel,
			Ended:         &now,
			Id:            nil,
			IgnoreFailure: nil,
			Parameters:    nil,
			PredecessorId: nil,
			Reason:        nil,
			Started:       &now,
			State:         nil,
		}
		markAsUsed(t, esw)
	})

	// ExperimentExecutionVariableAO
	t.Run("ExperimentExecutionVariableAO", func(t *testing.T) {
		origin := ExperimentExecutionVariableAOOrigin("ENVIRONMENT")
		value := "variable-value"
		ev := ExperimentExecutionVariableAO{
			Origin: &origin,
			Value:  &value,
		}
		markAsUsed(t, ev)
	})

	// MutatingEndpointReference
	t.Run("MutatingEndpointReference", func(t *testing.T) {
		mer := MutatingEndpointReference{
			Method: POST,
			Path:   "/mutate",
		}
		markAsUsed(t, mer)
	})

	// MutatingEndpointReferenceWithCallInterval
	t.Run("MutatingEndpointReferenceWithCallInterval", func(t *testing.T) {
		interval := "10s"
		merci := MutatingEndpointReferenceWithCallInterval{
			CallInterval: &interval,
			Method:       PUT,
			Path:         "/mutate-with-interval",
		}
		markAsUsed(t, merci)
	})

	// PreflightDescription
	t.Run("PreflightDescription", func(t *testing.T) {
		interval := "5s"
		start := MutatingEndpointReference{
			Method: POST,
			Path:   "/start",
		}
		status := MutatingEndpointReferenceWithCallInterval{
			CallInterval: &interval,
			Method:       PUT,
			Path:         "/status",
		}
		cancel := MutatingEndpointReference{
			Method: DELETE,
			Path:   "/cancel",
		}
		pd := PreflightDescription{
			Cancel:      &cancel,
			Description: "This is a test preflight description",
			Icon:        nil,
			Id:          "org.example.preflight",
			Label:       "Test Preflight",
			Start:       start,
			Status:      status,
			Version:     "1.0.0",
		}
		markAsUsed(t, pd)
	})

	// PreflightKitError
	t.Run("PreflightKitError", func(t *testing.T) {
		detail := "detailed error description"
		title := "error title"
		status := Failed
		pe := PreflightKitError{
			Detail: &detail,
			Title:  title,
			Status: &status,
		}
		markAsUsed(t, pe)
	})

	// PreflightList
	t.Run("PreflightList", func(t *testing.T) {
		der := DescribingEndpointReference{
			Method: GET,
			Path:   "/list",
		}
		pl := PreflightList{
			Preflights: []DescribingEndpointReference{der},
		}
		markAsUsed(t, pl)
	})

	// StartResult
	t.Run("StartResult", func(t *testing.T) {
		detail := "start error"
		title := "start error title"
		status := Failed
		errObj := PreflightKitError{
			Detail: &detail,
			Title:  title,
			Status: &status,
		}
		sr := StartResult{
			Error: &errObj,
		}
		markAsUsed(t, sr)
	})

	// StatusResult
	t.Run("StatusResult", func(t *testing.T) {
		detail := "status error"
		title := "status error title"
		status := Errored
		errObj := PreflightKitError{
			Detail: &detail,
			Title:  title,
			Status: &status,
		}
		st := StatusResult{
			Completed: false,
			Error:     &errObj,
		}
		markAsUsed(t, st)
	})

	// TargetAO
	t.Run("TargetAO", func(t *testing.T) {
		name := "target-1"
		agentHostname := "agent-1"
		attr := AttributeAO{
			Key:   "os",
			Value: "linux",
		}
		attributes := []AttributeAO{attr}
		ta := TargetAO{
			AgentHostname: &agentHostname,
			Attributes:    &attributes,
			Name:          &name,
		}
		markAsUsed(t, ta)
	})

	// UserSummaryAO
	t.Run("UserSummaryAO", func(t *testing.T) {
		email := "user@example.com"
		name := "Test User"
		pictureUrl := "http://example.com/avatar.png"
		username := "testuser"
		us := UserSummaryAO{
			Email:      &email,
			Name:       &name,
			PictureUrl: &pictureUrl,
			Username:   &username,
		}
		markAsUsed(t, us)
	})

	// CancelPreflightRequestBody
	t.Run("CancelPreflightRequestBody", func(t *testing.T) {
		id := uuid.New()
		cprb := CancelPreflightRequestBody{
			PreflightActionExecutionId: id,
		}
		markAsUsed(t, cprb)
	})

	// PreflightStatusRequestBody
	t.Run("PreflightStatusRequestBody", func(t *testing.T) {
		id := uuid.New()
		psrb := PreflightStatusRequestBody{
			PreflightActionExecutionId: id,
		}
		markAsUsed(t, psrb)
	})

	// StartPreflightRequestBody
	t.Run("StartPreflightRequestBody", func(t *testing.T) {
		now := time.Now()
		key := "exp-key-456"
		name := "Experiment 456"
		ee := ExperimentExecutionAO{
			Created: &now,
			Key:     &key,
			Name:    &name,
		}
		id := uuid.New()
		sprb := StartPreflightRequestBody{
			ExperimentExecution:        ee,
			PreflightActionExecutionId: id,
		}
		markAsUsed(t, sprb)
	})

	// Union response types (empty union data)
	t.Run("CancelPreflightResponse", func(t *testing.T) {
		cpr := CancelPreflightResponse{}
		markAsUsed(t, cpr)
	})
	t.Run("DescribePreflightResponse", func(t *testing.T) {
		dpr := DescribePreflightResponse{}
		markAsUsed(t, dpr)
	})
	t.Run("PreflightListResponse", func(t *testing.T) {
		plr := PreflightListResponse{}
		markAsUsed(t, plr)
	})
	t.Run("PreflightStatusResponse", func(t *testing.T) {
		psr := PreflightStatusResponse{}
		markAsUsed(t, psr)
	})
	t.Run("StartPreflightResponse", func(t *testing.T) {
		spr := StartPreflightResponse{}
		markAsUsed(t, spr)
	})
}
