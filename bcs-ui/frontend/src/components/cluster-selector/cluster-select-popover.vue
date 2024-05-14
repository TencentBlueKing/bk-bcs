<template>
  <div class="w-[320px]">
    <div class="px-[8px]">
      <bcs-input
        behavior="simplicity"
        clearable
        right-icon="bk-icon icon-search"
        v-model="keyword"
        :placeholder="$t('cluster.placeholder.searchCluster')">
      </bcs-input>
    </div>
    <bcs-exception
      type="empty"
      scene="part"
      class="!w-[320px]"
      v-if="isClusterDataEmpty">
    </bcs-exception>
    <div class="max-h-[460px] overflow-y-auto" v-else>
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
              'flex items-center justify-between px-[40px]',
              normalStatusList.includes(cluster.status || '')
                ? 'hover:bg-[#eaf3ff] hover:text-[#3a84ff] cursor-pointer'
                : 'text-[#c4c6cc] cursor-not-allowed',
              {
                'bg-[#eaf3ff] text-[#3a84ff]': selectable && localValue === cluster.clusterID,
              }
            ]"
            @mouseenter="hoverClusterID = cluster.clusterID"
            @mouseleave="hoverClusterID = ''"
            @click="handleClick(cluster)">
            <div class="flex-1 flex flex-col justify-center h-[50px]">
              <span class="bcs-ellipsis leading-[20px]" v-bk-overflow-tips>{{ cluster.clusterName }}</span>
              <span
                :class="[
                  'leading-[20px]',
                  {
                    'text-[#979BA5]': normalStatusList.includes(cluster.status || ''),
                    '!text-[#699DF4]': (normalStatusList.includes(cluster.status || '')
                      && hoverClusterID === cluster.clusterID)
                      || (selectable && localValue === cluster.clusterID)
                  }
                ]">
                ({{ cluster.clusterID }})
              </span>
            </div>
            <bcs-tag
              theme="danger"
              v-if="!normalStatusList.includes(cluster.status || '')">
              {{ $t('generic.label.abnormal') }}
            </bcs-tag>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
// popover场景的集群选择器
import { PropType, ref } from 'vue';

import CollapseTitle from './collapse-title.vue';
import useClusterSelector, { ClusterType } from './use-cluster-selector';

// import { CLUSTER_MAP } from '@/common/constant';

const props = defineProps({
  value: {
    type: String,
    default: '',
  },
  // 需要展示的集群类型
  clusterType: {
    type: [String, Array] as PropType<ClusterType|ClusterType[]>,
    default: () => ['independent', 'managed'], // 默认只展示独立集群和托管集群
  },
  // 切换集群时是否更新全局缓存
  updateStore: {
    type: Boolean,
    default: true,
  },
  // 是否展示选中的样式
  selectable: {
    type: Boolean,
    default: true,
  },
});

const emits = defineEmits(['change', 'click']);

const normalStatusList = ['RUNNING'];

const hoverClusterID = ref<string>();
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
  if (!normalStatusList.includes(cluster.status)) return;

  handleClusterChange(cluster.clusterID);

  emits('click', cluster.clusterID);
};
</script>
