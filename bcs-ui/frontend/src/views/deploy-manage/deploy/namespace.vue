<template>
  <bcs-select :loading="loading" searchable :clearable="false" :scroll-height="460" @change="handleClusterNsChange">
    <bcs-option-group
      v-for="cluster, index in clusterList"
      :key="cluster.clusterID"
      :name="cluster.clusterID"
      :is-collapse="collapseList.includes(cluster.clusterID)"
      :class="[
        'bcs-select-group mt-[8px]',
        index === (clusterList.length - 1) ? 'mb-[4px]' : ''
      ]">
      <template #group-name>
        <CollapseTitle
          :title="`${cluster.clusterName} (${clusterNsList[cluster.clusterID]?.length})`"
          :collapse="collapseList.includes(cluster.clusterID)"
          @click="handleToggleCollapse(cluster.clusterID)" />
      </template>
      <bcs-option
        v-for="ns in (clusterNsList[cluster.clusterID] || [])"
        :key="`${cluster.clusterID}/${ns.name}`"
        :id="`${cluster.clusterID}/${ns.name}`"
        :name="`${cluster.clusterName} / ${ns.name}`"
        class="!mt-[0px]">
        <div class="flex flex-col justify-center px-[12px]">
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ ns.name }}</span>
        </div>
      </bcs-option>
    </bcs-option-group>
  </bcs-select>
</template>
<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';

import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import { useCluster } from '@/composables/use-app';
import { INamespace, useNamespace } from '@/views/cluster-manage/namespace/use-namespace';

type Emits = (e: 'change', clusterID: string, ns: string) => void;

const emits = defineEmits<Emits>();

const { clusterList } = useCluster();
const { getNamespaceData } = useNamespace();

const loading = ref(false);

// 折叠集群分组
const collapseList = ref<string[]>([]);
function handleToggleCollapse(clusterID: string) {
  const index = collapseList.value.findIndex(id => id === clusterID);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(clusterID);
  }
}

// 获取命名空间
const clusterNsList = ref<Record<string, INamespace[]>>({});
async function getClusterNsList() {
  const list = clusterList.value.map(cluster => getNamespaceData({ $clusterId: cluster.clusterID }).then((data) => {
    clusterNsList.value[cluster.clusterID] = data;
  }));
  loading.value = true;
  await Promise.all(list);
  loading.value = false;
}

// 命名空间变更
function handleClusterNsChange(data) {
  const [clusterID, ns] = data?.split('/');
  emits('change', clusterID, ns);
}

onBeforeMount(() => {
  getClusterNsList();
});
</script>
