<template>
  <bcs-select
    class="cluster-select"
    v-model="localValue"
    :clearable="false"
    :searchable="searchable"
    :disabled="disabled"
    :popover-min-width="320"
    :remote-method="remoteMethod"
    :search-placeholder="$t('cluster.placeholder.searchCluster')"
    :size="size"
    :scroll-height="460"
    @change="handleClusterChange">
    <bcs-option-group
      v-for="item, index in clusterData"
      :key="item.type"
      :name="item.title"
      :is-collapse="collapseList.includes(item.type)"
      :class="[
        'mt-[8px]',
        index === (clusterData.length - 1) ? 'mb-[4px]' : ''
      ]">
      <template #group-name>
        <CollapseTitle
          :title="`${item.title} (${item.list.length})`"
          :collapse="collapseList.includes(item.type)"
          @click="handleToggleCollapse(item.type)" />
      </template>
      <bcs-option
        v-for="cluster in item.list"
        :key="cluster.clusterID"
        :id="cluster.clusterID"
        :name="cluster.clusterName"
        :disabled="cluster.status && !normalStatusList.includes(cluster.status)">
        <div class="flex items-center justify-between">
          <div
            class="flex-1 flex flex-col justify-center h-[50px] px-[12px]"
            @mouseenter="hoverClusterID = cluster.clusterID"
            @mouseleave="hoverClusterID = ''">
            <span class="leading-[20px] bcs-ellipsis" v-bk-overflow-tips>{{ cluster.clusterName }}</span>
            <span
              :class="[
                'leading-[20px]',
                {
                  'text-[#979BA5]': normalStatusList.includes(cluster.status || ''),
                  '!text-[#699DF4]': normalStatusList.includes(cluster.status || '')
                    && (hoverClusterID === cluster.clusterID || localValue === cluster.clusterID)
                }]">
              ({{ cluster.clusterID }})
            </span>
          </div>
          <bcs-tag
            theme="danger"
            v-if="!normalStatusList.includes(cluster.status || '')">
            {{ $t('generic.label.abnormal') }}
          </bcs-tag>
        </div>
      </bcs-option>
    </bcs-option-group>
  </bcs-select>
</template>
<script lang="ts">
import {  defineComponent, PropType, ref, toRefs, watch } from 'vue';

import CollapseTitle from './collapse-title.vue';
import useClusterSelector, { ClusterType } from './use-cluster-selector';

import { CLUSTER_MAP } from '@/common/constant';

export default defineComponent({
  name: 'ClusterSelect',
  components: { CollapseTitle },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    searchable: {
      type: Boolean,
      default: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    clusterType: {
      type: [String, Array] as PropType<ClusterType|ClusterType[]>,
      default: () => ['independent', 'managed'],
    },
    size: {
      type: String,
      default: '',
    },
    validateClusterId: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['change'],
  setup(props, ctx) {
    const { value, clusterType, validateClusterId } = toRefs(props);

    const normalStatusList = ['RUNNING'];
    const hoverClusterID = ref<string>();

    const {
      localValue,
      keyword,
      collapseList,
      clusterData,
      handleToggleCollapse,
      handleClusterChange,
    } = useClusterSelector(ctx.emit, value.value, clusterType.value, true, validateClusterId.value);

    watch(value, (v) => {
      localValue.value = v;
    });

    // 远程搜索
    const remoteMethod = (searhcKey) => {
      keyword.value = searhcKey;
    };

    return {
      localValue,
      collapseList,
      clusterData,
      CLUSTER_MAP,
      normalStatusList,
      hoverClusterID,
      handleToggleCollapse,
      remoteMethod,
      handleClusterChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-select {
    width: 254px;
    &:not(.is-disabled) {
      background-color: #fff;
    }
}
/deep/ .bk-option-group-name {
  border-bottom: 0 !important;
}
.bk-options .bk-option:first-child {
  margin-top: 0;
}
.bk-options .bk-option:last-child {
  margin-bottom: 0;
}
</style>
