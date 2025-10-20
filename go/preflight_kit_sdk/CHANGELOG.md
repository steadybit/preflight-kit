# Changelog

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

