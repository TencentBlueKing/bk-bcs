<template>
  <div>
    <bk-form
      class="bcs-small-form grid grid-cols-2 grid-rows-[repeat(3,auto-fill] gap-[16px]"
      v-bkloading="{ isLoading }">
      <DescList class="row-span-3" :title="$t('cluster.title.clusterInfo')">
        <bk-form-item :label="$t('cluster.labels.Kubernetes')">
          {{ clusterProvider || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('tke.label.account')" v-if="clusterData.cloudAccountID">
          {{ clusterData.cloudAccountID }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.clusterId')">
          {{ clusterData.clusterID }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.name')">
          <EditFormItem
            :maxlength="64"
            :value="clusterData.clusterName"
            :placeholder="$t('cluster.create.validate.name')"
            class="w-[300px]"
            @save="handleClusterNameChange" />
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.clusterType')">
          {{ clusterType }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.create.label.clusterVersion')">
          {{ clusterVersion }}
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.detail.label.tkeID')"
          v-if="clusterData.providerType === 'tke' && clusterData.clusterType !== 'virtual'">
          {{ clusterData.systemID || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.clusterLabel')">
          <div class="flex h-[32px] items-center" v-if="Object.keys(clusterData.labels || {}).length">
            <bcs-tag
              v-for="key, index in Object.keys(clusterData.labels || {})"
              :key="index"
              class="bcs-ellipsis"
              v-bk-overflow-tips>
              {{ `${key}=${clusterData.labels[key]}` }}
            </bcs-tag>
          </div>
          <span v-else>--</span>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.create.label.desc1')">
          <EditFormItem
            :maxlength="100"
            type="textarea"
            :value="clusterData.description"
            :placeholder="$t('cluster.placeholder.desc')"
            class="w-[300px]"
            @save="handleClusterDescChange" />
        </bk-form-item>
      </DescList>
      <DescList :title="$t('cluster.title.clusterConfig')">
        <template v-if="clusterData.clusterType !== 'virtual'">
          <bk-form-item :label="$t('cluster.ca.nodePool.create.containerRuntime.title')">
            {{ containerRuntime }}
          </bk-form-item>
          <bk-form-item :label="$t('cluster.ca.nodePool.create.runtimeVersion.title')">
            {{ runtimeVersion }}
          </bk-form-item>
        </template>
        <bk-form-item label="KubeConfig">
          <bcs-button
            text
            class="text-[12px]"
            @click="handleGotoToken">
            {{ $t('cluster.nodeTemplate.sops.status.running.detailBtn') }}
          </bcs-button>
        </bk-form-item>
        <bk-form-item :label="$t('deploy.variable.clusterEnv')" v-if="$INTERNAL">
          <LoadingIcon v-if="varLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
          <template v-else>
            <bcs-button
              text
              v-if="variableList.length"
              class="text-[12px]"
              @click="handleShowVarSideslider">
              {{ `${variableList.length} ${$t('units.suffix.units')}` }}
            </bcs-button>
            <span v-else>--</span>
          </template>
        </bk-form-item>
      </DescList>
      <DescList :title="$t('cluster.labels.env')">
        <bk-form-item :label="$t('cluster.labels.region')">
          {{ region }}
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.nodemanArea')"
          v-if="clusterData.clusterBasicSettings && clusterData.clusterBasicSettings.area">
          {{ clusterData.clusterBasicSettings && clusterData.clusterBasicSettings.area
            ? clusterData.clusterBasicSettings.area.bkCloudID
            : '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.env')">
          {{ CLUSTER_ENV[clusterData.environment] }}
        </bk-form-item>
      </DescList>
      <DescList :title="$t('cluster.title.operationRecord')">
        <bk-form-item :label="$t('cluster.labels.importCategory')">
          {{ getClusterImportCategory(clusterData) }}
        </bk-form-item>
        <bk-form-item :label="$t('generic.label.createdBy')">
          {{ clusterData.creator || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.createdAt')">
          {{ clusterData.createTime || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.updatedAt')">
          {{ clusterData.updateTime || '--' }}
        </bk-form-item>
      </DescList>
      <bk-sideslider
        :is-show.sync="showVariableInfo"
        :title="`${$t('cluster.title.setClusterVar')} ( ${clusterId} )`"
        :width="680"
        :before-close="handleBeforeClose"
        quick-close
        @hidden="showVariableInfo = false">
        <template #content>
          <div class="p-[40px] pt-[20px]" v-bkloading="{ isLoading: varLoading }">
            <KeyValue
              :model-value="variableList"
              :show-operate="false"
              :show-header="false"
              @data-change="setChanged(true)"
              @confirm="handleSaveVar"
              @cancel="handleCancelSetVar" />
          </div>
        </template>
      </bk-sideslider>
    </bk-form>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import { getClusterImportCategory, getClusterTypeName, useClusterInfo, useClusterList } from '../cluster/use-cluster';
import EditFormItem from '../components/edit-form-item.vue';

import $bkMessage from '@/common/bkmagic';
import { CLUSTER_ENV } from '@/common/constant';
import DescList from '@/components/desc-list.vue';
import KeyValue from '@/components/key-value.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import useSideslider from '@/composables/use-sideslider';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import useVariable from '@/views/deploy-manage/variable/use-variable';

export default defineComponent({
  name: 'ClusterInfo',
  components: { EditFormItem, LoadingIcon, KeyValue, DescList },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
  },
  setup(props) {
    const { clusterId } = toRefs(props);

    // 集群变量
    const showVariableInfo = ref(false);
    const varLoading = ref(false);
    const variableList = ref<any[]>([]);
    const { handleGetClusterVariables, handleUpdateSpecifyClusterVariables } = useVariable();
    const getClusterVar = async () => {
      varLoading.value = true;
      const data = await handleGetClusterVariables({ $clusterId: clusterId.value });
      variableList.value = data.results;
      reset();
      varLoading.value = false;
    };
    const handleShowVarSideslider = () => {
      showVariableInfo.value = true;
      getClusterVar();
    };
    const handleSaveVar = async (_, data) => {
      const newVarList = variableList.value.map((item) => {
        const value = data?.find(i => i.id === item.id)?.value;
        return {
          ...item,
          value,
        };
      });
      varLoading.value = true;
      const result = await handleUpdateSpecifyClusterVariables({
        $clusterId: clusterId.value,
        data: newVarList,
      });
      varLoading.value = false;
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.save'),
        });
        showVariableInfo.value = false;
      }
    };
    const handleCancelSetVar = () => {
      showVariableInfo.value = false;
    };
    // 抽屉关闭校验
    const { handleBeforeClose, reset, setChanged } = useSideslider(variableList);

    // 集群信息
    const { getClusterList, clusterList } = useClusterList();
    const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterId.value));
    const { clusterData, isLoading, getClusterDetail } = useClusterInfo();
    const clusterVersion = computed(() => clusterData.value?.clusterBasicSettings?.version || '--');
    const runtimeVersion = computed(() => clusterData.value?.clusterAdvanceSettings?.runtimeVersion || '--');
    const containerRuntime = computed(() => clusterData.value?.clusterAdvanceSettings?.containerRuntime || '--');
    const providerMap = {
      tencentPublicCloud: $i18n.t('provider.tencentPublicCloud'),
    };
    const clusterProvider = computed(() => {
      if (clusterData.value.clusterType === 'virtual') return clusterData.value?.extraInfo?.provider;
      return providerMap[clusterData.value?.provider] || clusterData.value?.provider;
    });
    const clusterType = computed(() => getClusterTypeName(clusterData.value));
    // 修改集群信息
    const handleModifyCluster = async (params  = {}) => {
      isLoading.value = true;
      const result = await $store.dispatch('clustermanager/modifyCluster', {
        $clusterId: clusterData.value?.clusterID,
        ...params,
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.modify'),
      });
      await Promise.all([
        getClusterList(),
        getClusterDetail(clusterId.value, true),
      ]);
      isLoading.value = false;
    };
    // 修改集群名称
    const handleClusterNameChange = async (clusterName) => {
      if (!clusterName.trim()) return;
      handleModifyCluster({ clusterName: clusterName.trim() });
    };
    // 修改集群描述
    const handleClusterDescChange = async (description) => {
      handleModifyCluster({ description });
    };
    // 集群kubeconfig
    const handleGotoToken = () => {
      const { href } = $router.resolve({ name: 'token' });
      window.open(href);
    };
    // 地域信息
    const regionList = ref<any[]>([]);
    const region = computed(() => regionList.value
      .find(item => item.region === clusterData.value.region)?.regionName || clusterData.value.region);
    const getRegionList = async () => {
      if (!['tencentCloud', 'tencentPublicCloud'].includes(curCluster.value?.provider || '')) return;
      regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
        $cloudId: curCluster.value?.provider,
      });
    };

    onMounted(async () => {
      getClusterVar();
      isLoading.value = true;
      await Promise.all([
        getRegionList(),
        getClusterDetail(clusterId.value, curCluster.value?.clusterType !== 'virtual'),
      ]);
      isLoading.value = false;
    });
    return {
      region,
      clusterVersion,
      showVariableInfo,
      varLoading,
      isLoading,
      clusterData,
      runtimeVersion,
      containerRuntime,
      variableList,
      handleClusterNameChange,
      handleClusterDescChange,
      handleShowVarSideslider,
      handleSaveVar,
      handleCancelSetVar,
      handleGotoToken,
      setChanged,
      handleBeforeClose,
      CLUSTER_ENV,
      clusterType,
      clusterProvider,
      getClusterImportCategory,
    };
  },
});
</script>
<style lang="postcss" scoped>
.bcs-small-form {
  &-item {
    margin-top: 0;
    font-size: 12px;
    width: 100%;
  }
  >>> .bk-label {
    font-size: 12px;
    line-height: 32px;
    padding-right: 8px;
    &::after {
      content: ':';
      margin-left: 4px;
    }
  }
  >>> .bk-form-item {
    margin: 0;
  }
  >>> .bk-form-content {
    line-height: 32px;
    font-size: 12px;
    color: #313238;
  }
}
</style>
