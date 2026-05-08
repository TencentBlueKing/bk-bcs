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
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"dario.cat/mergo"
	"github.com/helmfile/helmfile/pkg/environment"
	"github.com/helmfile/helmfile/pkg/filesystem"
	"github.com/helmfile/helmfile/pkg/helmexec"
	"github.com/helmfile/helmfile/pkg/plugins"
	"github.com/helmfile/helmfile/pkg/remote"
	"github.com/helmfile/helmfile/pkg/state"
	"github.com/helmfile/helmfile/pkg/tmpl"
	"github.com/helmfile/vals"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	sigyaml "sigs.k8s.io/yaml"
)

const defaultHelmfileEnvironment = "default"

// helmfileStateLoader is a thin adapter around exported helmfile packages.
// It intentionally reuses state.NewCreator/ExecuteTemplates/RenderReleaseValuesFileToBytes
// and only keeps the minimal glue that is unavailable because desiredStateLoader is not exported.
type helmfileStateLoader struct {
	env       string
	namespace string
	baseDir   string

	fs          *filesystem.FileSystem
	logger      *zap.SugaredLogger
	remote      *remote.Remote
	valsRuntime vals.Evaluator

	helmMu    sync.Mutex
	helmCache map[string]helmexec.Interface
}

// LoadHelmfileRelease resolves one selected helmfile release into the generator's normalized release model.
func LoadHelmfileRelease(input HelmfileLoadInput) (*HelmfileResolvedRelease, error) {
	if strings.TrimSpace(input.File) == "" {
		return nil, fmt.Errorf("helmfile path is required")
	}
	if strings.TrimSpace(input.ChartRepo) == "" {
		return nil, fmt.Errorf("chart repo is required")
	}

	absFile, err := filepath.Abs(input.File)
	if err != nil {
		return nil, fmt.Errorf("resolving helmfile path: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	logger := zap.NewNop().Sugar()
	fs := filesystem.DefaultFileSystem()
	valsRuntime, err := plugins.ValsInstance()
	if err != nil {
		return nil, fmt.Errorf("initializing vals runtime: %w", err)
	}

	loader := &helmfileStateLoader{
		env:         defaultHelmfileEnvironment,
		namespace:   strings.TrimSpace(input.Namespace),
		baseDir:     filepath.Dir(absFile),
		fs:          fs,
		logger:      logger,
		remote:      remote.NewRemote(logger, homeDir, fs),
		valsRuntime: valsRuntime,
		helmCache:   map[string]helmexec.Interface{},
	}

	st, err := loader.Load(absFile)
	if err != nil {
		return nil, err
	}

	templatedState, err := st.ExecuteTemplates()
	if err != nil {
		return nil, fmt.Errorf("executing helmfile release templates: %w", err)
	}
	st = templatedState

	st.Releases, err = st.GetReleasesWithOverrides()
	if err != nil {
		return nil, fmt.Errorf("applying helmfile release overrides: %w", err)
	}
	st.Releases = st.GetReleasesWithLabels()
	st.Selectors = append([]string(nil), input.Selectors...)
	if len(st.Selectors) > 0 {
		err = st.FilterReleases(false)
		if err != nil {
			return nil, fmt.Errorf("filtering helmfile releases: %w", err)
		}
	}

	release, err := selectSingleReleaseFromState(st)
	if err != nil {
		return nil, err
	}

	valuesYAML, err := buildReleaseValuesYAML(st, release, input.KeepFullValues)
	if err != nil {
		return nil, fmt.Errorf("building release values: %w", err)
	}

	namespace := effectiveHelmfileNamespace(release.Namespace, input.Namespace)

	hooks, err := loader.normalizeHelmfileHooks(st, release)
	if err != nil {
		return nil, fmt.Errorf("rendering release hooks: %w", err)
	}

	resolved := &HelmfileResolvedRelease{
		ReleaseName:     release.Name,
		Namespace:       namespace,
		Chart:           normalizeChartName(release.Chart, release.Version),
		ChartVersion:    strings.TrimSpace(release.Version),
		ChartRepo:       strings.TrimSpace(input.ChartRepo),
		HookImage:       strings.TrimSpace(input.HookImage),
		TargetNamespace: namespace,
		ValuesYAML:      valuesYAML,
		Hooks:           hooks,
		Wait:            effectiveReleaseWait(st, release),
		WaitForJob:      effectiveReleaseWaitForJob(st, release),
		Atomic:          effectiveReleaseAtomic(st, release),
		CreateNamespace: effectiveReleaseCreateNamespace(st, release),
		TimeoutSeconds:  effectiveReleaseTimeoutSeconds(st, release),
	}

	if plainHTTP := input.PlainHTTP || release.PlainHttp || st.HelmDefaults.PlainHttp; plainHTTP {
		resolved.PlainHTTP = boolPtrForLoader(true)
	}

	return resolved, nil
}

// Load resolves one helmfile entry point into a single merged HelmState.
// Namespace overrides are applied only after helmfile parsing so the loader
// still behaves like upstream helmfile for template and bases resolution.
func (ld *helmfileStateLoader) Load(file string) (*state.HelmState, error) {
	dir := filepath.Dir(file)
	name := filepath.Base(file)

	st, err := ld.loadFileWithEnv(nil, nil, dir, name, true)
	if err != nil {
		return nil, err
	}

	if ld.namespace != "" {
		if st.OverrideNamespace != "" {
			return nil, fmt.Errorf("cannot use namespace override and set namespace in helmfile simultaneously")
		}
		st.OverrideNamespace = ld.namespace
	}

	return st, nil
}

// loadFileWithEnv reads one helmfile document from disk and hands the bytes to
// the multi-document loader. Nested helmfiles reuse this entry point too.
func (ld *helmfileStateLoader) loadFileWithEnv(
	inheritedEnv *environment.Environment,
	overrodeEnv *environment.Environment,
	baseDir string,
	file string,
	evaluateBases bool,
) (*state.HelmState, error) {
	path := file
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, file)
	}

	fileBytes, err := ld.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading helmfile %s: %w", path, err)
	}

	return ld.loadBytes(inheritedEnv, overrodeEnv, filepath.Dir(path), path, fileBytes, evaluateBases)
}

