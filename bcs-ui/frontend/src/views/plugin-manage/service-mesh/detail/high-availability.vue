<template>
  <div class="text-[14px]">
    <div class="font-bold mb-[16px]">{{ $t('serviceMesh.label.highAvailability') }}</div>
    <FieldItem
      class="mb-[16px]"
      :label="$t('serviceMesh.label.replicas')"
      :value="data?.highAvailability?.replicaCount" />
    <FieldItem class="pb-[8px]" :label="$t('serviceMesh.label.resource')">
      <div class="grid grid-cols-4 bg-[#F5F7FA] rounded-sm py-[12px] px-[16px]">
        <FieldItem
          :label="$t('serviceMesh.label.cpuRequest')"
          :value="data?.highAvailability?.resourceConfig?.cpuRequest">
        </FieldItem>
        <FieldItem
          :label="$t('serviceMesh.label.cpuLimit')"
          :value="data?.highAvailability?.resourceConfig?.cpuLimit">
        </FieldItem>
        <FieldItem
          :label="$t('serviceMesh.label.memoryRequest')"
          :value="data?.highAvailability?.resourceConfig?.memoryRequest">
        </FieldItem>
        <FieldItem
          :label="$t('serviceMesh.label.memoryLimit')"
          :value="data?.highAvailability?.resourceConfig?.memoryLimit">
        </FieldItem>
      </div>
    </FieldItem>
    <!-- HPA -->
    <ContentSwitcher
      v-if="data?.highAvailability?.autoscaleEnabled"
      class="mt-[16px]"
      readonly
      :label="'HPA'">
      <div class="grid grid-cols-3 gap-1">
        <FieldItem
          :label="$t('serviceMesh.label.cpuPercent')"
          :value="data?.highAvailability?.targetCPUAverageUtilizationPercent" />
        <FieldItem
          :label="$t('serviceMesh.label.replicaMin')"
          :value="data?.highAvailability?.autoscaleMin" />
        <FieldItem
          :label="$t('serviceMesh.label.replicaMax')"
          :value="data?.highAvailability?.autoscaleMax" />
      </div>
    </ContentSwitcher>
    <!-- 专属节点 -->
    <ContentSwitcher
      v-if="data?.highAvailability?.dedicatedNode?.enabled"
      class="mt-[16px]"
      readonly
      :label="$t('serviceMesh.label.dedicatedNode')">
      <FieldItem :label="$t('generic.label.tag')">
        <div
          class="leading-[20px]"
          v-for="(value, key) of (data?.highAvailability?.dedicatedNode?.nodeLabels || {})"
          :key="`${key}-${value}`">
          {{ key }}: {{ value }}
        </div>
      </FieldItem>
    </ContentSwitcher>
  </div>
</template>
<script setup lang="ts">
import ContentSwitcher from '../content-switcher.vue';

import FieldItem from './field-item.vue';

defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
});
</script>
