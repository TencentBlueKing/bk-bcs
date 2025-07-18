/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package common 提供监控相关的YAML模板
package common

import "fmt"

// PodMonitorTemplate PodMonitor资源模板，用于Istio数据面监控
// nolint:lll
const PodMonitorTemplate = `apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  labels:
    monitoring: istio-proxies
    created-by: bcs-mesh-manager
  name: %s
  namespace: istio-system
spec:
  jobLabel: envoy-stats
  namespaceSelector:
    any: true
  podMetricsEndpoints:
  - bearerTokenSecret:
      key: ""
    interval: 60s
    metricRelabelings:
    - action: labeldrop
      regex: (tcloud_region_name|tcloud_region_abbr|org|environment|rollouts_pod_template_hash|security_istio_io_tlsMode|source_version|destination_version|source_principal|source_app|service_istio_io_canonical_name|destination_principal|destination_app)
    - action: drop
      regex: envoy_wasm_remote_load_fetch_successes|istio_agent_go_memstats_mcache_inuse_bytes|envoy_cluster_manager_cds_update_rejected|envoy_cluster_manager_update_out_of_merge_window|istio_request_bytes_bucket|envoy_cluster_manager_cds_init_fetch_timeout|istio_agent_pilot_endpoint_not_ready|envoy_cluster_manager_cluster_removed|istio_agent_pilot_duplicate_envoy_clusters|envoy_listener_manager_workers_started|istio_tcp_connections_opened_total|envoy_cluster_lb_zone_cluster_too_small|envoy_cluster_manager_warming_clusters|envoy_cluster_upstream_cx_connect_fail|envoy_server_seconds_until_first_ocsp_response_expiring|istio_agent_go_memstats_other_sys_bytes|envoy_cluster_manager_cluster_updated_via_merge|envoy_cluster_original_dst_host_invalid|istio_agent_go_memstats_sys_bytes|envoy_listener_manager_total_listeners_active|istio_agent_go_memstats_mallocs_total|envoy_cluster_upstream_cx_rx_bytes_total|envoy_listener_manager_listener_removed|istio_agent_go_memstats_lookups_total|envoy_cluster_membership_excluded|envoy_cluster_upstream_cx_destroy_local_with_active_rq|istio_agent_pilot_proxy_queue_time_count|envoy_cluster_manager_cds_version|envoy_listener_manager_total_listeners_draining|istio_agent_endpoint_no_pod|envoy_cluster_upstream_cx_http3_total|envoy_cluster_upstream_cx_none_healthy|envoy_server_live|envoy_server_state|envoy_server_total_connections|istio_agent_outgoing_latency|envoy_cluster_internal_upstream_rq|envoy_cluster_manager_update_merge_canceled|envoy_cluster_upstream_cx_rx_bytes_buffered|istio_agent_pilot_xds_pushes|envoy_cluster_internal_upstream_rq_completed|envoy_server_main_thread_watchdog_mega_miss|envoy_cluster_manager_cds_update_duration_sum|envoy_server_compilation_settings_fips_mode|istio_agent_go_memstats_heap_alloc_bytes|istio_agent_process_open_fds|envoy_cluster_assignment_timeout_received|envoy_cluster_lb_subsets_removed|envoy_cluster_upstream_rq_total|envoy_cluster_http2_inbound_empty_frames_flood|envoy_cluster_manager_cds_update_duration_bucket|istio_agent_go_memstats_heap_objects|istio_agent_scrapes_total|envoy_cluster_circuit_breakers_default_rq_open|envoy_cluster_update_failure|envoy_cluster_upstream_rq_maintenance_mode|envoy_server_days_until_first_cert_expiring|envoy_cluster_upstream_cx_connect_ms_count|envoy_cluster_upstream_flow_control_resumed_reading_total|envoy_server_memory_allocated|envoy_server_uptime|istio_agent_go_memstats_heap_idle_bytes|istio_agent_go_memstats_next_gc_bytes|istio_agent_pilot_vservice_dup_domain|istio_response_bytes_count|envoy_cluster_manager_cluster_updated|envoy_cluster_upstream_rq_pending_overflow|envoy_cluster_membership_total|envoy_cluster_upstream_rq|envoy_wasm_remote_load_cache_hits|envoy_cluster_http2_inbound_priority_frames_flood|envoy_cluster_http2_pending_send_bytes|envoy_cluster_lb_subsets_created|envoy_cluster_manager_cluster_added|envoy_cluster_upstream_cx_connect_ms_bucket|envoy_listener_manager_listener_modified|istio_agent_go_info|istio_request_bytes_count|envoy_cluster_circuit_breakers_high_cx_pool_open|envoy_cluster_internal_upstream_rq_200|envoy_listener_manager_lds_update_success|envoy_server_hot_restart_epoch|istio_agent_pilot_destrule_subsets|istio_agent_pilot_no_ip|istio_agent_pilot_xds_config_size_bytes_sum|envoy_cluster_upstream_flow_control_drained_total|envoy_cluster_upstream_rq_tx_reset|envoy_cluster_upstream_cx_tx_bytes_total|envoy_cluster_upstream_rq_retry_backoff_exponential|envoy_cluster_upstream_cx_length_ms_count|envoy_cluster_upstream_rq_per_try_timeout|envoy_listener_manager_listener_create_success|istio_agent_num_outgoing_requests|envoy_cluster_manager_cds_update_success|envoy_cluster_update_empty|istio_agent_dns_requests_total|istio_agent_pilot_xds_send_time_sum|envoy_listener_manager_lds_update_duration_sum|envoy_server_memory_heap_size|istio_agent_pilot_proxy_queue_time_bucket|istio_agent_pilot_xds_push_time_bucket|istio_agent_process_max_fds|istio_response_bytes_bucket|envoy_cluster_circuit_breakers_default_rq_retry_open|istio_agent_go_memstats_heap_inuse_bytes|istio_agent_pilot_xds_push_time_sum|envoy_cluster_lb_zone_routing_sampled|envoy_cluster_upstream_cx_destroy_with_active_rq|envoy_cluster_upstream_cx_destroy_local|envoy_server_debug_assertion_failures|envoy_wasm_remote_load_cache_misses|istio_agent_go_memstats_alloc_bytes_total|istio_agent_go_memstats_last_gc_time_seconds|istio_tcp_sent_bytes_total|envoy_cluster_http2_metadata_empty_frames|envoy_cluster_upstream_cx_destroy|istio_agent_go_gc_duration_seconds_sum|istio_build|envoy_cluster_http2_dropped_headers_with_underscores|envoy_cluster_update_attempt|istio_agent_dns_upstream_request_duration_seconds_count|istio_agent_pilot_push_triggers|envoy_cluster_http2_tx_reset|envoy_cluster_manager_active_clusters|envoy_server_envoy_bug_failures|istio_agent_pilot_conflict_outbound_listener_tcp_over_current_tcp|istio_agent_wasm_cache_entries|envoy_cluster_upstream_cx_http1_total|envoy_cluster_upstream_cx_max_requests|envoy_cluster_circuit_breakers_high_rq_retry_open|envoy_cluster_upstream_rq_retry_success|envoy_listener_manager_listener_stopped|envoy_cluster_circuit_breakers_default_cx_pool_open|envoy_cluster_circuit_breakers_high_rq_open|envoy_cluster_upstream_cx_protocol_error|envoy_server_wip_protos|istio_agent_num_file_secret_failures_total|envoy_cluster_lb_subsets_active|envoy_cluster_manager_cluster_modified|envoy_cluster_update_success|istio_agent_go_memstats_mspan_inuse_bytes|envoy_cluster_http2_inbound_window_update_frames_flood|envoy_cluster_http2_rx_messaging_error|istio_response_messages_total|istio_tcp_received_bytes_total|istio_agent_dns_upstream_request_duration_seconds_bucket|istio_agent_pilot_xds_push_time_count|envoy_server_main_thread_watchdog_miss|istio_agent_go_memstats_mcache_sys_bytes|envoy_server_worker_0_watchdog_miss|envoy_cluster_upstream_cx_idle_timeout|envoy_cluster_upstream_rq_retry_overflow|envoy_cluster_upstream_flow_control_paused_reading_total|istio_agent_process_resident_memory_bytes|envoy_cluster_http2_headers_cb_no_stream|envoy_cluster_upstream_cx_connect_ms_sum|istio_agent_go_memstats_alloc_bytes|istio_agent_pilot_xds_send_time_count|istio_agent_process_virtual_memory_bytes|istio_agent_scrape_failures_total|metric_cache_count|envoy_cluster_http2_rx_reset|envoy_cluster_upstream_cx_length_ms_sum|istio_agent_pilot_xds_config_size_bytes_bucket|istio_agent_pilot_xds_config_size_bytes_count|envoy_cluster_upstream_cx_pool_overflow|envoy_server_initialization_time_ms_bucket|istio_request_bytes_sum|envoy_cluster_lb_zone_no_capacity_left|envoy_listener_manager_lds_update_rejected|envoy_cluster_upstream_cx_active|envoy_cluster_upstream_cx_destroy_remote|envoy_cluster_upstream_rq_retry_backoff_ratelimited|envoy_listener_manager_lds_update_duration_bucket|envoy_cluster_lb_subsets_fallback|envoy_cluster_update_no_rebuild|envoy_listener_manager_total_listeners_warming|envoy_cluster_lb_subsets_selected|envoy_cluster_upstream_internal_redirect_failed_total|envoy_cluster_upstream_rq_rx_reset|envoy_cluster_circuit_breakers_default_rq_pending_open|envoy_cluster_circuit_breakers_high_rq_pending_open|envoy_cluster_upstream_rq_active|istio_tcp_connections_closed_total|envoy_cluster_http2_tx_flush_timeout|envoy_cluster_max_host_weight|envoy_cluster_upstream_cx_length_ms_bucket|istio_agent_pilot_conflict_outbound_listener_tcp_over_current_http|envoy_cluster_upstream_internal_redirect_succeeded_total|envoy_listener_manager_listener_added|envoy_server_static_unknown_fields|istio_agent_go_memstats_heap_released_bytes|istio_request_messages_total|envoy_cluster_http2_outbound_control_flood|envoy_cluster_membership_change|envoy_cluster_upstream_rq_retry|istio_agent_go_memstats_frees_total|istio_agent_pilot_xds_send_time_bucket|envoy_cluster_upstream_cx_connect_attempts_exceeded|envoy_cluster_upstream_cx_http2_total|envoy_cluster_default_total_match_count|envoy_cluster_upstream_rq_canceled|envoy_wasm_remote_load_cache_entries|istio_agent_go_memstats_gc_sys_bytes|envoy_cluster_lb_subsets_fallback_panic|envoy_server_parent_connections|istio_agent_istiod_connection_terminations|istio_agent_process_cpu_seconds_total|envoy_cluster_http2_requests_rejected_with_underscores_in_headers|envoy_cluster_upstream_cx_destroy_remote_with_active_rq|envoy_cluster_manager_cds_update_duration_count|envoy_cluster_upstream_cx_total|envoy_cluster_upstream_rq_200|envoy_cluster_upstream_rq_pending_total|envoy_listener_manager_lds_update_duration_count|envoy_server_dropped_stat_flushes|envoy_cluster_assignment_stale|envoy_cluster_circuit_breakers_default_cx_open|istio_agent_process_virtual_memory_max_bytes|envoy_wasm_remote_load_cache_negative_hits|istio_agent_go_threads|envoy_server_initialization_time_ms_count|istio_agent_go_memstats_buck_hash_sys_bytes|istio_agent_pilot_proxy_queue_time_sum|envoy_cluster_http2_keepalive_timeout|envoy_cluster_manager_cds_update_failure|istio_agent_startup_duration_seconds|envoy_cluster_lb_zone_routing_cross_zone|istio_agent_pilot_xds|envoy_cluster_lb_local_cluster_not_ok|envoy_cluster_lb_recalculate_zone_structures|envoy_cluster_manager_cds_update_attempt|envoy_listener_manager_lds_update_attempt|envoy_listener_manager_listener_create_failure|istio_agent_pilot_eds_no_instances|envoy_cluster_http2_header_overflow|envoy_cluster_lb_healthy_panic|envoy_cluster_upstream_rq_retry_limit_exceeded|envoy_listener_manager_total_filter_chains_draining|envoy_cluster_http2_outbound_flood|envoy_cluster_upstream_rq_per_try_idle_timeout|envoy_listener_manager_lds_init_fetch_timeout|envoy_server_concurrency|envoy_server_stats_recent_lookups|envoy_server_version|envoy_cluster_membership_degraded|envoy_cluster_upstream_rq_max_duration_reached|envoy_wasm_envoy_wasm_runtime_null_active|istio_agent_dns_upstream_request_duration_seconds_sum|istio_agent_dns_upstream_requests_total|istio_agent_go_memstats_heap_sys_bytes|envoy_listener_manager_lds_update_time|envoy_server_worker_1_watchdog_miss|istio_agent_go_goroutines|istio_agent_go_memstats_stack_inuse_bytes|istio_agent_process_start_time_seconds|envoy_listener_manager_lds_version|envoy_wasm_envoy_wasm_runtime_null_created|envoy_server_memory_physical_size|envoy_cluster_upstream_rq_timeout|envoy_http_outbound_0_0_0_0_80_rbac|envoy_cluster_upstream_cx_overflow|envoy_cluster_upstream_flow_control_backed_up_total|envoy_server_dynamic_unknown_fields|envoy_server_initialization_time_ms_sum|istio_agent_go_memstats_mspan_sys_bytes|istio_agent_go_memstats_stack_sys_bytes|envoy_cluster_manager_cds_update_time|envoy_cluster_upstream_cx_close_notify|envoy_listener_manager_listener_in_place_updated|envoy_wasm_remote_load_fetch_failures|istio_response_bytes_sum|envoy_cluster_circuit_breakers_high_cx_open|envoy_cluster_http2_streams_active|envoy_server_worker_1_watchdog_mega_miss|istio_agent_go_memstats_gc_cpu_fraction|envoy_cluster_membership_healthy|envoy_server_worker_0_watchdog_mega_miss|envoy_cluster_lb_zone_number_differs|envoy_cluster_upstream_rq_pending_failure_eject|envoy_cluster_version|istio_agent_pilot_conflict_inbound_listener|envoy_cluster_bind_errors|envoy_cluster_http2_trailers|envoy_cluster_retry_or_shadow_abandoned|envoy_cluster_upstream_cx_connect_timeout|envoy_cluster_upstream_cx_tx_bytes_buffered|envoy_cluster_upstream_rq_completed|envoy_cluster_upstream_rq_pending_active|envoy_listener_manager_lds_update_failure|envoy_cluster_http2_stream_refused_errors|envoy_cluster_lb_zone_routing_all_directly|istio_agent_pilot_conflict_outbound_listener_http_over_current_tcp|istio_agent_pilot_virt_services|istio_agent_go_gc_duration_seconds|istio_agent_go_gc_duration_seconds_count
      replacement: $1
      sourceLabels:
      - __name__
    path: /stats/prometheus
    relabelings:
    - action: keep
      regex: istio-proxy
      sourceLabels:
      - __meta_kubernetes_pod_container_name
    - action: keep
      sourceLabels:
      - __meta_kubernetes_pod_annotationpresent_prometheus_io_scrape
    - action: replace
      regex: ([^:]+)(?::\d+)?;(\d+)
      replacement: $1:$2
      sourceLabels:
      - __address__
      - __meta_kubernetes_pod_annotation_prometheus_io_port
      targetLabel: __address__
  selector:
    matchExpressions:
    - key: istio-prometheus-ignore
      operator: DoesNotExist`

