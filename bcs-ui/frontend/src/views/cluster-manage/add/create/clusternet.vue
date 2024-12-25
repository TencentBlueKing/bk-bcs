<!-- eslint-disable max-len -->
<template>
  <BcsContent :padding="0" :title="$t('cluster.button.addCluster')" :desc="$t('cluster.tips.createClusternetCluster')">
    <div class="h-full pt-[8px] bg-[#f0f1f5]">
      <bcs-tab
        :label-height="42"
        :validate-active="false"
        :active.sync="activeTabName"
        type="card-tab"
        class="h-full">
        <!-- 基本信息 -->
        <bcs-tab-panel :name="steps[0].name">
          <template #label>
            <StepTabLabel :title="$t('generic.title.basicInfo1')" :step-num="1" :active="activeTabName === steps[0].name" />
          </template>
          <bk-form
            :ref="steps[0].formRef"
            :model="basicInfo"
            :rules="basicInfoRules"
            class="k8s-form grid grid-cols-2 grid-rows-1 gap-[16px]">
            <DescList size="middle" :title="$t('cluster.create.label.clusterInfo')">
              <bk-form-item :label="$t('cluster.create.label.kubernetesProvider')">
                {{ $t('cluster.create.label.kubernetesCommunity') }}
              </bk-form-item>
              <bk-form-item :label="$t('cluster.labels.name')" property="federationClusterName" error-display-type="normal" required>
                <bk-input
                  :maxlength="64"
                  :placeholder="$t('cluster.create.validate.name')"
                  v-model.trim="basicInfo.federationClusterName">
                </bk-input>
              </bk-form-item>
              <bk-form-item
                :label="$t('k8s.label')"
                property="labels"
                error-display-type="normal">
                <KeyValue
                  v-model="basicInfo.federationClusterLabels"
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
                  ]" />
              </bk-form-item>
              <bk-form-item :label="$t('cluster.create.label.desc')">
                <bk-input maxlength="100" v-model="basicInfo.federationClusterDescription" type="textarea"></bk-input>
              </bk-form-item>
            </DescList>
            <div>
              <DescList size="middle" :title="$t('cluster.labels.env')">
                <bk-form-item :label="$t('cluster.labels.env')" property="federationClusterEnv" error-display-type="normal" required>
                  <bk-radio-group v-model="basicInfo.federationClusterEnv">
                    <bk-radio value="debug">
                      {{ $t('cluster.env.debug') }}
                    </bk-radio>
                    <bk-radio value="prod">
                      {{ $t('cluster.env.prod') }}
                    </bk-radio>
                  </bk-radio-group>
                </bk-form-item>
              </DescList>
            </div>
          </bk-form>
        </bcs-tab-panel>
        <!-- 控制面配置 -->
        <bcs-tab-panel :name="steps[1].name" :disabled="steps[1].disabled">
          <template #label>
            <StepTabLabel
              :title="$t('cluster.detail.title.controlConfig')"
              :step-num="2"
              :active="activeTabName === steps[1].name"
              :disabled="steps[1].disabled" />
          </template>
          <bk-form :ref="steps[1].formRef" :model="masterConfig" :rules="masterConfigRules">
            <bk-form-item
              key="master"
              class="tips-offset"
              :label="$t('cluster.create.label.hostResource')"
              property="resource">
              <bk-radio-group v-model="resourceType">
                <bk-radio value="createHostCluster" disabled>
                  {{ $t('cluster.create.label.hostCluster') }}
                </bk-radio>
                <bk-radio value="existingCluster">
                  {{ $t('cluster.create.label.existingClusters') }}
                </bk-radio>
              </bk-radio-group>
            </bk-form-item>
            <bk-form-item
              :label="$t('generic.label.cluster')"
              property="hostClusterId"
              error-display-type="normal"
              required>
              <div class="flex items-center">
                <bcs-select
                  :loading="isLoading"
                  class="w-[500px]"
                  v-model="hostClusterId"
                  searchable
                  :clearable="false">
                  <bcs-option
                    v-for="item in clusterData"
                    :key="item.clusterID"
                    :id="item.clusterID"
                    :name="item.clusterName"
                    :disabled="item.disabled"
                    v-bk-tooltips="{
                      content: item.disabled ? $t('cluster.create.federation.tips.disabledHost', [unavailableClustersMap[item.clusterID]]) : '',
                      delay: [300, 0],
                      interactive: false,
                    }"
                    v-authority="{
                      clickable: item.clickable,
                      actionId: 'cluster_manage',
                      resourceName: item.clusterName,
                      disablePerms: true,
                      permCtx: {
                        project_id: projectID,
                        cluster_id: item.clusterID
                      }
                    }" />
                </bcs-select>
                <bk-button icon="icon-refresh" @click="handleGetClusterList" class="ml10" :disabled="isLoading"></bk-button>
              </div>
            </bk-form-item>
          </bk-form>
        </bcs-tab-panel>
      </bcs-tab>
      <div class="flex items-center h-[48px] bg-[#FAFBFD] px-[24px] absolute bottom-0 w-full bcs-border-top">
        <bk-button v-if="activeTabName !== steps[0].name" @click="preStep">{{ $t('generic.button.pre') }}</bk-button>
        <bk-button
          theme="primary"
          class="ml10"
          v-if="activeTabName === steps[steps.length - 1].name"
          @click="handleShowConfirmDialog">
          {{ $t('cluster.create.button.createCluster') }}
        </bk-button>
        <bk-button theme="primary" class="ml10" v-else @click="nextStep">{{ $t('generic.button.next') }}</bk-button>
        <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </div>
  </BcsContent>
