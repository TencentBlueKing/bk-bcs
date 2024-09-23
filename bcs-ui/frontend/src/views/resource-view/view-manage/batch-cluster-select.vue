<template>
  <div class="w-[300px]">
    <bcs-input
      left-icon="bk-icon icon-search"
      behavior="simplicity"
      :placeholder="$t('view.placeholder.searchCluster')"
      class="mb-[4px]"
      clearable
      v-model.trim="searchKey">
    </bcs-input>
    <div class="max-h-[320px] overflow-auto">
      <bcs-checkbox-group class="!flex !flex-col" v-model="checkedClusters">
        <bcs-checkbox
          v-for="item, index in filterClusterList"
          :key="index"
          :value="item.clusterID"
          :disabled="checkedClusters.length === 1 && checkedClusters.includes(item.clusterID)"
          class="flex items-center flex-1 !bcs-dropdown-item">
          <span class="text-[12px] bcs-ellipsis" :title="item.clusterName">{{ item.clusterName }}</span>
        </bcs-checkbox>
      </bcs-checkbox-group>
    </div>
    <bk-exception
      type="search-empty"
      scene="part"
      v-if="searchKey && !filterClusterList.length"
      class="w-[300px]">
    </bk-exception>
    <SelectExtension
      class="!mx-[0px] !mt-[4px] !mb-[-4px] w-full bcs-border-top"
      :link-text="$t('view.button.addCluster')"
      :show-refresh="false"
      @link="handleGotoAddCluster" />
  </div>
</template>
<script setup lang="ts">
import { isEqual } from 'lodash';
import { computed, ref, watch } from 'vue';

import SelectExtension from '@/components/select-extension.vue';
import { useCluster } from '@/composables/use-app';
import $router from '@/router';

const props = defineProps({
  clusters: {
    type: Array,
    default: () => [],
  },
});
const emits = defineEmits(['change']);

const { clusterList } = useCluster();
const searchKey = ref('');
const checkedClusters = ref(props.clusters);

// 跳转添加集群入口
const handleGotoAddCluster = () => {
  const { href } = $router.resolve({ name: 'clusterCreate' });
  window.open(href);
};

const filterClusterList = computed(() => {
  const searchValue = searchKey.value.toLocaleLowerCase();
  return clusterList.value.filter((item) => {
    const clusterID = item.clusterID.toLocaleLowerCase();
    const clusterName = item.clusterName.toLocaleLowerCase();
    return clusterID?.includes(searchValue) || clusterName?.includes(searchValue);
  });
});

watch(() => props.clusters, () => {
  if (isEqual(props.clusters, checkedClusters.value)) {
    return;
  }
  checkedClusters.value = JSON.parse(JSON.stringify(props.clusters));
});

watch(checkedClusters, () => {
  emits('change', checkedClusters.value);
});
</script>
