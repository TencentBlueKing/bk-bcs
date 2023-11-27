<template>
  <div class="w-[320px]">
    <bcs-input
      behavior="simplicity"
      clearable
      right-icon="bk-icon icon-search"
      v-model="keyword"
      :placeholder="$t('cluster.placeholder.searchCluster')">
    </bcs-input>
    <bcs-exception
      type="empty"
      scene="part"
      class="!w-[320px]"
      v-if="isClusterDataEmpty">
    </bcs-exception>
    <div class="max-h-[400px] overflow-y-auto" v-else>
      <div v-for="item, index in clusterData" :key="item.type">
        <CollapseTitle
          :title="`${item.title} (${item.list.length})`"
          :collapse="collapseList.includes(item.type)"
          :class="[
            'px-[16px]',
            index === 0 ? 'mt-[8px]' : 'mt-[4px]'
          ]"
          @click="handleToggleCollapse(item.type)" />
        <ul
          class="bg-[#fff] overflow-auto text-[12px] text-[#63656e]"
          v-show="!collapseList.includes(item.type)">
          <li
            v-for="cluster in item.list"
            :key="cluster.clusterID"
            :class="[
              'px-[40px] hover:bg-[#eaf3ff]',
              {
                'bg-[#eaf3ff] text-[#3a84ff]': selectable && localValue === cluster.clusterID,
                'hover:text-[#3a84ff] cursor-pointer': cluster.status === 'RUNNING',
                'text-[#c4c6cc] cursor-not-allowed': cluster.status !== 'RUNNING'
              }
            ]"
            @click="handleClick(cluster)">
            <div class="flex flex-col justify-center h-[50px]" v-bk-tooltips="{
              content: $t('cluster.tips.clusterStatus', [CLUSTER_MAP[cluster.status || '']]),
              disabled: cluster.status === 'RUNNING',
              placement: 'right'
            }">
              <span class="bcs-ellipsis" v-bk-overflow-tips>{{ cluster.clusterName }}</span>
              <span
                :class="['mt-[8px]', { 'text-[#979BA5]': cluster.status === 'RUNNING' }]">
                {{ cluster.clusterID }}
              </span>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
// popover场景的集群选择器
import { PropType } from 'vue';

import CollapseTitle from './collapse-title.vue';
import useClusterSelector, { ClusterType } from './use-cluster-selector';

import { CLUSTER_MAP } from '@/common/constant';

const props = defineProps({
  value: {
    type: String,
    default: '',
  },
  clusterType: {
    type: [String, Array] as PropType<ClusterType|ClusterType[]>,
    default: () => ['independent', 'managed'], // 默认只展示独立集群和托管集群
  },
  updateStore: {
    type: Boolean,
    default: true,
  },
  selectable: {
    type: Boolean,
    default: true,
  },
});

const emits = defineEmits(['change', 'click']);

const {
  keyword,
  localValue,
  collapseList,
  isClusterDataEmpty,
  clusterData,
  handleToggleCollapse,
  handleClusterChange,
} = useClusterSelector(emits, props.value, props.clusterType, props.updateStore);

const handleClick = (cluster) => {
  if (cluster.status !== 'RUNNING') return;

  handleClusterChange(cluster.clusterID);

  emits('click', cluster.clusterID);
};
</script>
