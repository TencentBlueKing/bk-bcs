## ADDED Requirements

### Requirement: Subscription action can wait for child resources ready

#### Scenario: waitReady is not set (default behavior)
- **WHEN** a `Subscription` action is executed without `waitReady` (or `waitReady=false`)
- **THEN** the action is marked `Succeeded` after the Subscription CR is created successfully

#### Scenario: waitReady waits for scheduling first
- **WHEN** a `Subscription` action is executed with `waitReady=true`
- **THEN** the executor MUST wait until `Subscription.status.bindingClusters` becomes non-empty (within action timeout)

#### Scenario: waitReady waits for all feeds ready on all binding clusters
- **WHEN** `waitReady=true` and `Subscription.status.bindingClusters` contains \(N\) clusters
- **THEN** the executor MUST verify every feed referenced in `subscription.spec.feeds` is ready on every binding cluster

#### Scenario: readiness checks by kind
- **WHEN** checking a feed resource in a child cluster
- **THEN** readiness MUST be determined by resource kind:
  - **Deployment**: available replicas and updated replicas meet desired replicas
  - **StatefulSet**: ready replicas meet desired replicas
  - **DaemonSet**: numberReady meets desiredNumberScheduled
  - **Job**: Complete=True succeeds; Failed=True fails
  - **Others**: resource existence is sufficient

#### Scenario: missing resource results in non-ready until timeout
- **WHEN** a feed resource is not found in a child cluster
- **THEN** the executor MUST treat it as not ready and continue polling until ready or timeout

#### Scenario: timeout
- **WHEN** `waitReady=true` but readiness is not achieved within `action.timeout`
- **THEN** the action MUST fail with a timeout error message that includes Subscription namespace/name

### Requirement: drplan-gen enables waitReady for hook-generated Subscription actions

#### Scenario: hook actions in unified workflow default to waitReady true
- **WHEN** `drplan-gen` generates hook-related `DRWorkflow` actions of type `Subscription`
- **THEN** it MUST set `waitReady: true` on those actions by default

#### Scenario: main resource action does not set waitReady by default
- **WHEN** `drplan-gen` generates the main-resource `Subscription` action in the unified workflow
- **THEN** it MUST NOT set `waitReady` by default

#### Scenario: drplan-gen generates a single workflow by default
- **WHEN** `drplan-gen` generates plan/workflow YAML from rendered resources
- **THEN** it MUST generate a single stage and a single workflow by default
- **AND** hook actions and main-resource action MUST be emitted into the same generated workflow in the defined order

### Requirement: execution mode can drive action-level when filtering

#### Scenario: mode is provided and when matches
- **WHEN** `DRPlanExecution.spec.mode=Install` and action has `when: mode == "install"`
- **THEN** the action MUST be executed normally

#### Scenario: mode is provided and when does not match
- **WHEN** `DRPlanExecution.spec.mode=Upgrade` and action has `when: mode == "install"`
- **THEN** the action MUST be marked `Skipped` and workflow execution continues

#### Scenario: mode is not provided (compatibility)
- **WHEN** action has non-empty `when` but `DRPlanExecution.spec.mode` is empty
- **THEN** executor MUST keep backward-compatible behavior and execute the action (do not filter by when)

#### Scenario: unsupported when expression
- **WHEN** action uses unsupported condition syntax (non `mode == "install|upgrade"` single condition)
- **THEN** executor MUST fail the action with a clear validation error message

### Requirement: multiple Helm hook values are split

#### Scenario: one resource has multiple hook values
- **WHEN** rendered YAML contains `helm.sh/hook: pre-install,pre-upgrade`
- **THEN** classification MUST produce two hook entries for the same resource (pre-install and pre-upgrade)

