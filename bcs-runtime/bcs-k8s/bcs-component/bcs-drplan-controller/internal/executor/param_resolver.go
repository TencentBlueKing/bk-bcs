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
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/jsonpath"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/internal/utils"
)

// paramRESTMapper is a minimal interface for mapping GVK to GVR.
type paramRESTMapper interface {
	RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error)
}

// resolveParams resolves a list of Parameters (possibly with valueFrom) into a string map.
// alreadyResolved is the current context used to render namespace/name/labelSelector in manifestRef.
// The returned map contains only the resolved parameters from this call (not alreadyResolved).
func resolveParams(
	ctx context.Context,
	dc dynamic.Interface,
	mapper paramRESTMapper,
	params []drv1alpha1.Parameter,
	alreadyResolved map[string]interface{},
) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(params))
	for _, p := range params {
		val, err := resolveOneParam(ctx, dc, mapper, p, alreadyResolved)
		if err != nil {
			return nil, fmt.Errorf("resolving param %q: %w", p.Name, err)
		}
		result[p.Name] = val
	}
	return result, nil
}

func resolveOneParam(
	ctx context.Context,
	dc dynamic.Interface,
	mapper paramRESTMapper,
	p drv1alpha1.Parameter,
	resolved map[string]interface{},
) (string, error) {
	if p.ValueFrom == nil || p.ValueFrom.ManifestRef == nil {
		return p.Value, nil
	}
	if dc == nil || mapper == nil {
		return "", fmt.Errorf("param %q uses valueFrom but dynamicClient/mapper is not configured", p.Name)
	}

	ref := p.ValueFrom.ManifestRef
	templateData := &utils.TemplateData{Params: resolved}

	// Render namespace, name, labelSelector (they may contain $(params.xxx))
	namespace, err := utils.RenderTemplate(ref.Namespace, templateData)
	if err != nil {
		return "", fmt.Errorf("rendering namespace: %w", err)
	}
	name, err := utils.RenderTemplate(ref.Name, templateData)
	if err != nil {
		return "", fmt.Errorf("rendering name: %w", err)
	}
	labelSel, err := utils.RenderTemplate(ref.LabelSelector, templateData)
	if err != nil {
		return "", fmt.Errorf("rendering labelSelector: %w", err)
	}

	// Resolve GVR via mapper
	gvk := schema.FromAPIVersionAndKind(ref.APIVersion, ref.Kind)
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", fmt.Errorf("mapping GVK %s: %w", gvk, err)
	}
	gvr := mapping.Resource

	// Fetch the target object(s)
	var items []map[string]interface{}
	if name != "" {
		// Precise Get
		obj, getErr := dc.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return "", fmt.Errorf("get %s/%s/%s: %w", gvr, namespace, name, getErr)
		}
		items = []map[string]interface{}{obj.Object}
	} else {
		// List with optional label selector
		listOpts := metav1.ListOptions{}
		if labelSel != "" {
			listOpts.LabelSelector = labelSel
		}
		list, listErr := dc.Resource(gvr).Namespace(namespace).List(ctx, listOpts)
		if listErr != nil {
			return "", fmt.Errorf("list %s in %s: %w", gvr, namespace, listErr)
		}
		for i := range list.Items {
			items = append(items, list.Items[i].Object)
		}
	}

	if len(items) == 0 {
		return "", fmt.Errorf("no resource found for %s/%s/%s (selector=%q)", gvr, namespace, name, labelSel)
	}

	// Apply select strategy (only relevant when len > 1)
	selected, err := selectItem(items, ref.Select)
	if err != nil {
		return "", err
	}

	// Evaluate JSONPath
	return evalJSONPath(selected, ref.JSONPath)
}

// selectItem picks one item from a list according to the Select strategy.
func selectItem(items []map[string]interface{}, selectStrategy string) (map[string]interface{}, error) {
	if len(items) == 1 {
		return items[0], nil
	}

	switch selectStrategy {
	case "Single":
		return nil, fmt.Errorf("expected single match, got %d", len(items))
	case "First":
		sort.Slice(items, func(i, j int) bool {
			ti := creationTimestamp(items[i])
			tj := creationTimestamp(items[j])
			return ti.Before(&tj)
		})
		return items[0], nil
	case "Any":
		return items[0], nil
	default: // "Last" or empty (default is Last)
		sort.Slice(items, func(i, j int) bool {
			ti := creationTimestamp(items[i])
			tj := creationTimestamp(items[j])
			return ti.After(tj.Time)
		})
		return items[0], nil
	}
}

// creationTimestamp extracts creationTimestamp from an unstructured object.
func creationTimestamp(obj map[string]interface{}) metav1.Time {
	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return metav1.Time{}
	}
	tsStr, ok := meta["creationTimestamp"].(string)
	if !ok {
		return metav1.Time{}
	}
	t := &metav1.Time{}
	if err := t.UnmarshalQueryParameter(tsStr); err != nil {
		return metav1.Time{}
	}
	return *t
}

// evalJSONPath evaluates a JSONPath expression against an unstructured object.
func evalJSONPath(obj map[string]interface{}, expression string) (string, error) {
	j := jsonpath.New("param")
	j.AllowMissingKeys(false)
	if err := j.Parse(expression); err != nil {
		return "", fmt.Errorf("invalid jsonPath expression %q: %w", expression, err)
	}
	var buf bytes.Buffer
	if err := j.Execute(&buf, obj); err != nil {
		return "", fmt.Errorf("jsonPath execution failed for %q: %w", expression, err)
	}
	val := strings.TrimSpace(buf.String())
	if val == "" {
		return "", fmt.Errorf("jsonPath %q returned empty value", expression)
	}
	return val, nil
}
