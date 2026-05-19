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
	"strings"
	"testing"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestHasDependsOn(t *testing.T) {
	t.Run("all empty", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"}, {Name: "b"},
		}
		if hasDependsOn(actions) {
			t.Error("expected false when no DependsOn is set")
		}
	})

	t.Run("one has depends", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"},
			{Name: "b", DependsOn: []string{"a"}},
		}
		if !hasDependsOn(actions) {
			t.Error("expected true when DependsOn is present")
		}
	})
}

func TestBuildActionGraph(t *testing.T) {
	t.Run("valid graph", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"},
			{Name: "b", DependsOn: []string{"a"}},
			{Name: "c", DependsOn: []string{"a"}},
			{Name: "d", DependsOn: []string{"b", "c"}},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inDegree["a"] != 0 {
			t.Errorf("expected inDegree[a]=0, got %d", inDegree["a"])
		}
		if inDegree["b"] != 1 {
			t.Errorf("expected inDegree[b]=1, got %d", inDegree["b"])
		}
		if inDegree["d"] != 2 {
			t.Errorf("expected inDegree[d]=2, got %d", inDegree["d"])
		}
		if len(forward["a"]) != 2 {
			t.Errorf("expected forward[a] length 2, got %d", len(forward["a"]))
		}
	})

	t.Run("unknown reference", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a", DependsOn: []string{"nonexistent"}},
		}
		_, _, err := buildActionGraph(actions)
		if err == nil || !strings.Contains(err.Error(), "unknown action") {
			t.Errorf("expected 'unknown action' error, got: %v", err)
		}
	})

	t.Run("self-dependency", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a", DependsOn: []string{"a"}},
		}
		_, _, err := buildActionGraph(actions)
		if err == nil || !strings.Contains(err.Error(), "depends on itself") {
			t.Errorf("expected self-dependency error, got: %v", err)
		}
	})

	t.Run("no depends", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"}, {Name: "b"}, {Name: "c"},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, name := range []string{"a", "b", "c"} {
			if inDegree[name] != 0 {
				t.Errorf("expected inDegree[%s]=0, got %d", name, inDegree[name])
			}
		}
		if len(forward) != 0 {
			t.Errorf("expected empty forward map, got %d entries", len(forward))
		}
	})
}

func TestTopoSortLayers(t *testing.T) {
	t.Run("diamond DAG", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"},
			{Name: "b", DependsOn: []string{"a"}},
			{Name: "c", DependsOn: []string{"a"}},
			{Name: "d", DependsOn: []string{"b", "c"}},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("buildActionGraph: %v", err)
		}
		layers, err := topoSortLayers(actions, forward, inDegree)
		if err != nil {
			t.Fatalf("topoSortLayers: %v", err)
		}
		if len(layers) != 3 {
			t.Fatalf("expected 3 layers, got %d", len(layers))
		}
		if layers[0][0].Name != "a" {
			t.Errorf("layer 0: expected [a], got %v", layerNames(layers[0]))
		}
		if len(layers[1]) != 2 || layers[1][0].Name != "b" || layers[1][1].Name != "c" {
			t.Errorf("layer 1: expected [b,c], got %v", layerNames(layers[1]))
		}
		if layers[2][0].Name != "d" {
			t.Errorf("layer 2: expected [d], got %v", layerNames(layers[2]))
		}
	})

	t.Run("all independent", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"}, {Name: "b"}, {Name: "c"},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("buildActionGraph: %v", err)
		}
		layers, err := topoSortLayers(actions, forward, inDegree)
		if err != nil {
			t.Fatalf("topoSortLayers: %v", err)
		}
		if len(layers) != 1 {
			t.Fatalf("expected 1 layer (all parallel), got %d", len(layers))
		}
		if len(layers[0]) != 3 {
			t.Errorf("expected 3 actions in layer 0, got %d", len(layers[0]))
		}
	})

	t.Run("linear chain", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a"},
			{Name: "b", DependsOn: []string{"a"}},
			{Name: "c", DependsOn: []string{"b"}},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("buildActionGraph: %v", err)
		}
		layers, err := topoSortLayers(actions, forward, inDegree)
		if err != nil {
			t.Fatalf("topoSortLayers: %v", err)
		}
		if len(layers) != 3 {
			t.Fatalf("expected 3 layers (serial), got %d", len(layers))
		}
		for i, expected := range []string{"a", "b", "c"} {
			if len(layers[i]) != 1 || layers[i][0].Name != expected {
				t.Errorf("layer %d: expected [%s], got %v", i, expected, layerNames(layers[i]))
			}
		}
	})

	t.Run("cycle detection", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a", DependsOn: []string{"c"}},
			{Name: "b", DependsOn: []string{"a"}},
			{Name: "c", DependsOn: []string{"b"}},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("buildActionGraph: %v", err)
		}
		_, err = topoSortLayers(actions, forward, inDegree)
		if err == nil || !strings.Contains(err.Error(), "cycle detected") {
			t.Errorf("expected cycle error, got: %v", err)
		}
	})

	t.Run("preserves original order within layer", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "main"},
			{Name: "z-post", DependsOn: []string{"main"}},
			{Name: "a-post", DependsOn: []string{"main"}},
		}
		forward, inDegree, err := buildActionGraph(actions)
		if err != nil {
			t.Fatalf("buildActionGraph: %v", err)
		}
		layers, err := topoSortLayers(actions, forward, inDegree)
		if err != nil {
			t.Fatalf("topoSortLayers: %v", err)
		}
		if len(layers) != 2 {
			t.Fatalf("expected 2 layers, got %d", len(layers))
		}
		// z-post is defined before a-post, so it should come first despite alphabetical order
		if layers[1][0].Name != "z-post" || layers[1][1].Name != "a-post" {
			t.Errorf("layer 1: expected [z-post, a-post], got %v", layerNames(layers[1]))
		}
	})
}

func TestBuildActionGraph_DuplicateName(t *testing.T) {
	actions := []drv1alpha1.Action{
		{Name: "a"},
		{Name: "b"},
		{Name: "a"},
	}
	_, _, err := buildActionGraph(actions)
	if err == nil || !strings.Contains(err.Error(), "duplicate action name") {
		t.Errorf("expected duplicate name error, got: %v", err)
	}
}

func layerNames(actions []drv1alpha1.Action) []string {
	names := make([]string, len(actions))
	for i := range actions {
		names[i] = actions[i].Name
	}
	return names
}
