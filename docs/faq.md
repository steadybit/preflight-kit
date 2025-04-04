# FAQ

## I registered my preflight provider, but my preflight actions are not being executed before experiments?

There are two common issues causing this:

1. The agent couldn't (properly) communicate with the extension. To analyze this further, please inspect the agent log.
2. The preflight action is not properly registered. Ensure that the preflight action is correctly defined in the extension and that the agent has access to it.
3. The preflight action is not properly configured for all experiments or a specific team. Ensure that the preflight action is enabled for the specific experiment or team you are testing. See: [Preflight Integration](https://platform.steadybit.com/settings/integrations/preflightAction).

## Why is my experiment still running despite a preflight action condition not being met?

This could happen for several reasons:

1. Ensure that your preflight action's `status` field is properly set to `"fail"` and includes a descriptive message when conditions are not met.
2. Check if the agent logs for any errors related to the preflight execution.
3. Verify that the experiment is actually using the preflight action you've configured. See: [Preflight Integration](https://platform.steadybit.com/settings/integrations/preflightAction).
