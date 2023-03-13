<template>
  <bcs-select
    class="cluster-select"
    v-model="localValue"
    :clearable="false"
    :searchable="searchable"
    :disabled="disabled"
    @change="handleClusterChange">
    <bcs-option
      v-for="item in clusterData"
      :key="item.clusterID"
      :id="item.clusterID"
      :name="item.clusterName">
      <div class="flex flex-col justify-center h-[46px]">
        <span class="leading-none bcs-ellipsis">{{ item.clusterName }}</span>
        <span class="leading-none mt-[8px]">{{ item.clusterID }}</span>
      </div>
    </bcs-option>
  </bcs-select>
</template>
<script lang="ts">
import { computed, defineComponent, ref, watch, toRefs, onBeforeMount } from '@vue/composition-api';
import $store from '@/store';
import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'ClusterSelect',
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
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    clusterType: {
      type: String,
      default: 'normal',
    },
  },
  emits: ['change'],
  setup(props, ctx) {
    const { value, clusterType } = toRefs(props);

    watch(value, (v) => {
      localValue.value = v;
    });

    const { clusterList } = useCluster();
    const clusterData = computed(() => {
      if (clusterType.value === 'normal') {
        return clusterList.value.filter(item => !item.is_shared);
      }
      return clusterList.value;
    });
    const localValue = ref<string>(props.value || $store.getters.curClusterId);

    const handleClusterChange = (clusterId) => {
      localValue.value = clusterId;
      $store.commit('updateCurCluster', clusterData.value.find(item => item.clusterID === clusterId));
      ctx.emit('change', clusterId);
    };

    onBeforeMount(() => {
      const data = clusterData.value.find(item => item.clusterID === localValue.value);
      if (!data) {
        handleClusterChange(clusterData.value[0]?.clusterID);
      } else if (localValue.value !== props.value) {
        handleClusterChange(localValue.value);
      }
    });

    return {
      localValue,
      clusterData,
      handleClusterChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.cluster-select {
    min-width: 254px;
    max-width: 600px;
    &:not(.is-disabled) {
      background-color: #fff;
    }
}
</style>
