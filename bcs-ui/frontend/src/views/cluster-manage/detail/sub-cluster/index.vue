<template>
  <div class="h-full">
    <!-- 操作栏 -->
    <div class="flex items-center mb-[16px]">
      <bcs-button
        theme="primary"
        icon="plus"
        class="add-node mr10"
        @click="handleAddSubCluster">
        {{$t('cluster.button.addCluster')}}
      </bcs-button>
      <bk-input
        right-icon="bk-icon icon-search"
        class="flex-1"
        :placeholder="$t('cluster.placeholder.searchCluster')"
        v-model.trim="searchValue"
        clearable>
      </bk-input>
      <div
        class="flex items-center justify-center w-[32px] h-[32px] text-[12px] cursor-pointer ml-[10px] bcs-border"
        @click="getDataWithLoading()">
        <i class="bcs-icon bcs-icon-reset"></i>
      </div>
    </div>
    <!-- 表格 -->
    <ListMode
      v-bkloading="{ isLoading: loading }"
      :cluster-list="curClusterList"
      :overview="overview"
      :perms="perms"
      :search-value="searchValue"
      :cluster-nodes-map="clusterNodesMap"
      :status-text-map="statusTextMap"
      :status-color-map="statusColorMap"
      :failed-status-list="failedStatusList"
      :support-detail-status-list="supportDetailStatusList"
      :loading-status-list="failedStatusList"
      class="h-[calc(100%-48px)]"
      @clear="searchValue = ''"
      @active-row="handleChangeActiveRow">
      <template #action="{ row }">
        <bk-button
          :disabled="!supportDetailStatusList.includes(row.status)"
          text
          @click.stop="handleRemove(row)">
          {{$t('generic.button.remove')}}
        </bk-button>
      </template>
    </ListMode>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, set, watch } from 'vue';

import ListMode from '../../cluster/cluster-list.vue';

import { useFederation } from './use-federation';

import { clusterMeta, deleteFederalCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import useSearch from '@/composables/use-search';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import { useClusterOverview } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  name: 'SubCluster',
  components: { ListMode },
  props: {
    clusterId: {
      type: String,
      required: true,
    },
    perms: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    // 异常状态
    const failedStatusList = ['Creating', 'Deleting'];
    // 支持详情页展示的状态
    const supportDetailStatusList = ['RUNNING', 'CreateFailed'];

    const statusTextMap = ref({
      Deleting: $i18n.t('generic.status.deleting'),
      CreateFailed: $i18n.t('generic.status.createFailed'),
      RUNNING: $i18n.t('generic.status.ready'),
      Creating: $i18n.t('generic.status.creating'),
    });

    const statusColorMap = ref({
      CreateFailed: 'red',
      RUNNING: 'green',
    });

    const { curPageData, loading, curClusterId, getFederationCluster } = useFederation();
    const keys = ref(['name', 'clusterID']);
    const { searchValue, tableDataMatchSearch: curClusterList } = useSearch(curPageData, keys);

    // 联邦集群
    function handleAddSubCluster() {
      $router.push({
        name: 'addSubCluster',
        params: {
          clusterId: props.clusterId,
        },
      });
    };

    // 集群节点数
    const clusterNodesMap = ref<Record<string, number>>({});
    const clusterIDs = ref<string[]>([]);
    const handleGetClusterNodes = async () => {
      clusterIDs.value = curPageData.value.map(item => item.clusterID);
      if (!clusterIDs.value.length) return;

      clusterNodesMap.value = {};
      const data = await clusterMeta({
        clusters: clusterIDs.value,
      }).catch(() => []);
      data.forEach((item) => {
        set(clusterNodesMap.value, item.clusterId, item.clusterNodeNum);
      });
    };

    // 集群指标
    const { clusterOverviewMap: overview } = useClusterOverview(curClusterList);

    function handleChangeActiveRow(clusterID) {
      ctx.emit('active-row', clusterID);
    }

    const user = computed(() => $store.state.user);
    function handleRemove(row) {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('cluster.create.federation.title'),
        subTitle: `${row.clusterName}(${row.clusterID})`,
        defaultInfo: true,
        confirmLoading: true,
        confirmFn: async () => {
          const result = await deleteFederalCluster({
            $fedClusterId: row.federation_cluster_id,
            $subClusterId: row.sub_cluster_id,
            user: user.value.username,
          }).catch(() => false);
          if (result) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.deliveryTask'),
            });
            getFederationCluster();
          }
        },
      });
    }

    // 刷新
    async function getDataWithLoading() {
      loading.value = true;
      await getFederationCluster();
      loading.value = false;
    }

    watch(curPageData, async () => {
      // 列表数据变化时，重新获取集群节点数
      if (JSON.stringify(clusterIDs.value) !== JSON.stringify(curPageData.value.map(item => item.clusterID))) {
        await handleGetClusterNodes();
      }
    });

    onMounted(async () => {
      curClusterId.value = props.clusterId;
      await getDataWithLoading();
    });

    return {
      searchValue,
      curClusterList,
      supportDetailStatusList,
      statusTextMap,
      statusColorMap,
      failedStatusList,
      clusterNodesMap,
      overview,
      loading,
      handleAddSubCluster,
      handleChangeActiveRow,
      handleRemove,
      getDataWithLoading,
    };
  },
});
</script>
