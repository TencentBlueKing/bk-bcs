<template>
  <div class="text-[14px]">
    <div class="font-bold mb-[16px]">{{ $t('serviceMesh.label.featureConfig') }}</div>
    <FieldItem
      class="mb-[16px]"
      :label="$t('serviceMesh.label.egressMode')"
      :value="data?.featureConfigs?.outboundTrafficPolicy?.value" />
    <div class="grid grid-cols-4 mb-[16px]">
      <FieldItem :label="$t('serviceMesh.label.sidecarOn')">
        <bk-tag
          theme="success"
          v-if="data?.featureConfigs?.holdApplicationUntilProxyStarts?.value === 'true'">
          {{ $t('serviceMesh.status.on') }}</bk-tag>
        <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
      </FieldItem>
      <FieldItem :label="$t('serviceMesh.label.sidecarOff')">
        <bk-tag
          theme="success"
          v-if="data?.featureConfigs?.exitOnZeroActiveConnections?.value === 'true'">
          {{ $t('serviceMesh.status.on') }}</bk-tag>
        <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
      </FieldItem>
      <FieldItem :label="$t('serviceMesh.label.dnsCapture')">
        <bk-tag
          theme="success"
          v-if="data?.featureConfigs?.istioMetaDnsCapture?.value === 'true'">
          {{ $t('serviceMesh.status.on') }}</bk-tag>
        <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
      </FieldItem>
      <FieldItem :label="$t('serviceMesh.label.dnsAutoAllocate')">
        <bk-tag
          theme="success"
          v-if="data?.featureConfigs?.istioMetaDnsAutoAllocate?.value === 'true'">
          {{ $t('serviceMesh.status.on') }}</bk-tag>
        <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
      </FieldItem>
    </div>
    <FieldItem class="mb-[16px]" :label="$t('serviceMesh.label.sidecarResource')">
      <div class="grid grid-cols-4 bg-[#F5F7FA] rounded-sm py-[12px] px-[16px]">
        <FieldItem :label="$t('serviceMesh.label.cpuRequest')" :value="data?.sidecarResourceConfig?.cpuRequest">
        </FieldItem>
        <FieldItem :label="$t('serviceMesh.label.cpuLimit')" :value="data?.sidecarResourceConfig?.cpuLimit">
        </FieldItem>
        <FieldItem :label="$t('serviceMesh.label.memoryRequest')" :value="data?.sidecarResourceConfig?.memoryRequest">
        </FieldItem>
        <FieldItem :label="$t('serviceMesh.label.memoryLimit')" :value="data?.sidecarResourceConfig?.memoryLimit">
        </FieldItem>
      </div>
    </FieldItem>
    <FieldItem
      class="mb-[16px]"
      :label="$t('serviceMesh.label.excludeIPs')"
      :value="data?.featureConfigs?.excludeIPRanges?.value">
    </FieldItem>
    <FieldItem :label="$t('serviceMesh.label.http10')">
      <bk-tag
        theme="success"
        v-if="data?.featureConfigs?.istioMetaHttp10?.value === 'true'">
        {{ $t('serviceMesh.status.on') }}</bk-tag>
      <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
    </FieldItem>
    <bcs-divider class="!mt-[24px] !mb-[8px]"></bcs-divider>
    <div
      v-if="showObserveLabel"
      class="font-bold mb-[16px] text-[14px]">
      {{ $t('serviceMesh.label.observability') }}</div>
    <!-- 指标采集 -->
    <ContentSwitcher
      v-if="data?.observabilityConfig?.metricsConfig?.metricsEnabled"
      class="mt-[16px]"
      readonly
      :label="$t('serviceMesh.label.metrics')">
      <div class="grid grid-cols-2">
        <FieldItem :label="$t('serviceMesh.label.controlPlane')">
          <bk-tag
            theme="success"
            v-if="data?.observabilityConfig?.metricsConfig?.controlPlaneMetricsEnabled">
            {{ $t('serviceMesh.status.on') }}</bk-tag>
          <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
        </FieldItem>
        <FieldItem :label="$t('serviceMesh.label.dataPlane')">
          <bk-tag
            theme="success"
            v-if="data?.observabilityConfig?.metricsConfig?.dataPlaneMetricsEnabled">
            {{ $t('serviceMesh.status.on') }}</bk-tag>
          <bk-tag v-else>{{ $t('serviceMesh.status.off') }}</bk-tag>
        </FieldItem>
      </div>
    </ContentSwitcher>
    <!-- 日志输出 -->
    <ContentSwitcher
      v-if="data?.observabilityConfig?.logCollectorConfig?.enabled"
      class="mt-[16px]"
      readonly
      :label="$t('serviceMesh.label.logOutput')">
      <FieldItem
        class="mb-[16px]"
        :label="$t('serviceMesh.label.logType')"
        :value="data?.observabilityConfig?.logCollectorConfig?.accessLogEncoding" />
      <FieldItem
        :label="$t('serviceMesh.label.logFormat')"
        :value="data?.observabilityConfig?.logCollectorConfig?.accessLogFormat" />
    </ContentSwitcher>
    <!-- 全链路追踪 -->
    <ContentSwitcher
      v-if="data?.observabilityConfig?.tracingConfig?.enabled"
      class="mt-[16px]"
      readonly
      :label="$t('serviceMesh.label.tracing')">
      <FieldItem
        class="mb-[16px]"
        :label="$t('serviceMesh.label.sampleRate')"
        :value="data?.observabilityConfig?.tracingConfig?.traceSamplingPercent" />
      <FieldItem
        class="mb-[16px]"
        :label="$t('serviceMesh.label.endpoint')"
        :value="data?.observabilityConfig?.tracingConfig?.endpoint" />
      <FieldItem
        class="mb-[16px]"
        :label="$t('serviceMesh.label.bkToken')"
        :value="data?.observabilityConfig?.tracingConfig?.bkToken" />
    </ContentSwitcher>
  </div>
</template>
<script setup lang="ts">
import { computed, PropType } from 'vue';

import ContentSwitcher from '../content-switcher.vue';
import { IMesh } from '../use-mesh';

import FieldItem from './field-item.vue';

const props = defineProps({
  data: {
    type: Object as PropType<IMesh>,
    default: () => ({}),
  },
});

const showObserveLabel = computed(() => {
  const { metricsConfig, logCollectorConfig, tracingConfig } = props.data?.observabilityConfig || {};
  return metricsConfig?.metricsEnabled || logCollectorConfig?.enabled || tracingConfig?.enabled;
});
</script>
<style lang="postcss" scoped>
:deep(.bk-tag) {
  margin-left: 0;
}
</style>
