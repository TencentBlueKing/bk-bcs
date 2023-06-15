<template>
  <div>
    <bk-form class="bcs-small-form px-[60px] py-[24px]" v-bkloading="{ isLoading }">
      <!-- <bk-form-item :label="$t('添加方式')">
      {{ clusterData.clusterCategory === 'importer' ? $t('导入集群') : $t('创建集群') }}
    </bk-form-item> -->
      <DescList :title="$t('集群信息')">
        <bk-form-item :label="$t('集群ID')">
          {{ clusterData.clusterID }}
        </bk-form-item>
        <bk-form-item :label="$t('集群名称')">
          <EditFormItem
            :maxlength="64"
            :value="clusterData.clusterName"
            :placeholder="$t('请输入集群名称，不超过64个字符')"
            class="w-[300px]"
            @save="handleClusterNameChange" />
        </bk-form-item>
        <bk-form-item :label="$t('集群类型')">
          {{ clusterType }}
        </bk-form-item>
        <bk-form-item :label="$t('集群版本')">
          {{ clusterVersion }}
        </bk-form-item>
        <bk-form-item
          :label="$t('TKE集群ID')"
          v-if="clusterData.providerType === 'tke' && clusterData.clusterType !== 'virtual'">
          {{ clusterData.systemID || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('集群描述')">
          <EditFormItem
            :maxlength="100"
            type="textarea"
            :value="clusterData.description"
            :placeholder="$t('请输入集群描述')"
            class="w-[300px]"
            @save="handleClusterDescChange" />
        </bk-form-item>
      </DescList>
      <DescList :title="$t('集群配置')">
        <template v-if="clusterData.clusterType !== 'virtual'">
          <bk-form-item :label="$t('运行时组件')">
            {{ containerRuntime }}
          </bk-form-item>
          <bk-form-item :label="$t('运行时版本')">
            {{ runtimeVersion }}
          </bk-form-item>
        </template>
        <bk-form-item label="KubeConfig">
          <bcs-button text class="text-[12px]" @click="handleGotoToken">{{ $t('查看详情') }}</bcs-button>
        </bk-form-item>
        <bk-form-item :label="$t('集群变量')" v-if="$INTERNAL">
          <LoadingIcon v-if="varLoading">{{ $t('加载中') }}...</LoadingIcon>
          <template v-else>
            <bcs-button
              text
              v-if="variableList.length"
              class="text-[12px]"
              @click="handleShowVarSideslider">
              {{ `${variableList.length} ${$t('个')}` }}
            </bcs-button>
            <span v-else>--</span>
          </template>
        </bk-form-item>
      </DescList>
      <DescList :title="$t('集群环境')">
        <bk-form-item :label="$t('所属地域')">
          {{ region }}
        </bk-form-item>
        <bk-form-item :label="$t('Kubernetes 提供商')">
          {{ clusterProvider || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('集群环境')">
          {{ CLUSTER_ENV[clusterData.environment] }}
        </bk-form-item>
      </DescList>
      <DescList :title="$t('状态和记录')">
        <bk-form-item :label="$t('状态')">
          <StatusIcon
            :status-color-map="{
              'CREATE-FAILURE': 'red',
              'DELETE-FAILURE': 'red',
              'IMPORT-FAILURE': 'red',
              RUNNING: 'green'
            }"
            :status-text-map="{
              INITIALIZATION: $t('正在初始化中，请稍等···'),
              DELETING: $t('正在删除中，请稍等···'),
              'CREATE-FAILURE': $t('创建失败，请重试'),
              'DELETE-FAILURE': $t('删除失败，请重试'),
              'IMPORT-FAILURE': $t('导入失败'),
              RUNNING: $t('正常')
            }"
            :status="clusterData.status"
            :pending="['INITIALIZATION', 'DELETING'].includes(clusterData.status)"
            v-if="clusterData.status" />
          <span v-else>--</span>
        </bk-form-item>
        <bk-form-item :label="$t('创建人')">
          {{ clusterData.creator || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('创建时间')">
          {{ clusterData.createTime || '--' }}
        </bk-form-item>
        <bk-form-item :label="$t('更新时间')">
          {{ clusterData.updateTime || '--' }}
        </bk-form-item>
      </DescList>
      <bk-sideslider
        :is-show.sync="showVariableInfo"
        :title="`${$t('设置变量')} ( ${clusterId} )`"
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
    <!-- <bk-form-item :label="$t('集群添加方式')"></bk-form-item>
    <bk-form-item :label="$t('调度引擎')"></bk-form-item> -->
    </bk-form>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, toRefs, ref } from 'vue';
import StatusIcon from '@/components/status-icon';
import useVariable from '@/views/deploy-manage/variable/use-variable';
import EditFormItem from '../../components/edit-form-item.vue';
import LoadingIcon from '@/components/loading-icon.vue';
import KeyValue from '@/components/key-value.vue';
import $store from '@/store';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import { useClusterInfo, useClusterList } from '../use-cluster';
import useSideslider from '@/composables/use-sideslider';
import $bkMessage from '@/common/bkmagic';
import { CLUSTER_ENV } from '@/common/constant';
import DescList from '@/components/desc-list.vue';

export default defineComponent({
  name: 'ClusterInfo',
  components: { StatusIcon, EditFormItem, LoadingIcon, KeyValue, DescList },
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
          message: $i18n.t('保存成功'),
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
    const clusterType = computed(() => {
      if (clusterData.value.clusterType === 'virtual') return 'vCluster';
      return clusterData.value.manageType === 'INDEPENDENT_CLUSTER' ? $i18n.t('独立集群') : $i18n.t('托管集群');
    });
    // 修改集群信息
    const handleModifyCluster = async (params  = {}) => {
      isLoading.value = true;
      const result = await $store.dispatch('clustermanager/modifyCluster', {
        $clusterId: clusterData.value?.clusterID,
        ...params,
      });
      result && $bkMessage({
        theme: 'success',
        message: $i18n.t('修改成功'),
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
      $router.push({ name: 'token' });
    };
    // 地域信息
    const regionList = ref<any[]>([]);
    const region = computed(() => regionList.value
      .find(item => item.region === clusterData.value.region)?.regionName || clusterData.value.region);
    const getRegionList = async () => {
      if (curCluster.value?.provider !== 'tencentCloud') return;
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
    width: 50%;
  }
  >>> .bk-form-content {
    line-height: 32px;
    font-size: 12px;
    color: #313238;
  }
}
</style>
