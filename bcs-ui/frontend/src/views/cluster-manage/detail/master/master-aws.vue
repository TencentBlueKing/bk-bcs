<template>
  <bk-form class="bcs-small-form">
    <bk-form-item :label="$t('amazonCloud.label.clusterType')">
      <span class="text-[#313238]">
        {{ $t('amazonCloud.label.managed') }}
      </span>
      <span class="text-[#979BA5]">
        ({{ $t('amazonCloud.label.aksDesc') }})
      </span>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'AwsMaster',
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterList } = useCluster();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === props.clusterId));

    // google cloud locationType
    const locationType = computed(() => curCluster.value?.extraInfo?.locationType);


    return {
      curCluster,
      locationType,
    };
  },
});
</script>
