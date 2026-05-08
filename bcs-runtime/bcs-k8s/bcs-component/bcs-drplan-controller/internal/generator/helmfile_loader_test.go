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

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLoadHelmfileRelease_NormalizesSupportedHooks(t *testing.T) {
	fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "hooks", "helmfile.yaml.gotmpl")

	release, err := LoadHelmfileRelease(HelmfileLoadInput{
		File:      fixture,
		Selectors: []string{"name=demo-app"},
		ChartRepo: "oci://registry.example.com/charts",
	})
	if err != nil {
		t.Fatalf("LoadHelmfileRelease() error = %v", err)
	}

	if len(release.Hooks) != 3 {
		t.Fatalf("expected 3 supported hooks, got %d", len(release.Hooks))
	}

	expectedEvents := []string{"preapply", "presync", "postsync"}
	expectedCommands := []string{"sleep", "sleep", "sleep"}
	expectedArgs := [][]string{{"1"}, {"1"}, {"1"}}
	for i := range expectedEvents {
		if release.Hooks[i].Event != expectedEvents[i] {
			t.Fatalf("hook[%d] event = %q, want %q", i, release.Hooks[i].Event, expectedEvents[i])
		}
		if release.Hooks[i].Command != expectedCommands[i] {
			t.Fatalf("hook[%d] command = %q, want %q", i, release.Hooks[i].Command, expectedCommands[i])
		}
		if len(release.Hooks[i].Args) != len(expectedArgs[i]) {
			t.Fatalf("hook[%d] args len = %d, want %d", i, len(release.Hooks[i].Args), len(expectedArgs[i]))
		}
		for j := range expectedArgs[i] {
			if release.Hooks[i].Args[j] != expectedArgs[i][j] {
				t.Fatalf("hook[%d] args[%d] = %q, want %q", i, j, release.Hooks[i].Args[j], expectedArgs[i][j])
			}
		}
		if release.Hooks[i].Order != i {
			t.Fatalf("hook[%d] order = %d, want %d", i, release.Hooks[i].Order, i)
		}
	}
}

func TestLoadHelmfileRelease_IgnoresUnsupportedEvents(t *testing.T) {
	fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "hooks", "helmfile.yaml.gotmpl")

	release, err := LoadHelmfileRelease(HelmfileLoadInput{
		File:      fixture,
		Selectors: []string{"name=demo-app"},
		ChartRepo: "oci://registry.example.com/charts",
	})
	if err != nil {
		t.Fatalf("LoadHelmfileRelease() error = %v", err)
	}

	for _, hook := range release.Hooks {
		if hook.Event == "cleanup" || hook.Event == "prepare" {
			t.Fatalf("unsupported hook event should be filtered out, got %q", hook.Event)
		}
	}

	presyncFound := false
	for _, hook := range release.Hooks {
		if hook.Event != "presync" {
			continue
		}
		presyncFound = true
		if hook.Command != "sleep" {
			t.Fatalf("presync command = %q, want %q", hook.Command, "sleep")
		}
		expectedArgs := []string{"1"}
		if len(hook.Args) != len(expectedArgs) {
			t.Fatalf("presync args len = %d, want %d", len(hook.Args), len(expectedArgs))
		}
		for i := range expectedArgs {
			if hook.Args[i] != expectedArgs[i] {
				t.Fatalf("presync args[%d] = %q, want %q", i, hook.Args[i], expectedArgs[i])
			}
		}
		if hook.Order != 1 {
			t.Fatalf("presync order = %d, want 1", hook.Order)
		}
	}
	if !presyncFound {
		t.Fatal("expected normalized presync hook from hook entry with unsupported event")
	}
}

