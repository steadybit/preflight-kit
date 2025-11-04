# Changelog

## 1.4.0

- Aligned the spec to the platform OpenApi spec - as a result the constants for "ExperimentExecutionStepWaitAOStepType" and "ExperimentExecutionStepActionAOStepType" has been removed.
- Support property changes via actions (requires steadybit platform >= 2.3.25 and agent >= 2.2.2)

## 1.3.0

- Added state handling for preflight actions (similar to action-kit) (requires agent >= 2.2.1)

## 1.2.0

- Renamed `TargetAO` to `TargetExecutionAO`.

## 1.1.1 (requires platform version >= 2.3.8)

- Removed `experimentProperties` and `executionProperties` in `ExperimentExecutionAO`.
- Added `properties` in `ExperimentExecutionAO`.

## 1.0.2

- Fixed `experimentProperties` and `executionProperties` in `ExperimentExecutionAO`.

## 1.0.1

- Added `experimentProperties` and `executionProperties` to `ExperimentExecutionAO`.

## 1.0.0

- Initial release

