<template>
  <div>
    <div class="font-bold mb-[16px] text-[14px]">{{ $t('serviceMesh.label.info') }}</div>
    <div class="grid grid-cols-2 gap-4">
      <FieldItem :label="$t('serviceMesh.label.name')" :value="data?.name" />
      <FieldItem :label="$t('serviceMesh.label.ID')" :value="data?.meshID" />
      <FieldItem :label="$t('serviceMesh.label.version')" :value="data?.version" />
    </div>
    <FieldItem class="mt-[16px]" :label="$t('serviceMesh.label.description')" :value="data?.description" />
    <bcs-divider class="!mt-[24px] !mb-[8px]"></bcs-divider>
    <div class="font-bold mb-[16px] text-[14px]">{{ $t('serviceMesh.label.clusterConfig') }}</div>
    <FieldItem :label="$t('serviceMesh.label.primaryClusters')" :value="data?.primaryClusters?.join?.()" />
  </div>
</template>
<script setup lang="ts">
import FieldItem from './field-item.vue';

defineProps({
  data: {
    type: Object,
    default: () => ({}),
  },
});
</script>
