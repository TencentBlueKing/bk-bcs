<template>
  <bk-form class="bcs-small-form">
    <!-- 华为云 -->
    <bk-form-item :label="$t('huaweiCloud.label.clusterType')">
      <span class="text-[#313238]">
        {{ $t('huaweiCloud.label.managed') }}
      </span>
      <span class="text-[#979BA5]">
        ({{ $t('huaweiCloud.label.aksDesc') }})
      </span>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'HuaweiMaster',
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId) || {});

    // huawei cloud locationType
    const locationType = computed(() => curCluster.value?.extraInfo?.locationType);


    return {
      curCluster,
      locationType,
    };
  },
});
</script>
<style lang="postcss" scoped>
@import './form.css';
</style>
