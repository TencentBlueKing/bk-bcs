<template>
  <BcsContent :title="$t('事件查询')" hide-back>
    <EventQueryTable></EventQueryTable>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent } from '@vue/composition-api';
import BcsContent from '@/components/layout/Content.vue';
import EventQueryTable from './event-query-table.vue';

export default defineComponent({
  name: 'EventQuery',
  components: { BcsContent, EventQueryTable },
  props: {

  },
  setup() {
    return {};
  },
});
</script>