// loadBytes keeps helmfile's document-splitting behavior (`---`) and merges the
// resulting partial states back into one final HelmState.
func (ld *helmfileStateLoader) loadBytes(
	inheritedEnv *environment.Environment,
	overrodeEnv *environment.Environment,
	baseDir string,
	filename string,
	content []byte,
	evaluateBases bool,
) (*state.HelmState, error) {
	normalizedContent := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	parts := bytes.Split(normalizedContent, []byte("\n---\n"))

	var finalState *state.HelmState
	env := inheritedEnv

	for i, part := range parts {
		partID := fmt.Sprintf("%s.part.%d", filename, i)

		rendered, err := ld.renderTemplatesToYAML(baseDir, partID, part, env, overrodeEnv)
		if err != nil {
			return nil, fmt.Errorf("rendering helmfile %s: %w", partID, err)
		}

		currentState, err := ld.parseAndLoad(rendered.Bytes(), baseDir, filename, evaluateBases, env, overrodeEnv)
		if err != nil {
			return nil, err
		}

		if finalState == nil {
			finalState = currentState
		} else {
			mergeHelmState(finalState, currentState)
		}

		env = &finalState.Env
	}

	if finalState == nil {
		return nil, fmt.Errorf("no helmfile state loaded from %s", filename)
	}

	return finalState, nil
}

