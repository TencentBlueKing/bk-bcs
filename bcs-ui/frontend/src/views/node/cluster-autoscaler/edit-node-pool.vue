<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList"></HeaderNav>
    </template>
    <div v-bkloading="{ isLoading }">
      <bcs-tab class="node-pool-tab mb-[42px]">
        <bcs-tab-panel :label="$t('节点池信息')" name="basic">
          <NodePoolInfo
            :default-values="detailData"
            :schema="schema"
            :cluster="curCluster"
            is-edit
            :show-footer="false"
            v-if="!isLoading"
            ref="nodePoolInfoRef">
          </NodePoolInfo>
        </bcs-tab-panel>
        <bcs-tab-panel :label="$t('节点配置')" name="config">
          <NodeConfig
            :default-values="detailData"
            :schema="schema"
            :cluster="curCluster"
            is-edit
            ref="nodePoolConfigRef"
            v-if="!isLoading">
          </NodeConfig>
        </bcs-tab-panel>
      </bcs-tab>
      <div class="footer">
        <bcs-button
          theme="primary"
          style="min-width: 88px"
          :loading="saveLoading"
          @click="handleEditNodePool">{{$t('保存')}}</bcs-button>
        <bcs-button @click="handleCancel">{{$t('取消')}}</bcs-button>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { defineComponent, ref, computed, onMounted, toRefs } from '@vue/composition-api';
import BcsContent from '../bcs-content.vue';
import HeaderNav from '../header-nav.vue';
import { useClusterList } from '@/views/cluster/use-cluster';
import $i18n from '@/i18n/i18n-setup';
import NodePoolInfo from './node-pool-info.vue';
import NodeConfig from './node-config.vue';
import $store from '@/store/index';
import $router from '@/router/index';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    NodePoolInfo,
    NodeConfig,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    nodeGroupID: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { clusterId, nodeGroupID } = toRefs(props);

    const detailData = ref<any>(null);
    const nodePoolInfoRef = ref<any>(null);
    const nodePoolConfigRef = ref<any>(null);
    const { clusterList } = useClusterList(ctx);
    const curCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === clusterId.value) || {});
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === clusterId.value)?.clusterName,
        link: {
          name: 'clusterDetail',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterDetail',
          query: {
            active: 'AutoScaler',
          },
        },
      },
      {
        title: `${detailData.value?.nodeGroupID} (${detailData.value?.name}) `,
        link: {
          name: 'nodePoolDetail',
          params: {
            clusterId: props.clusterId,
            nodeGroupID: props.nodeGroupID,
          },
        },
      },
      {
        title: $i18n.t('编辑节点池'),
        link: null,
      },
    ]);

    const isLoading = ref(true);
    // 获取默认值
    const schema = ref({});
    const handleGetCloudDefaultValues = async () => {
      const data = await $store.dispatch('clustermanager/resourceSchema', {
        $cloudID: 'selfProvisionCloud',
        $name: 'nodegroup',
      });
      schema.value = data?.schema || {};
    };

    // 获取详情
    const handleGetNodeGroupDetail = async () => {
      const data = await $store.dispatch('clustermanager/nodeGroupDetail', {
        $nodeGroupID: nodeGroupID.value,
      });
      return data;
    };

    // 保存
    const user = computed(() => $store.state.user);
    const saveLoading = ref(false);
    const handleEditNodePool = async () => {
      const validate = await nodePoolInfoRef.value?.validate();
      if (!validate) return;

      saveLoading.value = true;
      const nodePoolData = nodePoolInfoRef.value?.getNodePoolData();
      const nodeConfigData = nodePoolConfigRef.value?.getNodePoolData();
      const data = {
        ...nodePoolData,
        ...nodeConfigData,
        $nodeGroupID: detailData.value.nodeGroupID,
        clusterID: curCluster.value.clusterID,
        region: curCluster.value.region,
        updater: user.value.username,
      };
      console.log(data);
      const result = await $store.dispatch('clustermanager/updateNodeGroup', data);
      saveLoading.value = false;
      if (result) {
        $router.push({
          name: 'clusterDetail',
          query: {
            active: 'AutoScaler',
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    onMounted(async () => {
      isLoading.value = true;
      await handleGetCloudDefaultValues();
      detailData.value = await handleGetNodeGroupDetail();
      isLoading.value = false;
    });
    return {
      saveLoading,
      curCluster,
      isLoading,
      schema,
      detailData,
      navList,
      nodePoolInfoRef,
      handleCancel,
      handleEditNodePool,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .node-pool-wrapper {
  .bk-resize-layout-main .main {
    max-height: calc(100vh - 270px);
  }
  .aside .content-wrapper {
    max-height: calc(100vh - 322px);
  }
}

>>> .node-pool-tab {
  .bk-tab-section {
    padding: 0;
  }
  .node-config {
    margin-bottom: 0;
  }
}
.footer {
  position: fixed;
  bottom: 0px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: start;
  padding: 0 24px;
  background-color: #fff;
  border-top: 1px solid #dcdee5;
  box-shadow: 0 -2px 4px 0 rgb(0 0 0 / 5%);
  z-index: 200;
  right: 0;
  width: calc(100% - 261px);
  .btn {
      width: 88px;
  }
}
</style>
