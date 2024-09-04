<template>
  <BcsContent :padding="0" class="overflow-auto">
    <div class="shadow bg-[#fff] px-[24px] py-[16px] pb-[32px] w-[800px] mx-auto my-[20px]">
      <div class="text-[14px] font-bold mb-[24px]">{{ $t('importBkSopsCloud.title') }}</div>
      <bk-form class="max-w-[640px]" :model="formData" :rules="formRules" ref="formRef">
        <bk-form-item
          :label="$t('importBkSopsCloud.label.clusterName')"
          required
          property="clusterName"
          error-display-type="normal">
          <bcs-input v-model="formData.clusterName"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.labels.env')"
          property="environment"
          error-display-type="normal"
          required>
          <bk-radio-group class="btn-group" v-model="formData.environment">
            <bk-radio class="btn-group-first" value="debug">
              {{ $t('cluster.env.debug') }}
            </bk-radio>
            <bk-radio value="prod">
              {{ $t('cluster.env.prod') }}
            </bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item
          :label="$t('importBkSopsCloud.label.controlPanel')"
          property="ipList"
          required
          :desc="$t('importBkSopsCloud.message.nodeTips')"
          error-display-type="normal"
          ref="ipSelectorRef">
          <bcs-button
            theme="primary"
            outline
            icon="plus"
            @click="handleAddNode">
            {{$t('importBkSopsCloud.label.addServer')}}
          </bcs-button>
          <bcs-table class="mt15 max-w-[800px]" height="200px" :data="formData.ipList" :key="tableKey">
            <bcs-table-column type="index" :label="$t('cluster.nodeList.label.index')" width="60"></bcs-table-column>
            <bcs-table-column
              :label="$t('generic.ipSelector.label.innerIp')"
              prop="bk_host_innerip"
              width="120">
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.ipSelector.label.agentStatus')" width="100">
              <template #default="{ row }">
                <StatusIcon :status="String(row.agent_alive)" :status-color-map="statusColorMap">
                  {{row.agent_alive ? $t('generic.status.ready') : $t('generic.status.error')}}
                </StatusIcon>
              </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('generic.label.action')" min-width="100">
              <template #default="{ row }">
                <bk-button text @click="handleRemoveIp(row)">{{$t('cluster.create.button.remove')}}</bk-button>
              </template>
            </bcs-table-column>
          </bcs-table>
        </bk-form-item>
        <bk-form-item
          :label="$t('importBkSopsCloud.label.advancedParameters')"
          property="extraInfo">
          <KeyValue
            v-model="formData.extraInfo"
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
        <bk-form-item :label="$t('importBkSopsCloud.label.desc')" property="description" error-display-type="normal">
          <bcs-input :maxlength="100" type="textarea" v-model="formData.description"></bcs-input>
        </bk-form-item>
        <bk-form-item>
          <span
            v-bk-tooltips="{
              disabled: !hasAbnormal,
              content: $t('importBkSopsCloud.message.errorAgent')
            }">
            <bk-button
              :loading="importLoading"
              :disabled="hasAbnormal"
              theme="primary"
              class="min-w-[88px] ml-[10px]"
              @click="handleImport">
              {{ $t('importBkSopsCloud.button.import') }}
            </bk-button>
          </span>
          <bk-button class="ml-[10px]" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
        </bk-form-item>
        <!-- IP选择器 -->
        <IpSelector
          :show-dialog="showIpSelector"
          :ip-list="formData.ipList"
          validate-vpc
          @confirm="chooseServer"
          @cancel="showIpSelector = false">
        </IpSelector>
      </bk-form>
    </div>
  </BcsContent>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';

import { importCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import { LABEL_KEY_REGEXP } from '@/common/constant';
import IpSelector from '@/components/ip-selector/ip-selector.vue';
import BcsContent from '@/components/layout/Content.vue';
import StatusIcon from '@/components/status-icon';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';

const formRef = ref<any>();
const formData = ref<{
  clusterName: string,
  description: string,
  extraInfo: object,
  environment: string,
  ipList: Array<{
    ip: string
    cloudArea: {
      id: string
    },
    agent_alive: number
  }>,
}>({
  clusterName: '',
  description: '',
  extraInfo: {},
  ipList: [],
  environment: '',
});
const formRules = ref({
  clusterName: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  ipList: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
  extraInfo: [
    {
      message: $i18n.t('generic.validate.label'),
      trigger: 'custom',
      validator: () => {
        const { extraInfo } = formData.value;
        const rule = new RegExp(LABEL_KEY_REGEXP);
        return Object.keys(extraInfo).every(key => rule.test(key) && rule.test(extraInfo[key]));
      },
    },
  ],
  environment: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
});
const tableKey = ref('');
const statusColorMap = ref({
  0: 'red',
  1: 'green',
});

// IP选择器
const showIpSelector = ref(false);
const handleAddNode = () => {
  showIpSelector.value = true;
};

const chooseServer = (data) => {
  formData.value.ipList = data;
  showIpSelector.value = false;
  formRef.value?.$refs?.ipSelectorRef?.validate();
};

const handleRemoveIp = (row) => {
  const index = formData.value.ipList
    .findIndex(item => item?.cloudArea?.id === row?.cloudArea?.id && item.ip === row.ip);
  if (index > -1) {
    formData.value.ipList.splice(index, 1);
    formData.value.ipList = [...formData.value.ipList];
    tableKey.value = `${Math.random() * 10}`;
  }
};

// 导入校验
const hasAbnormal = computed(() => formData.value.ipList.some(item => item?.agent_alive === 0));

// 集群导入
const curProject = computed(() => $store.state.curProject);
const user = computed(() => $store.state.user);
const importLoading = ref(false);
const handleImport = async () => {
  const validate = await formRef.value.validate().catch(() => false);
  if (!validate) return;
  const nodeIps = formData.value.ipList.map(item => item.ip);

  importLoading.value = true;
  const params = {
    clusterName: formData.value.clusterName,
    description: formData.value.description,
    projectID: curProject.value.projectID,
    businessID: String(curProject.value.businessID),
    provider: 'bluekingCloud',
    region: 'default',
    environment: formData.value.environment,
    engineType: 'k8s',
    isExclusive: true,
    clusterType: 'single',
    manageType: 'INDEPENDENT_CLUSTER',
    creator: user.value.username,
    cloudMode: {
      cloudID: '',
      kubeConfig: '',
      inter: true,
      nodeIps,
    },
    networkType: 'overlay',
    accountID: '',
    area: {},
    extraInfo: formData.value.extraInfo,
  };
  const result = await importCluster(params).then(() => true)
    .catch(() => false);
  importLoading.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.import'),
    });
    $router.push({ name: 'clusterMain' });
  }
};
// 取消
const handleCancel = () => {
  $router.back();
};
</script>