// parseAndLoad delegates parsing to helmfile's exported creator so environment,
// bases, and state inheritance stay aligned with upstream semantics.
func (ld *helmfileStateLoader) parseAndLoad(
	content []byte,
	baseDir string,
	filename string,
	evaluateBases bool,
	inheritedEnv *environment.Environment,
	overrodeEnv *environment.Environment,
) (*state.HelmState, error) {
	mergedEnv, err := environment.New(ld.env).Merge(nil)
	if err != nil {
		return nil, err
	}
	if inheritedEnv != nil {
		mergedEnv, err = inheritedEnv.Merge(nil)
		if err != nil {
			return nil, err
		}
	}
	if overrodeEnv != nil {
		mergedEnv, err = mergedEnv.Merge(overrodeEnv)
		if err != nil {
			return nil, err
		}
	}

	creator := state.NewCreator(
		ld.logger,
		ld.fs,
		ld.valsRuntime,
		ld.getHelm,
		"",
		"",
		ld.remote,
		false,
		"",
	)
	creator.LoadFile = func(inherited, overrode *environment.Environment, parentBaseDir, file string, nestedEvaluateBases bool) (*state.HelmState, error) {
		located, locateErr := ld.remote.Locate(file, "states")
		if locateErr != nil {
			return nil, fmt.Errorf("locate helmfile %s: %w", file, locateErr)
		}
		resolvedFile := located
		if !filepath.IsAbs(resolvedFile) {
			resolvedFile = filepath.Join(parentBaseDir, resolvedFile)
		}
		return ld.loadFileWithEnv(inherited, overrode, filepath.Dir(resolvedFile), filepath.Base(resolvedFile), nestedEvaluateBases)
	}

	st, err := creator.ParseAndLoad(content, baseDir, filename, ld.env, false, evaluateBases, mergedEnv, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing helmfile %s: %w", filename, err)
	}

	expandedHelmfiles, err := st.ExpandedHelmfiles()
	if err != nil {
		return nil, fmt.Errorf("expanding helmfiles in %s: %w", filename, err)
	}
	st.Helmfiles = expandedHelmfiles

	return st, nil
}

// renderTemplatesToYAML performs the first-stage gotmpl rendering for helmfile
// documents before state.ParseAndLoad applies the rest of helmfile semantics.
func (ld *helmfileStateLoader) renderTemplatesToYAML(
	baseDir string,
	filename string,
	content []byte,
	inheritedEnv *environment.Environment,
	overrodeEnv *environment.Environment,
) (*bytes.Buffer, error) {
	env := inheritedEnv
	if env == nil {
		env = &environment.Environment{Name: ld.env, Values: map[string]any(nil)}
	}

	finalEnv, err := env.Merge(overrodeEnv)
	if err != nil {
		return nil, err
	}
	valsMap, err := finalEnv.GetMergedValues()
	if err != nil {
		return nil, err
	}

	tmplData := state.NewEnvironmentTemplateData(*finalEnv, ld.namespace, valsMap)
	renderer := tmpl.NewFileRenderer(ld.fs, baseDir, tmplData)
	rendered, err := renderer.RenderTemplateContentToBuffer(content)
	if err != nil {
		return nil, fmt.Errorf("rendering %s: %w", filename, err)
	}

	return rendered, nil
}

// getHelm lazily constructs helm executors because helmfile may need a real helm
// binary during load time for secret decryption and related helper operations.
func (ld *helmfileStateLoader) getHelm(st *state.HelmState) (helmexec.Interface, error) {
	bin := st.DefaultHelmBinary
	if strings.TrimSpace(bin) == "" {
		bin = state.DefaultHelmBinary
	}
	kubeContext := st.HelmDefaults.KubeContext
	cacheKey := bin + "\x00" + kubeContext

	ld.helmMu.Lock()
	defer ld.helmMu.Unlock()

	if execer, ok := ld.helmCache[cacheKey]; ok {
		return execer, nil
	}

	execer, err := newHelmExec(bin, ld.logger, kubeContext, &helmexec.ShellRunner{
		Logger: ld.logger,
		Ctx:    context.Background(),
	})
	if err != nil {
		return nil, fmt.Errorf("create helm executor for binary %q: %w", bin, err)
	}

	ld.helmCache[cacheKey] = execer
	return execer, nil
}

// newHelmExec exists as a seam for tests; production always uses helmexec.New.
func newHelmExec(
	bin string,
	logger *zap.SugaredLogger,
	kubeContext string,
	runner helmexec.Runner,
) (helmexec.Interface, error) {
	return helmexec.New(bin, helmexec.HelmExecOptions{}, logger, "", kubeContext, runner)
}

