<template>
  <div class="w-[320px]">
    <bcs-input
      behavior="simplicity"
      clearable
      right-icon="bk-icon icon-search"
      v-model="keyword"
      :placeholder="$t('输入集群名或ID搜索')">
    </bcs-input>
    <bcs-exception
      type="empty"
      scene="part"
      class="!w-[320px]"
      v-if="!clusterData.length">
    </bcs-exception>
    <div class="max-h-[400px] overflow-y-auto" v-else>
      <CollapseTitle
        :title="`${$t('独立集群')} (${independentClusterList.length})`"
        :collapse="independentCollapse"
        class="px-[16px] mt-[8px]"
        @click="independentCollapse = !independentCollapse" />
      <ul class="bg-[#fff] overflow-auto text-[12px] text-[#63656e]" v-show="!independentCollapse">
        <li
          v-for="item in independentClusterList"
          :key="item.clusterID"
          :class="[
            'px-[40px] cursor-pointer hover:bg-[#eaf3ff] hover:text-[#3a84ff]',
            {
              'bg-[#eaf3ff] text-[#3a84ff]': selectable && localValue === item.clusterID
            }
          ]"
          @click="handleClick(item.clusterID)">
          <div class="flex flex-col justify-center h-[50px]">
            <span class="bcs-ellipsis" v-bk-overflow-tips>{{ item.clusterName }}</span>
            <span class="mt-[8px] text-[#979BA5]">{{ item.clusterID }}</span>
          </div>
        </li>
      </ul>
      <template v-if="clusterType !== 'independent'">
        <CollapseTitle
          :title="`${$t('共享集群')} (${sharedClusterList.length})`"
          :collapse="sharedCollapse"
          class="px-[16px] mt-[4px]"
          @click="sharedCollapse = !sharedCollapse" />
        <ul class="bg-[#fff] overflow-auto text-[12px] text-[#63656e]" v-show="!sharedCollapse">
          <li
            v-for="item in sharedClusterList"
            :key="item.clusterID"
            :class="[
              'px-[40px] cursor-pointer hover:bg-[#eaf3ff] hover:text-[#3a84ff]',
              {
                'bg-[#eaf3ff] text-[#3a84ff]': selectable && localValue === item.clusterID
              }
            ]"
            @click="handleClick(item.clusterID)">
            <div class="flex flex-col justify-center h-[50px]">
              <span class="bcs-ellipsis" v-bk-overflow-tips>{{ item.clusterName }}</span>
              <span class="mt-[8px] text-[#979BA5]">{{ item.clusterID }}</span>
            </div>
          </li>
        </ul>
      </template>
    </div>
  </div>
</template>
<script lang="ts" setup>
// popover场景的集群选择器
import { PropType } from 'vue';
import CollapseTitle from './collapse-title.vue';
import useClusterSelector, { ClusterType } from './use-cluster-selector';

const props = defineProps({
  value: {
    type: String,
    default: '',
  },
  clusterType: {
    type: String as PropType<ClusterType>,
    default: 'independent', // 默认只展示独立集群
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
  clusterData,
  sharedClusterList,
  independentClusterList,
  sharedCollapse,
  independentCollapse,
  handleClusterChange,
} = useClusterSelector(emits, props.value, props.clusterType, props.updateStore);

const handleClick = (clusterID) => {
  handleClusterChange(clusterID);

  emits('click', clusterID);
};
</script>
