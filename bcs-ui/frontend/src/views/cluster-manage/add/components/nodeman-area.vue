<template>
  <bk-select
    searchable
    :disabled="disabled"
    :clearable="false"
    :loading="nodemanCloudLoading"
    :value="value"
    class="bg-[#fff]"
    @change="handleValueChange">
    <bk-option
      v-for="item in nodemanCloudList"
      :key="item.bk_cloud_id"
      :id="item.bk_cloud_id"
      :name="item.bk_cloud_name">
    </bk-option>
    <template slot="extension">
      <SelectExtension
        :link-text="$t('tke.link.nodeman')"
        :link="`${PROJECT_CONFIG.nodemanHost}/#/cloud-manager`"
        @refresh="handleGetNodeManCloud" />
    </template>
  </bk-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';

import { nodemanCloud } from '@/api/modules/cluster-manager';
import SelectExtension from '@/components/select-extension.vue';
import { INodeManCloud } from '@/views/cluster-manage/types/types';

defineProps({
  value: {
    type: Number,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['input', 'change', 'list-change']);

// 管控区域
const nodemanCloudList = ref<Array<INodeManCloud>>([]);
const nodemanCloudLoading = ref(false);
const handleGetNodeManCloud = async () => {
  nodemanCloudLoading.value = true;
  nodemanCloudList.value = await nodemanCloud().catch(() => []);
  emits('list-change', nodemanCloudList.value);
  nodemanCloudLoading.value = false;
};

const handleValueChange = (v: number) => {
  emits('input', v);
  emits('change', v);
};

onBeforeMount(() => {
  handleGetNodeManCloud();
});
</script>
