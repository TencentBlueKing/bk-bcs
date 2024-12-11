<template>
  <BcsContent :title="$t('projects.project.quota')" hide-back>
    <bk-table
      :data="data"
      v-bkloading="{ isLoading: isLoading }"
      class="network-table">
      <bk-table-column :label="$t('projects.project.label.region')" prop="quota.zoneResources.region" />
      <bk-table-column :label="$t('projects.project.label.zone')" prop="quota.zoneResources.zoneName" />
      <bk-table-column :label="$t('projects.project.label.instance')" prop="quota.zoneResources.instanceType" />
      <bk-table-column :label="$t('projects.project.label.Num')" prop="quota.zoneResources.quotaNum" />
      <bk-table-column :label="$t('projects.project.label.used')" prop="quota.zoneResources.quotaUsed" />
      <bk-table-column :label="$t('projects.project.label.available')">
        <template #default="{ row }">
          <span>{{ row.quota?.zoneResources?.quotaNum - row.quota?.zoneResources?.quotaUsed }}</span>
        </template>
      </bk-table-column>
      <!-- <bk-table-column :label="$t('projects.project.label.AssociatedNodePool')">
        <template #default="{ row }">
          <div
            v-for="item in row.nodeGroups"
            :key="item.nodeGroupId"
            class="bk-primary bk-button-normal bk-button-text"
            @click="handleGotoDetail(item)">
            {{ item.nodeGroupId }}
          </div>
          <span v-if="!row.nodeGroups?.length">--</span>
        </template>
      </bk-table-column> -->
    </bk-table>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, onBeforeMount, ref } from 'vue';

import { fetchProjectQuotas } from '@/api/modules/project';
import BcsContent from '@/components/layout/Content.vue';
import { useProject } from '@/composables/use-app';
import $router from '@/router';

export default defineComponent({
  name: 'ProjectQuotas',
  components: { BcsContent },
  setup() {
    const { curProject } = useProject();

    const data = ref([]);
    const isLoading = ref(false);
    // 获取项目配额
    async function handleGetProjectQuotas() {
      if (!curProject.value.projectID) return;

      isLoading.value = true;
      const res = await fetchProjectQuotas({
        projectID: curProject.value.projectID,
        provider: 'selfProvisionCloud',
      }).catch(() => ({ results: [] }));
      data.value = res?.results || [];
      isLoading.value = false;
    }

    // 节点池详情
    const handleGotoDetail = (nodePool) => {
      $router.push({
        name: 'nodePoolDetail',
        params: {
          clusterId: nodePool.clusterId,
          nodeGroupID: nodePool.nodeGroupId,
        },
      }).catch((err) => {
        console.warn(err);
      });
    };

    onBeforeMount(() => {
      handleGetProjectQuotas();
    });

    return {
      isLoading,
      data,
      handleGotoDetail,
    };
  },
});
</script>
