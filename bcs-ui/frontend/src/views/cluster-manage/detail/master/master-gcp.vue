<template>
  <bk-form class="bcs-small-form">
    <!-- 谷歌云 -->
    <bk-form-item :label="$t('cluster.labels.clusterType')">
      <span class="text-[#313238]">
        {{ manageType === 'MANAGED_CLUSTER' ? $t('bcs.cluster.autopilot') : $t('bcs.cluster.standard') }}
      </span>
      <span v-if="manageType === 'MANAGED_CLUSTER'" class="text-[#979BA5]">
        ({{ $t('cluster.create.label.manageType.managed.gkeDesc') }})
      </span>
      <span v-if="manageType === 'INDEPENDENT_CLUSTER'" class="text-[#979BA5]">
        ({{ $t('cluster.create.google.tips.standard') }})
      </span>
    </bk-form-item>
    <bk-form-item
      :label="$t('cluster.create.label.manageType.managed.clusterLevel.text')">
      <div v-if="locationType === 'zones'">
        <span class="text-[#313238]">{{ $t('googleCloud.label.zoneCluster.title') }}</span>
        <span class="text-[#979BA5]">({{ $t('googleCloud.label.zoneCluster.desc') }})</span>
      </div>
      <div v-else-if="locationType === 'regions'">
        <span class="text-[#313238]">{{ $t('googleCloud.label.regionCluster.title') }}</span>
        <span class="text-[#979BA5]">({{ $t('googleCloud.label.regionCluster.desc') }})</span>
      </div>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { computed, defineComponent } from 'vue';

import { useCluster } from '@/composables/use-app';

export default defineComponent({
  name: 'GCPMaster',
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

    // google cloud manageType
    const manageType = computed(() => curCluster.value?.manageType);


    return {
      curCluster,
      locationType,
      manageType,
    };
  },
});
</script>
