package revision

import (
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/test"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/controller/history"
)

func TestMain(m *testing.M) {
	_ = apis.AddToScheme(scheme.Scheme)
}

func TestCreateApplyRevision(t *testing.T) {
	control := NewRevisionControl()
	set := test.NewGameDeployment(1)
	set.Status.CollisionCount = new(int32)
	revision, err := control.NewRevision(set, 1, set.Status.CollisionCount)
	if err != nil {
		t.Fatal(err)
	}
	set.Spec.Template.Spec.Containers[0].Name = "foo"
	if set.Annotations == nil {
		set.Annotations = make(map[string]string)
	}
	key := "foo"
	expectedValue := "bar"
	set.Annotations[key] = expectedValue
	restoredSet, err := control.ApplyRevision(set, revision)
	if err != nil {
		t.Fatal(err)
	}
	restoredRevision, err := control.NewRevision(restoredSet, 2, restoredSet.Status.CollisionCount)
	if err != nil {
		t.Fatal(err)
	}
	if !history.EqualRevision(revision, restoredRevision) {
		t.Errorf("wanted %v got %v", string(revision.Data.Raw), string(restoredRevision.Data.Raw))
	}
	value, ok := restoredRevision.Annotations[key]
	if !ok {
		t.Errorf("missing annotation %s", key)
	}
	if value != expectedValue {
		t.Errorf("for annotation %s wanted %s got %s", key, expectedValue, value)
	}
}

func TestApplyRevision(t *testing.T) {
	control := NewRevisionControl()
	set := test.NewGameDeployment(1)
	set.Status.CollisionCount = new(int32)
	currentSet := set.DeepCopy()
	currentRevision, err := control.NewRevision(set, 1, set.Status.CollisionCount)
	if err != nil {
		t.Fatal(err)
	}

	set.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{{Name: "foo", Value: "bar"}}
	updateSet := set.DeepCopy()
	updateRevision, err := control.NewRevision(set, 2, set.Status.CollisionCount)
	if err != nil {
		t.Fatal(err)
	}

	restoredCurrentSet, err := control.ApplyRevision(set, currentRevision)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(currentSet.Spec.Template, restoredCurrentSet.Spec.Template) {
		t.Errorf("want %v got %v", currentSet.Spec.Template, restoredCurrentSet.Spec.Template)
	}

	restoredUpdateSet, err := control.ApplyRevision(set, updateRevision)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(updateSet.Spec.Template, restoredUpdateSet.Spec.Template) {
		t.Errorf("want %v got %v", updateSet.Spec.Template, restoredUpdateSet.Spec.Template)
	}
}