</template>
<script lang="ts" setup>
import { merge } from 'lodash';
import { computed, getCurrentInstance, onMounted, ref, watch } from 'vue';

import { isAdding, useClusterList } from '../../cluster/use-cluster';

import { createFederalCluster, getFederationClusterId } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import DescList from '@/components/desc-list.vue';
import BcsContent from '@/components/layout/Content.vue';
import { ICluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import StepTabLabel from '@/views/cluster-manage/add/components/step-tab-label.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';

const steps = ref([
  { name: 'basicInfo', formRef: 'basicInfoRef', disabled: false },
  { name: 'master', formRef: 'masterRef', disabled: true },
]);
const activeTabName = ref<typeof steps.value[number]['name']>('basicInfo');

// 基本信息
const basicInfo = ref({
  federationClusterName: '',
  federationClusterEnv: '',
  federationClusterDescription: '',
  federationClusterLabels: {},
});
const resourceType = ref('existingCluster');
const basicInfoRules = ref({
  federationClusterName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  federationClusterEnv: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  labels: [
    {
      message: $i18n.t('generic.validate.label'),
      trigger: 'custom',
      validator: () => {
        const { federationClusterLabels: labels } = basicInfo.value;
        const rule = new RegExp(LABEL_KEY_REGEXP);
        return Object.keys(labels).every(key => rule.test(key) && rule.test(labels[key]));
      },
    },
  ],
});

// master配置
const masterConfig = ref({});
const hostClusterId = ref('');
// 动态 i18n 问题，这里使用computed
const masterConfigRules = ref({
  hostClusterId: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => !!hostClusterId.value,
    },
  ],
});

