# Examples

This document provides examples of preflight checks that can be implemented using PreflightKit.

## Example Preflight Check Structure

Preflight checks are typically implemented as Go structs that satisfy the preflight interface. Here's a basic example of a preflight check that demonstrates the structure:

```go
package extpreflight

import (
	"context"
	"errors"
	"github.com/google/uuid"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"sync"
)

// MaintenanceWindowPreflight checks if experiments run within allowed time windows
type MaintenanceWindowPreflight struct {
	// You can add fields here for configuration, etc.
	runningPreflights sync.Map // Used to track running preflight checks
	statusCounts      sync.Map // Used to count status calls
}

// NewMaintenanceWindowPreflight creates a new maintenance window preflight
func NewMaintenanceWindowPreflight() *MaintenanceWindowPreflight {
	return &MaintenanceWindowPreflight{}
}

// Describe returns the preflight description
func (p *MaintenanceWindowPreflight) Describe() preflight_kit_api.PreflightDescription {
	return preflight_kit_api.PreflightDescription{
		Id:          "com.example.preflights.maintenance-window",
		Version:     "v0.1.0",
		Label:       "Maintenance Window Check",
		Description: "Ensures experiments only run during specified maintenance windows",
		Icon: extutil.Ptr("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZD0iTTEyIDIxQzE2Ljk3MDYgMjEgMjEgMTYuOTcwNiAyMSAxMkMyMSA3LjAyOTQ0IDE2Ljk3MDYgMyAxMiAzQzcuMDI5NDQgMyAzIDcuMDI5NDQgMyAxMkMzIDE2Ljk3MDYgNy4wMjk0NCAyMSAxMiAyMVoiIHN0cm9rZT0iIzE4MTgxOCIgc3Ryb2tlLXdpZHRoPSIyIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiLz4KPHBhdGggZD0iTTEyIDYuNzVWMTJINi43NjgwMSIgc3Ryb2tlPSIjMTgxODE4IiBzdHJva2Utd2lkdGg9IjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCIvPgo8L3N2Zz4K"),
		TargetAttributeIncludes: []string{"host.hostname", "k8s.deployment"},
		// Define endpoint references for the preflight lifecycle
		Start: preflight_kit_api.MutatingEndpointReference{},
		Status: preflight_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1s"), // Status will be checked every 1 second
		},
		Cancel: &preflight_kit_api.MutatingEndpointReference{},
	}
}

// Start initiates the preflight check
func (p *MaintenanceWindowPreflight) Start(_ context.Context, request preflight_kit_api.StartPreflightRequestBody) (*preflight_kit_api.StartResult, error) {
	// Store the experiment execution details for later use
	p.runningPreflights.Store(request.PreflightActionExecutionId, request.ExperimentExecution)

	// Example of how to return an error if needed
	if request.ExperimentExecution.Name != nil && *request.ExperimentExecution.Name == "TechnicalError" {
		return nil, extutil.Ptr(extension_kit.ToError("Technical error during preflight start", errors.New("detailed error info")))
	}

	// Example of returning a failure
	if request.ExperimentExecution.Name != nil && *request.ExperimentExecution.Name == "StartFailure" {
		return &preflight_kit_api.StartResult{
			Error: extutil.Ptr(preflight_kit_api.PreflightKitError{
				Title:  "Preflight start failure",
				Status: extutil.Ptr(preflight_kit_api.Failed),
			}),
		}, nil
	}

	// Normal successful start
	return &preflight_kit_api.StartResult{}, nil
}

// Status checks the current status of the preflight
func (p *MaintenanceWindowPreflight) Status(_ context.Context, request preflight_kit_api.StatusPreflightRequestBody) (*preflight_kit_api.StatusResult, error) {
	// Increment the status counter for this preflight
	count := p.incrementStatusCounter(request.PreflightActionExecutionId)

	// Return not completed for the first few calls to simulate a longer-running check
	if count < 2 {
		return &preflight_kit_api.StatusResult{Completed: false}, nil
	}

	// Get the stored experiment execution
	executionObj, ok := p.runningPreflights.Load(request.PreflightActionExecutionId)
	if !ok {
		return nil, extutil.Ptr(extension_kit.ToError("Preflight not found", errors.New("no preflight with given ID")))
	}

	execution := executionObj.(preflight_kit_api.ExperimentExecutionAO)

	// Example check logic - in a real implementation, you would check
	// if the current time is within the allowed maintenance window
	if execution.Name != nil && *execution.Name == "OutsideMaintenanceWindow" {
		return &preflight_kit_api.StatusResult{
			Completed: true,
			Error: &preflight_kit_api.PreflightKitError{
				Title:  "Outside maintenance window",
				Detail: extutil.Ptr("Experiment is scheduled outside the allowed maintenance window (8 PM to 6 AM UTC)"),
				Status: extutil.Ptr(preflight_kit_api.Failed),
			},
		}, nil
	}

	// If all checks pass, return completed with no error
	return &preflight_kit_api.StatusResult{Completed: true}, nil
}

// Cancel stops the preflight check
func (p *MaintenanceWindowPreflight) Cancel(_ context.Context, request preflight_kit_api.CancelPreflightRequestBody) (*preflight_kit_api.CancelResult, error) {
	// Clean up any resources associated with this preflight
	p.runningPreflights.Delete(request.PreflightActionExecutionId)
	p.statusCounts.Delete(request.PreflightActionExecutionId)

	return &preflight_kit_api.CancelResult{}, nil
}

// incrementStatusCounter is a helper function to track status call counts
func (p *MaintenanceWindowPreflight) incrementStatusCounter(id uuid.UUID) int {
	current, _ := p.statusCounts.LoadOrStore(id, 0)
	count := current.(int) + 1
	p.statusCounts.Store(id, count)
	return count
}

```



These example demonstrate basic patterns for implementing preflight checks. Depending on your organization's needs, you can create more sophisticated checks that integrate with various systems to ensure experiments only run when conditions are appropriate.