// ServiceMonitorTemplate ServiceMonitor资源模板，用于Istio控制面监控
const ServiceMonitorTemplate = `apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  labels:
    app: istiod
    created-by: bcs-mesh-manager
    monitoring: istio-control-plane
  name: %s
  namespace: istio-system
spec:
  endpoints:
  - interval: 30s
    path: /metrics
    port: http-monitoring
  namespaceSelector:
    matchNames:
    - istio-system
  selector:
    matchLabels:
      app: istiod`

// TelemetryTemplate Telemetry资源模板，用于Istio链路追踪
const TelemetryTemplate = `apiVersion: telemetry.istio.io/v1alpha1
kind: Telemetry
metadata:
  name: bcs-istio-tracing
  namespace: istio-system
spec:
  tracing:
  - providers:
    - name: otel-tracing
    randomSamplingPercentage: %d`

// GetPodMonitorYAML 获取PodMonitor YAML模板
func GetPodMonitorYAML(name string) string {
	return fmt.Sprintf(PodMonitorTemplate, name)
}

// GetServiceMonitorYAML 获取ServiceMonitor YAML模板
func GetServiceMonitorYAML(name string) string {
	return fmt.Sprintf(ServiceMonitorTemplate, name)
}

// GetTelemetryYAML 获取Telemetry YAML模板
func GetTelemetryYAML(samplePercentage int) string {
	return fmt.Sprintf(TelemetryTemplate, samplePercentage)
}

// MonitoringResourceNames 监控资源名称常量
const (
	// PodMonitorName PodMonitor资源名称
	PodMonitorName = "bcs-istio-proxy-metrics"
	// ServiceMonitorName ServiceMonitor资源名称
	ServiceMonitorName = "bcs-istio-control-plane-metrics"
	// TelemetryName Telemetry资源名称
	TelemetryName = "bcs-istio-tracing"
)
