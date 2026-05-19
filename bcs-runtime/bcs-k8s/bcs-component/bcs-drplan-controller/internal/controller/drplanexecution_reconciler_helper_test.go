package controller

import (
	"testing"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// NOCC:tosa/fn_length(设计如此)
func TestValidatePlanReadyAllowsExecuteWhenPlanExecuted(t *testing.T) {
	r := &DRPlanExecutionReconciler{}

	plan := &drv1alpha1.DRPlan{
		Status: drv1alpha1.DRPlanStatus{Phase: drv1alpha1.PlanPhaseExecuted},
	}

	if err := r.validatePlanReady(plan); err != nil {
		t.Fatalf("expected Executed plan to allow Execute, got error: %v", err)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestValidatePlanReadyRejectsInvalidPlanPhase(t *testing.T) {
	r := &DRPlanExecutionReconciler{}

	plan := &drv1alpha1.DRPlan{
		Status: drv1alpha1.DRPlanStatus{Phase: drv1alpha1.PlanPhaseInvalid},
	}

	if err := r.validatePlanReady(plan); err == nil {
		t.Fatal("expected Invalid plan phase to be rejected")
	}
}