// mergeHelmState mirrors helmfile's "later document wins" merge style while
// preserving appended release lists and the final rendered environment.
func mergeHelmState(dst, src *state.HelmState) {
	_ = mergo.Merge(&dst.ReleaseSetSpec, &src.ReleaseSetSpec, mergo.WithOverride, mergo.WithAppendSlice)
	dst.RenderedValues = src.RenderedValues
	dst.Env = src.Env
}

// selectSingleReleaseFromState intentionally projects only release names first,
// so selector-related errors stay small and user-facing.
func selectSingleReleaseFromState(st *state.HelmState) (*state.ReleaseSpec, error) {
	releases := make([]HelmfileResolvedRelease, 0, len(st.Releases))
	for _, release := range st.Releases {
		releases = append(releases, HelmfileResolvedRelease{ReleaseName: release.Name})
	}

	selected, err := selectSingleRelease(releases)
	if err != nil {
		return nil, err
	}
	for i := range st.Releases {
		if st.Releases[i].Name == selected.ReleaseName {
			return &st.Releases[i], nil
		}
	}
	return nil, fmt.Errorf("selected release %q not found", selected.ReleaseName)
}

func selectSingleRelease(releases []HelmfileResolvedRelease) (*HelmfileResolvedRelease, error) {
	if len(releases) != 1 {
		return nil, fmt.Errorf("expected exactly 1 release after applying selectors, got %d", len(releases))
	}

	selected := releases[0]
	return &selected, nil
}

func effectiveHelmfileNamespace(releaseNamespace, inputNamespace string) string {
	namespace := strings.TrimSpace(releaseNamespace)
	if namespace == "" {
		namespace = strings.TrimSpace(inputNamespace)
	}
	if namespace == "" {
		namespace = "default"
	}
	return namespace
}

func (ld *helmfileStateLoader) normalizeHelmfileHooks(
	st *state.HelmState,
	release *state.ReleaseSpec,
) ([]HelmfileResolvedHook, error) {
	if release == nil || len(release.Hooks) == 0 {
		return nil, nil
	}

	hooks := make([]HelmfileResolvedHook, 0, len(release.Hooks))
	order := 0
	for _, hook := range release.Hooks {
		for _, rawEvent := range hook.Events {
			event := strings.TrimSpace(rawEvent)
			switch event {
			case "preapply", "presync", "postsync":
				command, err := ld.renderHelmfileHookText(st, release, event, hook.Command)
				if err != nil {
					return nil, fmt.Errorf("rendering hook[%d] command for event %s: %w", order, event, err)
				}
				args := make([]string, len(hook.Args))
				for i := range hook.Args {
					args[i], err = ld.renderHelmfileHookText(st, release, event, hook.Args[i])
					if err != nil {
						return nil, fmt.Errorf("rendering hook[%d] args[%d] for event %s: %w", order, i, event, err)
					}
				}
				hooks = append(hooks, HelmfileResolvedHook{
					Event:   event,
					Command: command,
					Args:    args,
					Order:   order,
				})
				order++
			}
		}
	}

	return hooks, nil
}

func (ld *helmfileStateLoader) renderHelmfileHookText(
	st *state.HelmState,
	release *state.ReleaseSpec,
	event string,
	text string,
) (string, error) {
	values := map[string]any{}
	if st != nil && st.RenderedValues != nil {
		values = st.RenderedValues
	}
	data := map[string]any{
		"Environment":     st.Env,
		"Namespace":       st.OverrideNamespace,
		"Event":           map[string]any{"Name": event, "Error": nil},
		"Values":          values,
		"Release":         release,
		"HelmfileCommand": "drplan-gen helmfile",
	}
	renderer := tmpl.NewTextRenderer(ld.fs, ld.baseDir, data)
	return renderer.RenderTemplateText(text)
}

