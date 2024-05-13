<template>
  <bk-form class="bcs-small-form">
    <!-- 微软云 -->
    <bk-form-item :label="$t('azureCloud.label.clusterType')">
      <span class="text-[#313238]">
        {{ $t('azureCloud.label.managed') }}
      </span>
      <span class="text-[#979BA5]">
        ({{ $t('azureCloud.label.aksDesc') }})
      </span>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'AzureMaster',
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

    // google cloud locationType
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
