<template>
  <bcs-select
    class="cluster-select"
    v-model="localValue"
    :clearable="false"
    :searchable="searchable"
    :disabled="disabled"
    :popover-min-width="320"
    :remote-method="remoteMethod"
    :search-placeholder="$t('输入集群名或ID搜索')"
    @change="handleClusterChange">
    <!-- 独立集群 -->
    <bcs-option-group
      :name="$t('独立集群')"
      :is-collapse="independentCollapse"
      :class="['mt-[8px]', { 'mb-[4px]': clusterType === 'independent' }]">
      <template #group-name>
        <CollapseTitle
          :title="`${$t('独立集群')} (${independentClusterList.length})`"
          :collapse="independentCollapse"
          @click="independentCollapse = !independentCollapse" />
      </template>
      <bcs-option
        v-for="item in independentClusterList"
        :key="item.clusterID"
        :id="item.clusterID"
        :name="item.clusterName">
        <div class="flex flex-col justify-center h-[50px] px-[12px]">
          <span class="leading-6 bcs-ellipsis" v-bk-overflow-tips>{{ item.clusterName }}</span>
          <span class="leading-4 text-[#979BA5]">{{ item.clusterID }}</span>
        </div>
      </bcs-option>
    </bcs-option-group>
    <!-- 共享集群 -->
    <bcs-option-group
      :name="$t('共享集群')"
      :is-collapse="sharedCollapse"
      class="mb-[8px]"
      v-if="clusterType !== 'independent'">
      <template #group-name>
        <CollapseTitle
          :title="`${$t('共享集群')} (${sharedClusterList.length})`"
          :collapse="sharedCollapse"
          @click="sharedCollapse = !sharedCollapse" />
      </template>
      <bcs-option
        v-for="item in sharedClusterList"
        :key="item.clusterID"
        :id="item.clusterID"
        :name="item.clusterName">
        <div class="flex flex-col justify-center h-[50px] px-[12px]">
          <span class="leading-none bcs-ellipsis" v-bk-overflow-tips>{{ item.clusterName }}</span>
          <span class="leading-none mt-[8px] text-[#979BA5]">{{ item.clusterID }}</span>
        </div>
      </bcs-option>
    </bcs-option-group>
  </bcs-select>
</template>
<script lang="ts">
import {  defineComponent, watch, toRefs, PropType } from 'vue';
import CollapseTitle from './collapse-title.vue';
import useClusterSelector, { ClusterType } from './use-cluster-selector';

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
      type: String as PropType<ClusterType>,
      default: 'independent', // 默认只展示独立集群
    },
  },
  emits: ['change'],
  setup(props, ctx) {
    const { value, clusterType } = toRefs(props);

    const {
      localValue,
      keyword,
      sharedClusterList,
      independentClusterList,
      sharedCollapse,
      independentCollapse,
      handleClusterChange,
    } = useClusterSelector(ctx.emit, value.value, clusterType.value);

    watch(value, (v) => {
      localValue.value = v;
    });

    // 远程搜索
    const remoteMethod = (searhcKey) => {
      keyword.value = searhcKey;
    };

    return {
      localValue,
      sharedClusterList,
      independentClusterList,
      sharedCollapse,
      independentCollapse,
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
