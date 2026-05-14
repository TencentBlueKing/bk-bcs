/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// jobRESTMapper is a minimal RESTMapper for batch/v1 Job
type jobRESTMapper struct{}

const testJobNameMyJob5 = "my-job-5"

func (m jobRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	return &meta.RESTMapping{
		Resource: schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"},
		GroupVersionKind: schema.GroupVersionKind{
			Group: gk.Group, Version: "v1", Kind: gk.Kind,
		},
		Scope: meta.RESTScopeNamespace,
	}, nil
}

func newJobDynClient(jobs ...*batchv1.Job) *fake.FakeDynamicClient {
	sc := runtime.NewScheme()
	_ = batchv1.AddToScheme(sc)
	objs := make([]runtime.Object, len(jobs))
	for i, j := range jobs {
		objs[i] = j
	}
	return fake.NewSimpleDynamicClient(sc, objs...)
}

func makeJob(name, namespace string, createdAt time.Time) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: createdAt},
		},
	}
}

func TestResolveParams_StaticValue(t *testing.T) {
	params := []drv1alpha1.Parameter{
		{Name: "foo", Value: "bar"},
		{Name: "num", Value: "42"},
	}
	result, err := resolveParams(context.Background(), nil, nil, params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", result["foo"])
	}
	if result["num"] != "42" {
		t.Errorf("expected num=42, got %v", result["num"])
	}
}

func TestResolveParams_ValueFrom_ByName(t *testing.T) {
	job := makeJob(testJobNameMyJob5, "default", time.Now())
	dc := newJobDynClient(job)

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Namespace:  "default",
					Name:       testJobNameMyJob5,
					JSONPath:   "{.metadata.name}",
				},
			},
		},
	}

	result, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["jobName"] != testJobNameMyJob5 {
		t.Errorf("expected jobName=%s, got %v", testJobNameMyJob5, result["jobName"])
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_ValueFrom_LabelSelector_Last(t *testing.T) {
	t1 := time.Now().Add(-10 * time.Minute)
	t2 := time.Now()
	job1 := makeJob("my-job-4", "default", t1)
	job2 := makeJob(testJobNameMyJob5, "default", t2)
	job1.Labels = map[string]string{"app": "foo"}
	job2.Labels = map[string]string{"app": "foo"}
	dc := newJobDynClient(job1, job2)

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion:    "batch/v1",
					Kind:          "Job",
					Namespace:     "default",
					LabelSelector: "app=foo",
					JSONPath:      "{.metadata.name}",
					Select:        "Last",
				},
			},
		},
	}

	result, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["jobName"] != testJobNameMyJob5 {
		t.Errorf("expected jobName=%s (Last), got %v", testJobNameMyJob5, result["jobName"])
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_ValueFrom_LabelSelector_First(t *testing.T) {
	t1 := time.Now().Add(-10 * time.Minute)
	t2 := time.Now()
	job1 := makeJob("my-job-4", "default", t1)
	job2 := makeJob(testJobNameMyJob5, "default", t2)
	job1.Labels = map[string]string{"app": "foo"}
	job2.Labels = map[string]string{"app": "foo"}
	dc := newJobDynClient(job1, job2)

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion:    "batch/v1",
					Kind:          "Job",
					Namespace:     "default",
					LabelSelector: "app=foo",
					JSONPath:      "{.metadata.name}",
					Select:        "First",
				},
			},
		},
	}

	result, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["jobName"] != "my-job-4" {
		t.Errorf("expected jobName=my-job-4 (First), got %v", result["jobName"])
	}
}

// TestResolveParams_ValueFrom_Single_TooMany verifies Select=Single returns an error
// when the fake dynamic client returns multiple items (fake client ignores label filtering).
// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_ValueFrom_Single_TooMany(t *testing.T) {
	job1 := makeJob("my-job-4", "default", time.Now())
	job2 := makeJob(testJobNameMyJob5, "default", time.Now())
	dc := newJobDynClient(job1, job2)

	// No labelSelector → fake client returns both jobs, triggering the Single>1 error path.
	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Namespace:  "default",
					JSONPath:   "{.metadata.name}",
					Select:     "Single",
				},
			},
		},
	}

	_, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err == nil {
		t.Error("expected error for Single with multiple matches, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "expected single match") {
		t.Errorf("expected 'expected single match' in error, got: %v", err)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_ValueFrom_EmptyList(t *testing.T) {
	dc := newJobDynClient() // no jobs

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion:    "batch/v1",
					Kind:          "Job",
					Namespace:     "default",
					LabelSelector: "app=nonexistent",
					JSONPath:      "{.metadata.name}",
				},
			},
		},
	}

	_, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err == nil {
		t.Error("expected error for empty list, got nil")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_ValueFrom_TemplateNamespace(t *testing.T) {
	job := makeJob(testJobNameMyJob5, "blueking", time.Now())
	dc := newJobDynClient(job)

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Namespace:  "$(params.ns)",
					Name:       testJobNameMyJob5,
					JSONPath:   "{.metadata.name}",
				},
			},
		},
	}

	already := map[string]interface{}{"ns": "blueking"}
	result, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, already)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["jobName"] != testJobNameMyJob5 {
		t.Errorf("expected jobName=%s, got %v", testJobNameMyJob5, result["jobName"])
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestResolveParams_DynamicClientCalled(t *testing.T) {
	job := makeJob("target-job", "ns1", time.Now())
	dc := newJobDynClient(job)

	params := []drv1alpha1.Parameter{
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Namespace:  "ns1",
					Name:       "target-job",
					JSONPath:   "{.metadata.name}",
				},
			},
		},
	}

	_, err := resolveParams(context.Background(), dc, jobRESTMapper{}, params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	actions := dc.Actions()
	if len(actions) == 0 {
		t.Error("expected dynamic client to be called, got no actions")
	}
	var hasGetOrList bool
	for _, a := range actions {
		if a.GetVerb() == "get" || a.GetVerb() == "list" {
			hasGetOrList = true
		}
	}
	if !hasGetOrList {
		t.Errorf("expected get/list action, got: %v", actions)
	}
}
