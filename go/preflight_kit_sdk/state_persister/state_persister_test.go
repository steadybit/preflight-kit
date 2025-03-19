// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package state_persister

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type exampleState struct {
	stringValue string
	intValue    int
}

func TestInmemoryStatePersister_basics(t *testing.T) {
	persister := NewInmemoryStatePersister()
	exe1 := uuid.New()
	exe2 := uuid.New()

	err := persister.PersistState(context.Background(), &PersistedState{exe1, "preflight-1"})
	require.NoError(t, err)
	err = persister.PersistState(context.Background(), &PersistedState{exe2, "preflight-1"})
	require.NoError(t, err)

	executionIds, err := persister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	require.Len(t, executionIds, 2)

	err = persister.DeleteState(context.Background(), exe1)
	require.NoError(t, err)

	executionIds, err = persister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	require.Len(t, executionIds, 1)
}

func TestInmemoryStatePersister_should_ignore_not_found(t *testing.T) {
	persister := NewInmemoryStatePersister()
	exe1 := uuid.New()
	err := persister.PersistState(context.Background(), &PersistedState{exe1, "preflight-1"})
	require.NoError(t, err)

	err = persister.DeleteState(context.Background(), uuid.New())
	require.NoError(t, err)

	executionIds, err := persister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	require.Len(t, executionIds, 1)
}

func TestInmemoryStatePersister_should_update_existing_values(t *testing.T) {
	persister := NewInmemoryStatePersister()
	exe1 := uuid.New()
	err := persister.PersistState(context.Background(), &PersistedState{exe1, "preflight-1"})
	require.NoError(t, err)

	err = persister.PersistState(context.Background(), &PersistedState{exe1, "preflight-1"})
	require.NoError(t, err)

	executionIds, err := persister.GetExecutionIds(context.Background())
	require.NoError(t, err)
	require.Len(t, executionIds, 1)

	state, err := persister.GetState(context.Background(), executionIds[0])
	require.NoError(t, err)
	require.NotNil(t, state)
	require.Equal(t, executionIds[0], state.PreflightActionExecutionId)
	require.Equal(t, "preflight-1", state.PreflightActionId)
}
