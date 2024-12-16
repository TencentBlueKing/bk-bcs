<template>
  <div class="flex items-center">
    <bk-popover
      class="size-[16px] mr-[4px]"
      :content="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconText"
      :tippy-options="{ interactive: false }">
      <i
        class="text-[14px] p-[1px]"
        :class="sourceTypeMap?.[handleGetExtData(row.metadata.uid, 'createSource')]?.iconClass"></i>
    </bk-popover>
    <span
      v-bk-overflow-tips="{ interactive: false }"
      class="bcs-ellipsis" v-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Template'">
      {{ `${handleGetExtData(row.metadata.uid, 'templateName') || '--'}:${
        handleGetExtData(row.metadata.uid, 'templateVersion') || '--'}` }}
    </span>
    <span
      v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
      v-else-if="handleGetExtData(row.metadata.uid, 'createSource') === 'Helm'">
      {{ handleGetExtData(row.metadata.uid, 'chart')
        ?`${handleGetExtData(row.metadata.uid, 'chart') || '--'}`
        : 'Helm' }}
    </span>
    <span
      v-bk-overflow-tips="{ interactive: false }" class="bcs-ellipsis"
      v-else>{{ handleGetExtData(row.metadata.uid, 'createSource') }}</span>
  </div>
</template>
<script lang="ts">
import { defineComponent, inject } from 'vue';

export default defineComponent({
  name: 'SourceTableCell',
  props: {
    row: {
      type: Object,
      default: () => ({}),
      required: true,
    },
    sourceTypeMap: {
      type: Object,
      default: () => ({}),
    },
  },
  setup() {
    // 接收handleGetExtData方法
    const handleGetExtData = inject('handleGetExtData') as (uid: string, ext?: string, defaultData?: any) => any;

    return {
      handleGetExtData,
    };
  },
});
</script>
