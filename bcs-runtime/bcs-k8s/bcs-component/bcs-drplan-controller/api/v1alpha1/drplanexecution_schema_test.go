package v1alpha1

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDRPlanExecutionCRD_AllowsCanceledNestedStatuses(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve caller")
	}

	root := filepath.Join(filepath.Dir(filename), "..", "..")
	crdPath := filepath.Join(root, "config", "crd", "bases", "dr.bkbcs.tencent.com_drplanexecutions.yaml")
	content, err := os.ReadFile(filepath.Clean(crdPath))
	if err != nil {
		t.Fatalf("read CRD: %v", err)
	}

	crd := string(content)
	expectContains(t, crd, "Phase is the stage phase: Pending, Running, Succeeded,")
	expectContains(t, crd, "Phase is the workflow phase: Pending, Running,")
	expectContains(t, crd, "Phase is the action phase: Pending,")
	if got := strings.Count(crd, "- Canceled"); got < 4 {
		t.Fatalf("expected nested execution status enums to include Canceled, found %d occurrences", got)
	}
}

func expectContains(t *testing.T, content, want string) {
	t.Helper()
	if !strings.Contains(content, want) {
		t.Fatalf("expected CRD to contain %q", want)
	}
}
