# PreflightKit Go SDK

This module contains helper and interfaces which will help you to implement preflights using
the [preflight kit go api](https://github.com/steadybit/preflight-kit/tree/main/go/preflight_kit_api).

The module encapsulates the following technical aspects:

- JSON marshalling and unmarshalling of preflight inputs and outputs
- The sdk will wrap around your `describe` call and will provide some meaningful defaults for your endpoint definitions.
- An additional layer of rollback stability. The SDK will keep a copy of your preflight state in memory to be able to roll back to the previous state in case
  of connections issues.
- Automatic handling of `file` parameters. The SDK will automatically download the file, store it in a temporary directory and delete the file after the preflight
  has stopped. The `Config`-map in `preflight_kit_api.PreparePreflightRequestBody` will contain the path to the downloaded file.

## Installation

Add the following to your `go.mod` file:

```
go get github.com/steadybit/preflight-kit/go/preflight_kit_sdk
```

## Usage

1. Implement at least the `preflight_kit_sdk.Preflight` interface:
    - Examples:
        - [go/preflight_kit_sdk/example_preflight_test.go](./example_preflight_test.go)

2. Implement other interfaces if you need them:
    - `preflight_kit_sdk.PreflightWithStatus`
    - `preflight_kit_sdk.PreflightWithStop`
    - `preflight_kit_sdk.PreflightWithMetricQuery`

3. Register your preflight:
   ```go
   preflight_kit_sdk.RegisterPreflight(NewRolloutRestartPreflight())
   ```

4. Add your registered preflights to the index endpoint of your extension:
   ```go
   exthttp.RegisterHttpHandler("/preflights", exthttp.GetterAsHandler(preflight_kit_sdk.GetPreflightList))
   ```