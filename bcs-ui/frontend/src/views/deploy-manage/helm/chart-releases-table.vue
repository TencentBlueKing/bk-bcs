<template>
  <div>
    <bcs-alert type="warning" :title="$t('deploy.helm.deleteWarning')" class="mb-[10px]"></bcs-alert>
    <bcs-table :data="data">
      <bcs-table-column :label="$t('generic.label.cluster1')" prop="clusterID"></bcs-table-column>
      <bcs-table-column :label="$t('k8s.namespace')" prop="namespace"></bcs-table-column>
      <bcs-table-column :label="$t('generic.label.name')" prop="name"></bcs-table-column>
      <!-- <bcs-table-column :label="$t('generic.label.action')" width="80">
        <template #default="{ row }">
          <bcs-button text @click="handleGotoChartRelease(row)">{{$t('generic.button.view')}}</bcs-button>
        </template>
      </bcs-table-column> -->
    </bcs-table>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue';
import $router from '@/router';

export default defineComponent({
  name: 'ChartReleasesTable',
  props: {
    data: {
      type: Array,
      default: () => [],
    },
  },
  setup() {
    const handleGotoChartRelease = (row) => {
      const { href } = $router.resolve({
        name: 'releaseList',
        query: {
          name: row.name,
          namespace: row.namespace,
        },
      });
      window.open(href);
    };

    return {
      handleGotoChartRelease,
    };
  },
});
</script>
