# Globalization Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `Globalization` as a first-class workflow action with typed schema, executor, validation, RBAC, and regression tests.

**Architecture:** Extend the DRWorkflow action model with a dedicated `GlobalizationAction` that reuses Clusternet's upstream `GlobalizationSpec`. Implement a dedicated executor following the existing `Localization` and `Subscription` patterns, then wire it into webhook validation/defaulting and generated manifests.

**Tech Stack:** Go, controller-runtime fake client, Kubebuilder CRD markers, Clusternet `apps/v1alpha1`

---

### Task 1: Lock Behavior With Tests

**Files:**
- Create: `internal/executor/globalization_executor_test.go`
- Create: `internal/webhook/drworkflow_webhook_test.go`

- [ ] Write failing executor tests for `Create`, `Apply`, `Patch`, and `Delete`
- [ ] Run targeted executor tests and confirm compile/runtime failure before implementation
- [ ] Write failing webhook tests for defaulting and rollback validation
- [ ] Run targeted webhook tests and confirm failure before implementation

### Task 2: Implement Globalization Action

**Files:**
- Modify: `api/v1alpha1/constants.go`
- Modify: `api/v1alpha1/common_types.go`
- Modify: `cmd/main.go`
- Create: `internal/executor/globalization_executor.go`
- Modify: `internal/webhook/drworkflow_webhook.go`

- [ ] Add `ActionTypeGlobalization`, `GlobalizationAction`, and `ActionOutputs.GlobalizationRef`
- [ ] Register the new executor in `cmd/main.go`
- [ ] Implement `Create`, `Apply`, `Patch`, `Delete`, and rollback behavior in `internal/executor/globalization_executor.go`
- [ ] Extend webhook defaulting and validation for `Globalization`

### Task 3: Regenerate and Verify

**Files:**
- Modify: `config/rbac/role.yaml`
- Modify: `install/helm/bcs-drplan-controller/templates/clusterrole.yaml`
- Modify: generated files under `api/v1alpha1`, `config/crd/bases`, and `install/helm/bcs-drplan-controller/crds`

- [ ] Update RBAC to include `apps.clusternet.io/globalizations`
- [ ] Run `make manifests` and `make generate`
- [ ] Run targeted `go test` for executor and webhook packages
- [ ] Run a broader verification command if the environment permits
