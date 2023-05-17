<template>
  <div>
    <bcs-alert type="warning" :title="$t('您需要先删除以下所有Release, 再进行删除操作')" class="mb-[10px]"></bcs-alert>
    <bcs-table :data="data">
      <bcs-table-column :label="$t('所属集群')" prop="clusterID"></bcs-table-column>
      <bcs-table-column :label="$t('命名空间')" prop="namespace"></bcs-table-column>
      <bcs-table-column :label="$t('名称')" prop="name"></bcs-table-column>
      <!-- <bcs-table-column :label="$t('操作')" width="80">
        <template #default="{ row }">
          <bcs-button text @click="handleGotoChartRelease(row)">{{$t('查看')}}</bcs-button>
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
