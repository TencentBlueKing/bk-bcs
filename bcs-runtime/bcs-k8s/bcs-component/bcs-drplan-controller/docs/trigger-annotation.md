# ⚠️ DEPRECATED: Annotation Trigger Mechanism

**Status**: DEPRECATED - This mechanism has been removed in favor of explicit DRPlanExecution CR.

## Why was it removed?

The annotation-based trigger mechanism had several issues:
1. **Revert ambiguity**: Annotations couldn't specify which execution to revert, only the last one
2. **Idempotency complexity**: Required tracking `LastProcessedTrigger` to avoid duplicate executions  
3. **Lack of precision**: Users couldn't control exactly what to revert when multiple executions existed
4. **GitOps unfriendly**: Annotations are mutable and don't work well with declarative GitOps workflows

## How to trigger executions now

### Execute a Plan

Create a DRPlanExecution CR:

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: my-plan-exec-001
  namespace: default
spec:
  planRef: my-plan
  operationType: Execute
```

Apply it:

```bash
kubectl apply -f drplanexecution-execute.yaml
```

### Revert an Execution

**IMPORTANT**: You must explicitly specify which execution to revert using `revertExecutionRef`:

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: my-plan-revert-001
  namespace: default
spec:
  planRef: my-plan
  operationType: Revert
  revertExecutionRef: my-plan-exec-001  # Required!
```

Apply it:

```bash
kubectl apply -f drplanexecution-revert.yaml
```

## Benefits of the new approach

✅ **Explicit control**: You specify exactly which execution to revert  
✅ **GitOps friendly**: CRs are immutable and declarative  
✅ **Clear history**: Each execution is a separate CR with full status  
✅ **Better audit**: All executions are trackable Kubernetes resources  
✅ **Type safety**: Webhook validates `revertExecutionRef` before execution  

## Migration guide

If you were using annotation triggers:

**Before (DEPRECATED)**:
```bash
kubectl annotate drplan my-plan dr.bkbcs.tencent.com/trigger=execute --overwrite
kubectl annotate drplan my-plan dr.bkbcs.tencent.com/trigger=revert --overwrite
```

**After (CURRENT)**:
```bash
# Execute
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: my-plan-exec-$(date +%Y%m%d-%H%M%S)
spec:
  planRef: my-plan
  operationType: Execute
EOF

# Revert (must specify which execution)
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: my-plan-revert-$(date +%Y%m%d-%H%M%S)
spec:
  planRef: my-plan
  operationType: Revert
  revertExecutionRef: my-plan-exec-20260203-100000  # Required!
EOF
```

## See also

- [User Guide](user-guide.md) - Complete guide for using DRPlanExecution CRs
- [Quick Start](../specs/001-drplan-action-executor/quickstart.md) - Getting started guide
- [Example README](../example/README.md) - Real-world examples
