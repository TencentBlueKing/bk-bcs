<template>
  <bcs-select
    :loading="nsLoading"
    multiple
    :popover-min-width="360"
    searchable
    :key="Date.now()"
    :clearable="false"
    :display-tag="displayTag"
    :placeholder="$t('view.labels.all')"
    selected-style="checkbox"
    v-model="nsData"
    @selected="handleNsChange">
    <!-- <bcs-option key="all" id="" :name="$t('view.labels.all')">
      <bcs-checkbox :value="nsData.includes('')">
        <span class="text-[12px]">{{ $t('view.labels.all') }}</span>
      </bcs-checkbox>
    </bcs-option> -->
    <bcs-option
      v-for="item in nsList"
      :key="item.name"
      :id="item.name"
      :name="item.name">
    </bcs-option>
    <template #extension>
      <SelectExtension
        :link-text="$t('dashboard.ns.create.title')"
        @link="handleGotoNs"
        @refresh="handleGetNsData" />
    </template>
  </bcs-select>
</template>
<script setup lang="ts">
import { isEqual } from 'lodash';
import { computed, ref, watch } from 'vue';

import SelectExtension from '@/components/select-extension.vue';
import { useCluster } from '@/composables/use-app';
import $router from '@/router';
import { INamespace, useNamespace } from '@/views/cluster-manage/namespace/use-namespace';

const props = defineProps({
  clusterId: {
    type: String,
    default: '',
  },
  value: {
    type: Array,
    default: () => [],
  },
  displayTag: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['change', 'input']);

const { getNamespaceData } = useNamespace();

const nsData = ref<string[]>([]);
const curNsData = computed(() => nsData.value.filter(item => !!item));// 过滤全部命名空间
watch(() => props.value, () => {
  if (isEqual(props.value, curNsData.value)) return;
  if (!props.value?.length) {
    nsData.value = [];// 全部命名空间逻辑
  } else {
    nsData.value = JSON.parse(JSON.stringify(props.value));
  }
}, { immediate: true });
watch(curNsData, (newValue, oldValue) => {
  if (isEqual(newValue, oldValue)) return;
  emits('change', curNsData.value);
  emits('input', curNsData.value);
});

const handleNsChange = (nsList) => {
  const last = nsList[nsList.length - 1];
  // 移除全选
  if (last) {
    nsData.value = nsData.value.filter(item => !!item);
  } else {
    nsData.value = [];
  }
};

// 组件使用的数据
const nsList = ref<Array<INamespace>>([]);
const nsLoading = ref(false);
const { clusterList } = useCluster();
const handleGetNsData = async () => {
  const exist = clusterList.value.find(item => item.clusterID === props.clusterId);
  if (!exist) return;
  nsLoading.value = true;
  nsList.value = await getNamespaceData({ $clusterId: props.clusterId });
  nsLoading.value = false;
};

// 跳转命名空间
const handleGotoNs = () => {
  const { href } = $router.resolve({
    name: 'createNamespace',
    params: {
      clusterId: props.clusterId,
    },
  });
  window.open(href);
};

watch(() => props.clusterId, () => {
  handleGetNsData();
}, { immediate: true });
</script>
