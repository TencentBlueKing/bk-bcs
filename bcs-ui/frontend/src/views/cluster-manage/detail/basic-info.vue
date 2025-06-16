<template>
  <div>
    <bk-form
      class="bcs-small-form grid grid-cols-2 grid-rows-[repeat(3,auto-fill] gap-[16px]"
      v-bkloading="{ isLoading }">
      <DescList class="row-span-3" :title="$t('cluster.title.clusterInfo')">
        <bk-form-item :label="$t('cluster.labels.Kubernetes')">
          <div class="flex items-center">
            <svg class="size-[20px] mr-[5px]" v-if="providerNameMap[clusterProvider]">
              <use :xlink:href="providerNameMap[clusterProvider]?.className"></use>
            </svg>
            <span>{{ providerNameMap[clusterProvider]?.label || '--' }}</span>
          </div>
        </bk-form-item>
        <bk-form-item :label="$t('tke.label.account')" v-if="clusterData.cloudAccountID">
          <LoadingIcon v-if="cloudAccountLoading">{{ $t('generic.status.loading') }}...</LoadingIcon>
          <span v-else>{{ cloudAccount }}</span>
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
            :disable-edit="!editable"
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
          <div class="flex items-center" v-if="!isEdit">
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
            <span
              class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
              v-if="editable"
              @click="isEdit = true">
              <i class="bk-icon icon-edit-line"></i>
            </span>
          </div>
          <template v-else>
            <div class="flex items-center">
              <KeyValue2
                :value="clusterData.labels || {}"
                :key-rules="[
                  {
                    message: $t('generic.validate.label'),
                    validator: LABEL_KEY_REGEXP,
                  }
                ]"
                :value-rules="[
                  {
                    message: $t('generic.validate.label'),
                    validator: LABEL_KEY_REGEXP,
                  }
                ]"
                :min-item="0"
                class="flex-1"
                @validate="setValidate"
                @change="setClusterLabel" />
              <div class="flex items-center h-[32px] ml-[16px]">
                <bk-button
                  class="text-[12px] leading-none"
                  text
                  @click="handleModifyLabel">{{ $t('generic.button.save') }}</bk-button>
                <bk-button
                  class="text-[12px] ml-[8px] leading-none"
                  text
                  @click="isEdit = false">{{ $t('generic.button.cancel') }}</bk-button>
              </div>
            </div>
          </template>
        </bk-form-item>
        <bk-form-item :label="$t('cluster.create.label.desc1')">
          <EditFormItem
            :maxlength="100"
            type="textarea"
            :value="clusterData.description"
            :placeholder="$t('cluster.placeholder.desc')"
            class="w-[300px]"
            :disable-edit="!editable"
            @save="handleClusterDescChange" />
        </bk-form-item>
        <bk-form-item :label="$t('cluster.labels.visibleRange')">
          <ClusterVisibleRange
            :editable="rangeEdit"
            :value="clusterData?.sharedRanges?.projectIdOrCodes"
            :is-shared="clusterData?.is_shared"
            :loading="isLoading"
            :cluster-id="clusterData.clusterID"
            :disable-edit="!editable"
            @edit="rangeEdit = true"
            @cancel="rangeEdit = false"
            @save="handleVisibleRangeChange" />
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
          <span>{{ clusterData.region }}</span>
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.nodemanArea')">
          <template v-if="isEditModule">
            <div class="flex items-center">
              <NodemanArea
                v-model="newCloudID"
                class="flex-1"
                @list-change="handleAreaListChange" />
              <span
                class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
                text
                @click="handleSaveNodemanArea">{{ $t('generic.button.save') }}</span>
              <span
                class="text-[12px] text-[#3a84ff] ml-[8px] cursor-pointer"
                text
                @click="isEditModule = false">{{ $t('generic.button.cancel') }}</span>
            </div>
          </template>
          <template v-else>
            <span>
              {{ nodemanArea || '--' }}
            </span>
            <span
              v-if="editable"
              class="hover:text-[#3a84ff] cursor-pointer ml-[8px]"
              @click="handleEditNodemanArea">
              <i class="bk-icon icon-edit-line"></i>
            </span>
          </template>
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
import { merge } from 'lodash';
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import { getClusterImportCategory, getClusterTypeName, useClusterInfo, useClusterList } from '../cluster/use-cluster';
import ClusterVisibleRange from '../components/cluster-visible-range.vue';
import EditFormItem from '../components/edit-form-item.vue';

import { modifyCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { CLUSTER_ENV, LABEL_KEY_REGEXP } from '@/common/constant';
import DescList from '@/components/desc-list.vue';
import KeyValue from '@/components/key-value.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import useSideslider from '@/composables/use-sideslider';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import NodemanArea from '@/views/cluster-manage/add/components/nodeman-area.vue';
import KeyValue2 from '@/views/cluster-manage/components/key-value.vue';
import { INodeManCloud } from '@/views/cluster-manage/types/types';
import useCloud from '@/views/cluster-manage/use-cloud';
import useVariable from '@/views/deploy-manage/variable/use-variable';

export default defineComponent({
  name: 'ClusterInfo',
  components: {
    EditFormItem,
    LoadingIcon,
    KeyValue,
    KeyValue2,
    DescList,
    NodemanArea,
    ClusterVisibleRange,
  },
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
    const clusterProvider = computed(() => {
      if (clusterData.value.clusterType === 'virtual') return clusterData.value?.extraInfo?.provider;
      return clusterData.value?.provider;
    });
    const clusterType = computed(() => getClusterTypeName(clusterData.value, isFromCurProject.value));
    // 是否可编辑
    const editable = computed(() => clusterData.value.status === 'RUNNING' && (!clusterData.value.is_shared || isFromCurProject.value));
    // 当前项目集群
    const isFromCurProject = computed(() => $store.state.curProject?.projectID === clusterData.value.projectID);
    // 修改集群信息
    const handleModifyCluster = async (params  = {}) => {
      isLoading.value = true;
      const result = await modifyCluster({
        $clusterId: clusterData.value?.clusterID,
        ...params,
      }).then(() => true)
        .catch(() => false);
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('generic.msg.success.modify'),
      });
      await Promise.all([
        getClusterList(),
        getClusterDetail(clusterId.value, true),
      ]);
      isLoading.value = false;
      return result;
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
    // 修改标签
    const isEdit = ref(false);
    const labels = ref({});
    const isLabelValidate = ref(true);
    const setValidate = (result: boolean) => {
      isLabelValidate.value = result;
    };
    const setClusterLabel = (data = {}) => {
      labels.value = data;
    };
    const handleModifyLabel = async () => {
      if (!isLabelValidate.value) return;

      if (JSON.stringify(labels) === JSON.stringify(clusterData.value.labels)) {
        isEdit.value = false;
        return;
      }
      await handleModifyCluster({ labels2: { values: labels.value } });
      isEdit.value = false;
    };
    // 修改可见范围
    const user = computed(() => $store.state.user);
    const rangeEdit = ref(false);
    async function handleVisibleRangeChange(val) {
      const result = await handleModifyCluster({
        is_shared: !val?.isOnlyCurrentPorject,
        sharedRanges: !val?.isOnlyCurrentPorject && !val?.isAll ? {
          projectIdOrCodes: val?.value,
        } : {},
        updator: user.value.username,
      });
      result && (rangeEdit.value = false);
    };

    // 集群kubeconfig
    const handleGotoToken = () => {
      const { href } = $router.resolve({ name: 'token' });
      window.open(href);
    };

    const {
      cloudAccountLoading,
      cloudAccountList,
      cloudAccounts,
      nodemanCloudList,
      handleGetNodeManCloud,
      providerNameMap,
    } = useCloud();

    const cloudAccount = computed(() => {
      const data = cloudAccountList.value.find(item => item?.account?.accountID === clusterData.value.cloudAccountID);
      return data ? `${data.account?.accountName}(${data.account.accountID})` : clusterData.value.cloudAccountID;
    });

    // 修改云区域
    const newCloudID = ref<number>();
    const isEditModule = ref(false);
    const nodemanArea = computed(() => {
      const data = nodemanCloudList.value
        .find(item => item.bk_cloud_id === clusterData.value.clusterBasicSettings?.area?.bkCloudID);
      return data ? `${data.bk_cloud_name}(${data.bk_cloud_id})` : clusterData.value.clusterBasicSettings?.area?.bkCloudID;
    });
    const handleAreaListChange = (data: INodeManCloud[]) => {
      nodemanCloudList.value = data;
    };
    const handleEditNodemanArea = () => {
      newCloudID.value = clusterData.value?.clusterBasicSettings?.area?.bkCloudID;
      isEditModule.value = true;
    };
    const handleSaveNodemanArea = async () => {
      if (newCloudID.value === clusterData.value?.clusterBasicSettings?.area) return;

      const result = await handleModifyCluster({
        clusterBasicSettings: merge(
          clusterData.value.clusterBasicSettings,
          {
            area: {
              bkCloudID: newCloudID.value,
            },
          },
        ),
      });

      if (result) {
        isEditModule.value = false;
      }
    };

    onMounted(async () => {
      getClusterVar();
      isLoading.value = true;
      await getClusterDetail(clusterId.value, curCluster.value?.clusterType !== 'virtual');
      isLoading.value = false;
      cloudAccounts(clusterData.value.provider);
      handleGetNodeManCloud();
    });
    return {
      cloudAccountLoading,
      cloudAccount,
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
      LABEL_KEY_REGEXP,
      isEdit,
      setValidate,
      setClusterLabel,
      handleModifyLabel,
      isEditModule,
      nodemanArea,
      newCloudID,
      handleSaveNodemanArea,
      handleEditNodemanArea,
      handleAreaListChange,
      providerNameMap,
      rangeEdit,
      handleVisibleRangeChange,
      editable,
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