// 集群列表
const {
  clusterList,
  getClusterList,
} = useClusterList();
const isLoading = ref(false);
/**
 * 尚未组成联邦的TKE独立集群，按照label
 * federation.bkbcs.tencent.com/is-host-cluster： 联邦host集群
 * federation.bkbcs.tencent.com/is-sub-cluster:  联邦子集群
 * federation.bkbcs.tencent.com/is-fed-cluster： 联邦代理集群
*/
const clusterWebAnnotations = computed(() => $store.state.cluster.clusterWebAnnotations.perms);
// 展示TKE独立集群
const clusterData = computed(() => clusterList.value
  ?.filter(cur => !cur.is_shared /* 非共享集群 */
    && cur.clusterType !== 'virtual' /* 非virtual集群 */
    && cur.manageType === 'INDEPENDENT_CLUSTER' /* 非独立集群 */
    && cur.provider === 'tencentCloud') /* TKE集群 */
  ?.map(item => ({
    clusterID: item.clusterID,
    clusterName: item.clusterName,
    disabled: item.labels?.['federation.bkbcs.tencent.com/is-host-cluster'] === 'true',
    clickable: isClickable(item),
  })));
async function handleGetClusterList() {
  isLoading.value = true;
  await getClusterList();
  isLoading.value = false;
};

const unavailableClusters = computed(() => clusterData.value.filter(item => item.disabled));
const unavailableClustersMap = ref({});

// 没有集群管理权限：禁用
function isClickable(cluster: ICluster) {
  return clusterWebAnnotations.value[cluster.clusterID]?.cluster_manage;
};

// 上一步
const preStep = async () => {
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index - 1 >= 0) {
    activeTabName.value = steps.value[index - 1]?.name;
  }
};
// 下一步
const nextStep = async () => {
  const $refs = proxy?.$refs || {};
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  const validate = await ($refs[steps.value[index]?.formRef] as any)?.validate().catch(() => false);
  if (!validate) return;
  if (index > -1 && index + 1 < steps.value.length) {
    steps.value[index + 1].disabled = false;
    activeTabName.value = steps.value[index + 1]?.name;
  }
};
const handleCancel = () => {
  $router.back();
};
// 创建集群
const { proxy } = getCurrentInstance() || { proxy: null };
const handleShowConfirmDialog = async () => {
  const $refs = proxy?.$refs || {};
  const validateList = steps.value.map(step => ($refs[step.formRef] as any)?.validate().catch(() => {
    activeTabName.value = step.name;
    return false;
  }));
  const validateResults = await Promise.all(validateList.filter(item => item));
  if (validateResults.some(result => !result)) return;

  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('cluster.create.button.confirmCreateCluster.text'),
    defaultInfo: true,
    confirmFn: async () => {
      await handleCreateCluster();
    },
  });
};
const { curProject, projectID } = useProject();
const user = computed(() => $store.state.user);
const handleCreateCluster = async () => {
  const params: Record<string, any> = merge({
    $hostClusterId: hostClusterId.value,
    federationProjectId: curProject.value.projectID,
    federationProjectCode: curProject.value.projectCode,
    federationBusinessId: String(curProject.value.businessID),
    creator: user.value.username,
  }, basicInfo.value);
  const result = await createFederalCluster(params).then(() => true)
    .catch(() => false);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.deliveryTask'),
    });
    $router.push({ name: 'clusterMain' });
    // 异步添加中，接口可能未返回数据，需将轮询标识设置为true
    isAdding.value = true;
  }
};

// 获取host集群对应的联邦集群ID
watch(unavailableClusters, () => {
  unavailableClusters.value.forEach(async (cluster) => {
    if (unavailableClustersMap.value[cluster.clusterID]) return;
    const result = await getFederationClusterId({
      $hostClusterId: cluster.clusterID,
    }).catch(() => ({ cluster: {} }));
    if (result.cluster) {
      unavailableClustersMap.value[cluster.clusterID] = result.cluster?.federation_cluster_id;
    }
  });
});

onMounted(() => {
  handleGetClusterList();
});
</script>
<style lang="postcss" scoped>
>>> .bk-tab-header {
  padding: 0 8px;
}
>>> .bk-tab-section {
  overflow: auto;
  height: calc(100% - 80px);
}

>>> .k8s-form .bk-form-content {
  max-width: 600px;
  padding-right: 24px;
}
</style>