func TestLoadHelmfileRelease_RendersHookTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	chartPath := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "charts", "nginx-0.1.1.tgz")
	helmfilePath := filepath.Join(tmpDir, "helmfile.yaml")
	helmfileContent := []byte(`
environments:
  default:
    values:
      - bknodeman:
          bkrepo:
            repoName: node-repo
            username: node-user
            password: node-password
---
releases:
  - name: demo-app
    namespace: default
    chart: ` + chartPath + `
    version: 0.1.1
    hooks:
      - events:
          - presync
        command: echo
        args:
          - '{{ "{{ .Values.bknodeman.bkrepo.repoName }}" }}'
          - '{{ "{{ .Values.bknodeman.bkrepo.username }}" }}'
          - '{{ "{{ .Values.bknodeman.bkrepo.password }}" }}'
          - '{{ "{{ .Release.Name }}" }}'
          - '{{ "{{ .Event.Name }}" }}'
`)
	if err := os.WriteFile(helmfilePath, helmfileContent, 0o600); err != nil {
		t.Fatalf("write helmfile fixture: %v", err)
	}

	release, err := LoadHelmfileRelease(HelmfileLoadInput{
		File:      helmfilePath,
		Selectors: []string{"name=demo-app"},
		ChartRepo: "oci://registry.example.com/charts",
	})
	if err != nil {
		t.Fatalf("LoadHelmfileRelease() error = %v", err)
	}

	if len(release.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(release.Hooks))
	}
	hook := release.Hooks[0]
	if hook.Command != "echo" {
		t.Fatalf("hook command = %q, want echo", hook.Command)
	}
	expectedArgs := []string{"node-repo", "node-user", "node-password", "demo-app", "presync"}
	if len(hook.Args) != len(expectedArgs) {
		t.Fatalf("hook args len = %d, want %d: %#v", len(hook.Args), len(expectedArgs), hook.Args)
	}
	for i := range expectedArgs {
		if hook.Args[i] != expectedArgs[i] {
			t.Fatalf("hook args[%d] = %q, want %q", i, hook.Args[i], expectedArgs[i])
		}
	}
}

var _ = Describe("Helmfile loader helpers", func() {
	It("should normalize local tgz chart name", func() {
		got := normalizeChartName("./charts/bcs-services-stack-1.2.3.tgz", "1.2.3")
		Expect(got).To(Equal("bcs-services-stack"))
	})

	It("should select exactly one release", func() {
		_, err := selectSingleRelease([]HelmfileResolvedRelease{
			{ReleaseName: "a"},
			{ReleaseName: "b"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 release"))
	})

	It("should strip repo alias when chart is repo/chart", func() {
		got := normalizeChartName("bitnami/nginx", "1.2.3")
		Expect(got).To(Equal("nginx"))
	})

	It("should preserve basename when version suffix does not match", func() {
		got := normalizeChartName("./charts/custom-bcs-stack.tgz", "1.2.3")
		Expect(got).To(Equal("custom-bcs-stack"))
	})

	It("should reject empty release selection", func() {
		_, err := selectSingleRelease(nil)
		Expect(err).To(HaveOccurred())
		Expect(strings.ToLower(err.Error())).To(ContainSubstring("expected exactly 1 release"))
	})

	It("should load one release from helmfile fixture and merge final values", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:      fixture,
			Selectors: []string{"name=nginx-demo"},
			ChartRepo: "oci://registry.example.com/charts",
			PlainHTTP: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ReleaseName).To(Equal("nginx-demo"))
		Expect(release.Namespace).To(Equal("default"))
		Expect(release.Chart).To(Equal("nginx"))
		Expect(release.ChartVersion).To(Equal("0.1.1"))
		Expect(release.TimeoutSeconds).To(Equal(int32(600)))
		Expect(release.Atomic).NotTo(BeNil())
		Expect(*release.Atomic).To(BeTrue())
		Expect(release.Wait).NotTo(BeNil())
		Expect(*release.Wait).To(BeFalse())
		Expect(release.PlainHTTP).NotTo(BeNil())
		Expect(*release.PlainHTTP).To(BeTrue())
		Expect(release.ValuesYAML).To(ContainSubstring("replicaCount: 2"))
		Expect(release.ValuesYAML).To(ContainSubstring("tag: 0.1.1"))
		Expect(release.ValuesYAML).To(ContainSubstring("commonLabels:"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("repository: nginx"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("type: ClusterIP"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("port: 80"))
	})

	It("should keep full values when keep-full-values behavior is requested", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:           fixture,
			Selectors:      []string{"name=nginx-demo"},
			ChartRepo:      "oci://registry.example.com/charts",
			KeepFullValues: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ValuesYAML).To(ContainSubstring("replicaCount: 2"))
		Expect(release.ValuesYAML).To(ContainSubstring("repository: nginx"))
		Expect(release.ValuesYAML).To(ContainSubstring("type: ClusterIP"))
		Expect(release.ValuesYAML).To(ContainSubstring("port: 80"))
	})

	It("should tolerate duplicate keys in rendered values files by keeping the last value", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "duplicate", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:      fixture,
			Selectors: []string{"name=dup-demo"},
			ChartRepo: "oci://registry.example.com/charts",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ValuesYAML).To(ContainSubstring("foo: second"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("foo: first"))
	})

	It("should reject remote chart refs when default-value diff is requested", func() {
		stateFile := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		_, err := diffValuesAgainstChartDefaults(stateFile, "oci://registry.example.com/charts/nginx", map[string]interface{}{
			"replicaCount": int64(2),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("--keep-full-values"))
	})
})
