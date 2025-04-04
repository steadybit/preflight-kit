# How to Write a Preflight Action

Preflight actions provide a way to validate conditions before an experiment is executed, ensuring experiments only run when it's safe and appropriate to do so. This guide walks you through creating a preflight action using the Steadybit PreflightKit.

## Understanding Preflight Actions

Before diving into implementation details, let's understand what a preflight action does:

1. **Validation**: Preflights check conditions that must be met for an experiment to run.
2. **Prevention**: They can prevent experiments from running when conditions aren't appropriate.
3. **Integration**: They integrate with external systems to gather information for decision-making.

Common use cases for preflight actions include:

- Ensuring experiments only run during appropriate time windows
- Verifying system health before experiment execution
- Enforcing approval workflows
- Limiting the rate of experiments

## Getting Started

To create a preflight action, you'll need to:

1. Set up an extension using the SDK
2. Register one or more preflight actions
3. Implement the check logic
4. Deploy the extension and register it with Steadybit agents

## Step 1: Set Up the Extension

Clone our [Scaffold extension](https://github.com/steadybit/extension-scaffold) or build your own.

Add to the main.go file:

```go

// register your preflight action
preflight_kit_sdk.RegisterPreflight(extpreflight.NewSimplePreflight())

// add it to the extension list response type
type ExtensionListResponse struct {
    ...
    preflight_kit_api.PreflightList `json:",inline"`
}

func getExtensionList() ExtensionListResponse {
    return ExtensionListResponse{
        ...
        
        // See this document to learn more about the preflight list:
        // https://github.com/steadybit/preflight-kit/main/docs/preflight-api.md#index-response
        PreflightList: preflight_kit_sdk.GetPreflightList(),
    }
}
```

Create a new file `preflight.go` and implement your preflight action:

See our [example](examples.md) for a preflight implementation.



## Step 2: Register with Steadybit Agents

To register your extension with Steadybit agents, you need to tell the agents where to find it. Follow the [Extension Registration](../preflight-registration.md) documentation to register your extension.

## Step 3: Use the Preflight Action in Experiments

Once registered, your preflight action will be available in the Steadybit platform. To use it:

1. Go to the Extension Settings section: https://platform.steadybit.com/settings/extensions;tab=preflightDefinitions
2. Ensure your extension is listed and enabled.
3. Go to the Integration Settings section: https://platform.steadybit.com/settings/integrations/preflightAction
4. Click on Add Preflight Action Integration.
5. Select your preflight action from the list and configure it as needed.
Now, when the experiment is executed, Steadybit will first run your preflight action to determine if the experiment should proceed.

## Advanced Preflight Actions

### Integrating with External Systems

In real-world scenarios, you might want to integrate with external systems like:

- Maintenance window schedulers
- Change management systems
- Monitoring tools
- CI/CD pipelines

To do this, your preflight action would:

1. Make API calls to these external systems
2. Process the response
3. Make a decision based on the information received

## Best Practices

1. **Make checks specific**: Each preflight action should validate one specific condition.
2. **Clear messaging**: Provide clear, actionable messages when a check fails.
3. **Error handling**: Handle errors gracefully and provide meaningful error messages.
4. **Performance**: Keep checks lightweight and fast - they should respond within a few seconds.
5. **Logging**: Include sufficient logging to diagnose issues.
6. **Security**: Be mindful of security implications, especially when integrating with external systems.
7. **Configurability**: Make checks configurable to accommodate different environments and use cases.

## Conclusion

Preflight actions are a powerful way to ensure experiments run only under the right conditions. By following this guide, you can create custom preflight actions tailored to your organization's needs, integrating with your existing systems and processes to make chaos engineering safer and more effective.

For more examples, see the [Examples](../examples.md) document.
