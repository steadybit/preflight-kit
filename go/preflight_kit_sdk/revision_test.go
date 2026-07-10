// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package preflight_kit_sdk

import (
	"testing"

	"github.com/steadybit/extension-kit/exthttp"
	"github.com/stretchr/testify/assert"
)

func TestRegisterPreflightBumpsRevision(t *testing.T) {
	ClearRegisteredPreflights()
	t.Cleanup(ClearRegisteredPreflights)

	before := exthttp.Revision()
	RegisterPreflight[ExampleState](NewExamplePreflight(make(chan Call, 10)))
	assert.NotEqual(t, before, exthttp.Revision(), "RegisterPreflight must bump the index revision")
}

func TestClearRegisteredPreflightsBumpsRevision(t *testing.T) {
	before := exthttp.Revision()
	ClearRegisteredPreflights()
	assert.NotEqual(t, before, exthttp.Revision(), "ClearRegisteredPreflights must bump the index revision")
}
