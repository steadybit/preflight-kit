# Changelog

## 2.1.1

- fix: prevent data races and panics in the preflight stop/heartbeat handling — guard the shared `stopEvents` slice with a mutex, make `heartbeat.Monitor.Stop` idempotent, and make `RecordHeartbeat` a non-blocking, closed-safe send, so concurrent stop/status/timeout paths can no longer crash the extension (double-close / send-on-closed-channel / slice race)
- fix: stop and replace an existing heartbeat monitor when a preflight is started again for the same execution id, instead of leaking the previous monitor's goroutines
- refactor: use the shared `extheartbeat` watchdog from extension-kit instead of the local `heartbeat` package (removed); requires extension-kit v1.10.7

## 2.0.2

- Update dependencies

## 2.0.1

- Added missing request to start preflight action method
- Fixed state handling bug:
    - when
        - an action was returning an error directly (not as part of the `*Result` types)
        - from the `start` and `status` methods
        - and the action had modified the state before
    - then
        - the modified state was not passed to the `cancel` method

## 2.0.0

- Added state handling for preflight actions (requires agent >= 2.2.1)

## 1.0.1

- Update dependencies

## 1.0.0

- Initial release