// normalizeChartName converts local tgz paths, local directories, and repo/chart
// references into the plain chart name expected by Clusternet HelmChart.spec.chart.
func normalizeChartName(chartRef, version string) string {
	base := filepath.Base(strings.TrimSpace(chartRef))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if version != "" {
		base = strings.TrimSuffix(base, "-"+version)
	}
	return base
}

// buildReleaseValuesYAML reproduces helmfile's final merged values and then,
// unless keepFullValues is requested, trims them down to the delta from chart defaults.
func buildReleaseValuesYAML(st *state.HelmState, release *state.ReleaseSpec, keepFullValues bool) (string, error) {
	merged := map[string]interface{}{}
	for _, entry := range release.Values {
		valueMaps, err := loadReleaseValueMaps(st, release, entry)
		if err != nil {
			return "", err
		}
		for _, valueMap := range valueMaps {
			if len(valueMap) == 0 {
				continue
			}
			mergeStringMaps(merged, valueMap)
		}
	}

	if len(merged) == 0 {
		return "", nil
	}

	valuesForOutput := merged
	if !keepFullValues {
		diffValues, err := diffValuesAgainstChartDefaults(st.FilePath, release.Chart, merged)
		if err != nil {
			return "", err
		}
		if len(diffValues) == 0 {
			return "", nil
		}
		valuesForOutput = diffValues
	}

	rendered, err := sigyaml.Marshal(valuesForOutput)
	if err != nil {
		return "", fmt.Errorf("marshaling merged values: %w", err)
	}
	return string(rendered), nil
}

// loadReleaseValueMaps normalizes all supported helmfile values entry shapes
// into `[]map[string]interface{}` so later merge logic can stay uniform.
func loadReleaseValueMaps(st *state.HelmState, release *state.ReleaseSpec, entry any) ([]map[string]interface{}, error) {
	switch value := entry.(type) {
	case string:
		paths, skip, err := resolveReleaseValueFiles(st, release, value)
		if err != nil {
			return nil, err
		}
		if skip {
			return nil, nil
		}
		if len(paths) > 1 {
			return nil, fmt.Errorf("glob patterns in release values are not supported yet. please submit a feature request if necessary")
		}

		valueMaps := make([]map[string]interface{}, 0, len(paths))
		for _, path := range paths {
			yamlBytes, err := st.RenderReleaseValuesFileToBytes(release, path)
			if err != nil {
				return nil, fmt.Errorf("rendering release values file %s: %w", value, err)
			}
			valueMap := map[string]interface{}{}
			if err := sigyaml.Unmarshal(yamlBytes, &valueMap); err != nil {
				return nil, fmt.Errorf("unmarshaling release values file %s: %w", value, err)
			}
			valueMaps = append(valueMaps, valueMap)
		}
		return valueMaps, nil
	case map[string]interface{}, map[any]any:
		valueMap, err := normalizeReleaseValueEntry(value)
		if err != nil {
			return nil, err
		}
		return []map[string]interface{}{valueMap}, nil
	default:
		return nil, fmt.Errorf("unsupported release values entry type %T", entry)
	}
}

// resolveReleaseValueFiles honors helmfile's missingFileHandler behavior.
// Missing files can therefore be skipped instead of hard-failing when the state requests it.
func resolveReleaseValueFiles(st *state.HelmState, release *state.ReleaseSpec, value string) ([]string, bool, error) {
	path := release.ValuesPathPrefix + value
	if remote.IsRemote(path) {
		return nil, false, fmt.Errorf("remote release values file %q is not supported in helmfile mode", value)
	}

	normalizedPath := path
	if !filepath.IsAbs(normalizedPath) {
		normalizedPath = filepath.Join(filepath.Dir(st.FilePath), normalizedPath)
	}
	normalizedPath = filepath.Clean(normalizedPath)

	matches, err := filepath.Glob(normalizedPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed processing %s: %w", value, err)
	}
	sort.Strings(matches)
	if len(matches) > 0 {
		return matches, false, nil
	}

	handler := effectiveReleaseMissingFileHandler(st, release)
	switch *handler {
	case state.MissingFileHandlerWarn, state.MissingFileHandlerInfo, state.MissingFileHandlerDebug:
		return nil, true, nil
	default:
		return nil, false, fmt.Errorf("values file matching %q does not exist in %q", path, filepath.Dir(st.FilePath))
	}
}

