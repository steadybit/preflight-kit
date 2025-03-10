# PreflightKit Go API

This module exposes Go types that you will find helpful when implementing an PreflightKit extension.

The types are generated automatically from the PreflightKit [OpenAPI specification](https://github.com/steadybit/preflight-kit/tree/main/openapi).

## Installation

Add the following to your `go.mod` file:

```
go get github.com/steadybit/preflight-kit/go/preflight_kit_api
```

## Usage

```go
import (
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
)

preflightList := preflight_kit_api.PreflightList{
    Preflights: []preflight_kit_api.DescribingEndpointReference{
        {
            "GET",
            "/preflights/check-experiment-start",
        },
    },
}
```