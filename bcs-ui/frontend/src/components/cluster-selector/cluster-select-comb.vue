<template>
  <div class="flex items-center">
    <ClusterSelect
      class="mr-[5px]"
      :value="clusterId"
      :cluster-type="clusterType"
      v-show="showClusterSelect"
      @change="handleClusterChange">
    </ClusterSelect>
    <bcs-input
      right-icon="bk-icon icon-search"
      class="mr-[5px] w-[320px]"
      :placeholder="placeholder"
      :value="search"
      clearable
      v-show="showSearch"
      @change="handleSearchChange"
      @enter="handleEnter">
    </bcs-input>
    <div
      class="refresh"
      @click="handleRefresh">
      <i class="bcs-icon bcs-icon-reset"></i>
    </div>
  </div>
</template>
<script lang="ts">
import { PropType, defineComponent } from 'vue';
import ClusterSelect from './cluster-select.vue';
import { ClusterType } from './use-cluster-selector';

export default defineComponent({
  name: 'ClusterSelectComb',
  components: { ClusterSelect },
  props: {
    showClusterSelect: {
      type: Boolean,
      default: true,
    },
    showSearch: {
      type: Boolean,
      default: true,
    },
    placeholder: String,
    clusterId: String,
    search: String,
    clusterType: {
      type: String as PropType<ClusterType>,
      default: 'independent',
    },
  },
  emits: ['cluster-change', 'search-change', 'search-enter', 'refresh', 'update:clusterId', 'update:search'],
  setup(props, ctx) {
    const handleClusterChange = (clusterID) => {
      ctx.emit('update:clusterId', clusterID);
      ctx.emit('cluster-change', clusterID);
    };
    const handleSearchChange = (value) => {
      ctx.emit('update:search', value);
      ctx.emit('search-change', value);
    };
    const handleEnter = (value) => {
      ctx.emit('search-enter', value);
    };

    const handleRefresh = () => {
      ctx.emit('refresh');
    };
    return {
      handleClusterChange,
      handleRefresh,
      handleSearchChange,
      handleEnter,
    };
  },
});
</script>
<style lang="postcss" scoped>
.refresh {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 1px solid #c4c6cc;
  font-size: 14px;
  background-color: #fff;
  cursor: pointer;
}
</style>