func effectiveReleaseMissingFileHandler(st *state.HelmState, release *state.ReleaseSpec) *string {
	defaultMissingFileHandler := state.MissingFileHandlerError

	switch {
	case release.MissingFileHandler != nil:
		return release.MissingFileHandler
	case st.MissingFileHandler != nil:
		return st.MissingFileHandler
	default:
		return &defaultMissingFileHandler
	}
}

// diffValuesAgainstChartDefaults loads the local chart and removes any key whose
// final rendered value is identical to the chart's built-in default.
func diffValuesAgainstChartDefaults(stateFile, chartRef string, merged map[string]interface{}) (map[string]interface{}, error) {
	chartPath, err := resolveChartDefaultSource(stateFile, chartRef)
	if err != nil {
		return nil, err
	}

	ch, err := loader.Load(chartPath)
	if err != nil {
		return nil, fmt.Errorf("loading chart %q: %w", chartPath, err)
	}

	defaultValues, err := chartutil.CoalesceValues(ch, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("coalescing chart default values for %q: %w", chartPath, err)
	}

	normalizedDefaults, err := normalizeStringMap(defaultValues)
	if err != nil {
		return nil, fmt.Errorf("normalizing chart default values for %q: %w", chartPath, err)
	}

	return diffStringMaps(merged, normalizedDefaults), nil
}

// resolveChartDefaultSource deliberately supports only local chart directories or
// local tgz archives. Remote repos would require fetching artifacts during generation.
func resolveChartDefaultSource(stateFile, chartRef string) (string, error) {
	chartRef = strings.TrimSpace(chartRef)
	if chartRef == "" {
		return "", fmt.Errorf("release chart is required")
	}

	if strings.Contains(chartRef, "://") {
		return "", fmt.Errorf("chart %q is not a local chart path; use --keep-full-values to keep rendered values without diffing against chart defaults", chartRef)
	}

	chartPath := chartRef
	if !filepath.IsAbs(chartPath) {
		chartPath = filepath.Clean(filepath.Join(filepath.Dir(stateFile), chartPath))
	}

	info, err := os.Stat(chartPath)
	if err == nil {
		if info.IsDir() || strings.HasSuffix(chartPath, ".tgz") {
			return chartPath, nil
		}
		return chartPath, nil
	}
	if os.IsNotExist(err) {
		if looksLikeLocalChartRef(chartRef) {
			return "", fmt.Errorf("local chart path %q does not exist", chartPath)
		}
		return "", fmt.Errorf("chart %q is not a local chart path; use --keep-full-values to keep rendered values without diffing against chart defaults", chartRef)
	}
	return "", fmt.Errorf("checking chart path %q: %w", chartPath, err)
}

func looksLikeLocalChartRef(chartRef string) bool {
	return filepath.IsAbs(chartRef) ||
		strings.HasPrefix(chartRef, ".") ||
		strings.HasSuffix(chartRef, ".tgz") ||
		strings.Contains(chartRef, string(filepath.Separator))
}

