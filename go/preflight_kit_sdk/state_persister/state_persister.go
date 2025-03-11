// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package state_persister

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type PersistedState struct {
	PreflightActionExecutionId uuid.UUID
	PreflightActionId          string
}

type StatePersister interface {
	PersistState(ctx context.Context, state *PersistedState) error
	GetExecutionIds(ctx context.Context) ([]uuid.UUID, error)
	GetState(ctx context.Context, uuid uuid.UUID) (*PersistedState, error)
	DeleteState(ctx context.Context, executionId uuid.UUID) error
}

func NewInmemoryStatePersister() StatePersister {
	return &inmemoryStatePersister{states: sync.Map{}}
}

type inmemoryStatePersister struct {
	states sync.Map // map[uuid.UUID]*PersistedState
}

func (p *inmemoryStatePersister) PersistState(_ context.Context, state *PersistedState) error {
	p.states.Store(state.PreflightActionExecutionId, state)
	return nil
}

func (p *inmemoryStatePersister) GetExecutionIds(_ context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	p.states.Range(func(key, value interface{}) bool {
		ids = append(ids, key.(uuid.UUID))
		return true
	})
	return ids, nil
}

func (p *inmemoryStatePersister) GetState(_ context.Context, uuid uuid.UUID) (*PersistedState, error) {
	state, ok := p.states.Load(uuid)
	if !ok {
		return nil, fmt.Errorf("state not found for execution id %s", uuid)
	}
	return state.(*PersistedState), nil
}

func (p *inmemoryStatePersister) DeleteState(_ context.Context, executionId uuid.UUID) error {
	p.states.Delete(executionId)
	return nil
}
