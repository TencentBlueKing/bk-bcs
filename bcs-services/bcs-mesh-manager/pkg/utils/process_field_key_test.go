package utils

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
	pointer "k8s.io/utils/pointer"
)

// getMap tries to convert interface{} to map[string]interface{} or map[interface{}]interface{}
func getMap(m interface{}) (map[string]interface{}, map[interface{}]interface{}, bool) {
	if m == nil {
		return nil, nil, false
	}
	if mm, ok := m.(map[string]interface{}); ok {
		return mm, nil, true
	}
	if mm, ok := m.(map[interface{}]interface{}); ok {
		return nil, mm, true
	}
	return nil, nil, false
}

func TestProcessFieldKey(t *testing.T) {
	t.Run("TestProcessFieldKeyWithAutoscaleDisabled", func(t *testing.T) {
		values := `pilot:
  autoscaleEnabled: true
  autoscaleMin: 2
  autoscaleMax: 5
  cpu:
    targetAverageUtilization: 80`

		options := &UpdateValuesOptions{
			AutoscaleEnabled: pointer.Bool(false),
		}

		result, err := ProcessValues(values, options)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		if strings.Contains(result, "autoscaleMin") {
			t.Error("autoscaleMin should be deleted when autoscale is disabled")
		}
		if strings.Contains(result, "autoscaleMax") {
			t.Error("autoscaleMax should be deleted when autoscale is disabled")
		}
		if strings.Contains(result, "cpu:") {
			t.Error("cpu should be deleted when autoscale is disabled")
		}
	})

	t.Run("TestProcessFieldKeyWithDedicatedNodeDisabled", func(t *testing.T) {
		values := `pilot:
  nodeSelector:
    node-type: istio-control
    zone: az1
  tolerations:
    - operator: Exists`

		options := &UpdateValuesOptions{
			DedicatedNodeEnabled: pointer.Bool(false),
		}

		result, err := ProcessValues(values, options)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		if strings.Contains(result, "nodeSelector") {
			t.Error("nodeSelector should be deleted when dedicatedNode is disabled")
		}
		if strings.Contains(result, "tolerations") {
			t.Error("tolerations should be deleted when dedicatedNode is disabled")
		}
	})

	t.Run("TestProcessFieldKeyWithLogCollectorDisabled", func(t *testing.T) {
		values := `meshConfig:
  accessLogFile: /dev/stdout
  accessLogFormat: '{\"timestamp\":\"%START_TIME%\"}'
  accessLogEncoding: JSON`

		options := &UpdateValuesOptions{
			LogCollectorConfigEnabled: pointer.Bool(false),
		}

		result, err := ProcessValues(values, options)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		if strings.Contains(result, "accessLogFile") {
			t.Error("accessLogFile should be deleted when logCollector is disabled")
		}
		if strings.Contains(result, "accessLogFormat") {
			t.Error("accessLogFormat should be deleted when logCollector is disabled")
		}
		if strings.Contains(result, "accessLogEncoding") {
			t.Error("accessLogEncoding should be deleted when logCollector is disabled")
		}
	})

	t.Run("TestProcessFieldKeyWithTracingDisabled", func(t *testing.T) {
		values := `meshConfig:
  enableTracing: true
  extensionProviders:
    - name: otel-tracing
  defaultConfig:
    tracingConfig:
      zipkin:
        address: http://zipkin:9411/api/v2/spans
pilot:
  traceSampling: 0.1`

		options := &UpdateValuesOptions{
			EnableTracing: pointer.Bool(false),
		}

		result, err := ProcessValues(values, options)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		// 验证追踪相关字段被删除
		var resultMap map[string]interface{}
		yaml.Unmarshal([]byte(result), &resultMap)

		meshConfig, _, ok := getMap(resultMap["meshConfig"])
		if ok {
			if _, exists := meshConfig["extensionProviders"]; exists {
				t.Error("extensionProviders should be deleted when tracing is disabled")
			}
			defCfg, _, ok := getMap(meshConfig["defaultConfig"])
			if ok {
				tracingCfg, _, ok := getMap(defCfg["tracingConfig"])
				if ok {
					if _, exists := tracingCfg["zipkin"]; exists {
						t.Error("zipkin should be deleted when tracing is disabled")
					}
				}
			}
		}
		pilot, _, ok := getMap(resultMap["pilot"])
		if ok {
			if _, exists := pilot["traceSampling"]; exists {
				t.Error("traceSampling should be deleted when tracing is disabled")
			}
		}
	})

	t.Run("TestProcessFieldKeyWithAllOptionsEnabled", func(t *testing.T) {
		values := `pilot:
  autoscaleMin: 2
  autoscaleMax: 5
  cpu:
    targetAverageUtilization: 80
  nodeSelector:
    node-type: istio-control
  tolerations:
    - operator: Exists
meshConfig:
  accessLogFile: /dev/stdout
  accessLogFormat: '{"timestamp":"%START_TIME%"}'
  accessLogEncoding: JSON
  extensionProviders:
    - name: otel-tracing
  defaultConfig:
    tracingConfig:
      zipkin:
        address: http://zipkin:9411/api/v2/spans`

		options := &UpdateValuesOptions{
			AutoscaleEnabled:          pointer.Bool(true),
			DedicatedNodeEnabled:      pointer.Bool(true),
			LogCollectorConfigEnabled: pointer.Bool(true),
			EnableTracing:             pointer.Bool(true),
		}

		result, err := ProcessValues(values, options)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		// 当所有选项都启用时，字段应该保持不变（不被删除）
		if !strings.Contains(result, "autoscaleMin:") {
			t.Error("autoscaleMin should remain when autoscale is enabled")
		}
		if !strings.Contains(result, "autoscaleMax:") {
			t.Error("autoscaleMax should remain when autoscale is enabled")
		}
		if !strings.Contains(result, "cpu:") {
			t.Error("cpu should remain when autoscale is enabled")
		}
		if !strings.Contains(result, "nodeSelector:") {
			t.Error("nodeSelector should remain when dedicatedNode is enabled")
		}
		if !strings.Contains(result, "tolerations:") {
			t.Error("tolerations should remain when dedicatedNode is enabled")
		}
		if !strings.Contains(result, "accessLogFile:") {
			t.Error("accessLogFile should remain when logCollector is enabled")
		}
		if !strings.Contains(result, "accessLogFormat:") {
			t.Error("accessLogFormat should remain when logCollector is enabled")
		}
		if !strings.Contains(result, "accessLogEncoding:") {
			t.Error("accessLogEncoding should remain when logCollector is enabled")
		}
		if !strings.Contains(result, "extensionProviders:") {
			t.Error("extensionProviders should remain when tracing is enabled")
		}
	})

	t.Run("TestProcessFieldKeyWithNilOptions", func(t *testing.T) {
		values := `pilot:
  autoscaleEnabled: true
meshConfig:
  accessLogFile: /dev/stdout`

		result, err := ProcessValues(values, nil)
		if err != nil {
			t.Fatalf("ProcessFieldKey failed: %v", err)
		}

		if result != values {
			t.Error("Result should be unchanged when options is nil")
		}
	})
}