// normalizeStringMap round-trips through YAML so nested map[any]any values are
// converted into consistent `map[string]interface{}` structures before diffing.
func normalizeStringMap(value any) (map[string]interface{}, error) {
	raw, err := sigyaml.Marshal(value)
	if err != nil {
		return nil, err
	}

	normalized := map[string]interface{}{}
	if err := sigyaml.Unmarshal(raw, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

// diffStringMaps is the recursive core of the values minimization step.
// Keys absent from defaults or changed relative to defaults are retained.
func diffStringMaps(current, defaults map[string]interface{}) map[string]interface{} {
	diff := map[string]interface{}{}

	for key, currentValue := range current {
		defaultValue, exists := defaults[key]
		if !exists {
			diff[key] = currentValue
			continue
		}

		currentMap, currentIsMap := currentValue.(map[string]interface{})
		defaultMap, defaultIsMap := defaultValue.(map[string]interface{})
		if currentIsMap && defaultIsMap {
			nestedDiff := diffStringMaps(currentMap, defaultMap)
			if len(nestedDiff) > 0 {
				diff[key] = nestedDiff
			}
			continue
		}

		if !reflect.DeepEqual(currentValue, defaultValue) {
			diff[key] = currentValue
		}
	}

	return diff
}

// normalizeReleaseValueEntry handles inline values entries from helmfile state
// by converting arbitrary YAML-compatible maps into string-keyed maps.
func normalizeReleaseValueEntry(entry any) (map[string]interface{}, error) {
	switch value := entry.(type) {
	case map[string]interface{}, map[any]any:
		rawBytes, err := sigyaml.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("marshaling inline values: %w", err)
		}
		valueMap := map[string]interface{}{}
		if err := sigyaml.Unmarshal(rawBytes, &valueMap); err != nil {
			return nil, fmt.Errorf("unmarshaling inline values: %w", err)
		}
		return valueMap, nil
	default:
		return nil, fmt.Errorf("unsupported embedded release values entry type %T", entry)
	}
}

// mergeStringMaps applies later values entries on top of earlier ones, matching
// the same precedence users expect from helmfile values merging.
func mergeStringMaps(dst, src map[string]interface{}) {
	for key, value := range src {
		srcMap, srcIsMap := value.(map[string]interface{})
		dstMap, dstIsMap := dst[key].(map[string]interface{})
		if srcIsMap && dstIsMap {
			mergeStringMaps(dstMap, srcMap)
			dst[key] = dstMap
			continue
		}
		dst[key] = value
	}
}

// effectiveReleaseWait mirrors helmfile's precedence: release override first,
// then helmDefaults.
func effectiveReleaseWait(st *state.HelmState, release *state.ReleaseSpec) *bool {
	if release.Wait != nil {
		return boolPtrForLoader(*release.Wait)
	}
	return boolPtrForLoader(st.HelmDefaults.Wait)
}

// effectiveReleaseWaitForJob mirrors helmfile's waitForJobs precedence.
func effectiveReleaseWaitForJob(st *state.HelmState, release *state.ReleaseSpec) *bool {
	if release.WaitForJobs != nil {
		return boolPtrForLoader(*release.WaitForJobs)
	}
	return boolPtrForLoader(st.HelmDefaults.WaitForJobs)
}

// effectiveReleaseAtomic mirrors helmfile's atomic precedence and later maps
// to both install and upgrade atomic behavior in Clusternet generation.
func effectiveReleaseAtomic(st *state.HelmState, release *state.ReleaseSpec) *bool {
	if release.Atomic != nil {
		return boolPtrForLoader(*release.Atomic)
	}
	return boolPtrForLoader(st.HelmDefaults.Atomic)
}

// effectiveReleaseCreateNamespace mirrors helmfile's createNamespace
// precedence. Helmfile defaults to true when neither release nor defaults set it.
func effectiveReleaseCreateNamespace(st *state.HelmState, release *state.ReleaseSpec) *bool {
	if release.CreateNamespace != nil {
		return boolPtrForLoader(*release.CreateNamespace)
	}
	if st.HelmDefaults.CreateNamespace != nil {
		return boolPtrForLoader(*st.HelmDefaults.CreateNamespace)
	}
	return boolPtrForLoader(true)
}

// effectiveReleaseTimeoutSeconds uses release timeout first and falls back to helmDefaults.
func effectiveReleaseTimeoutSeconds(st *state.HelmState, release *state.ReleaseSpec) int32 {
	if release.Timeout != nil && *release.Timeout > 0 {
		return int32(*release.Timeout)
	}
	if st.HelmDefaults.Timeout > 0 {
		return int32(st.HelmDefaults.Timeout)
	}
	return 0
}

func boolPtrForLoader(v bool) *bool {
	return &v
}
