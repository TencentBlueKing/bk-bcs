<template>
  <BcsContent :title="$t('事件查询')" hide-back>
    <EventQueryTable :namespace="namespace" :cluster-id="clusterId"></EventQueryTable>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, ref } from 'vue';
import BcsContent from '@/components/layout/Content.vue';
import EventQueryTable from './event-query-table.vue';
import $store from '@/store';

export default defineComponent({
  name: 'EventQuery',
  components: { BcsContent, EventQueryTable },
  setup() {
    const namespace = ref($store.state.curNamespace);
    const clusterId = ref($store.state.curCluster?.clusterID);

    return {
      namespace,
      clusterId,
    };
  },
});
</script>
