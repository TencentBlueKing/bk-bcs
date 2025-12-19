<template>
  <div>
    <bk-form :rules="formRules" ref="formRef" class="bg-[#fff] pb-[20px]">
      <bk-form-item v-if="curCluster.provider === 'tencentCloud'" :label="$t('cluster.nodeList.label.nodeType')">
        <bk-radio-group v-model="formData.advance.isGPUNode" @change="handleNodeTypeChange">
          <bk-radio :value="false">{{$t('cluster.nodeList.label.cvmNode')}}</bk-radio>
          <bk-radio :value="true">{{$t('cluster.nodeList.label.gpuNode')}}</bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item
        :label="$t('cluster.nodeList.label.selectNode')"
        property="ip"
        error-display-type="normal"
        ref="ipSelectorRef">
        <bcs-button
          theme="primary"
          icon="plus"
          @click="handleAddNode">
          {{$t('cluster.nodeList.create.text')}}
        </bcs-button>
        <bcs-table class="mt15 max-w-[800px]" :data="formData.ipList" :key="tableKey">
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
          <bcs-table-column
            :label="$t('generic.ipSelector.label.serverModel')"
            prop="instanceType">
          </bcs-table-column>
          <bcs-table-column :label="$t('generic.label.action')" width="100">
            <template #default="{ row }">
              <bk-button text @click="handleRemoveIp(row)">{{$t('cluster.create.button.remove')}}</bk-button>
            </template>
          </bcs-table-column>
        </bcs-table>
      </bk-form-item>
      <bk-form-item :label="$t('cluster.create.label.initNodeTemplate')">
        <TemplateSelector
          class="max-w-[500px]"
          :provider="curCluster.provider"
          @template-change="handleTemplateChange" />
      </bk-form-item>
      <!-- 用户名和密码 -->
      <bk-form-item
        :label="$t('tke.label.loginType.text')"
        property="masterLogin"
        error-display-type="normal"
        required
        v-if="curCluster.provider === 'tencentPublicCloud'"
        ref="loginTypeRef"
        :key="loginType">
        <LoginType
          :region="curCluster.region"
          :cloud-account-i-d="curCluster.cloudAccountID"
          :cloud-i-d="curCluster.provider"
          init-data
          :value="workerLogin"
          :type="loginType"
          @pass-blur="validateLogin('custom')"
          @confirm-pass-blur="validateLogin('')"
          @key-secret-blur="validateLogin('')"
          @change="handleLoginValueChange"
          @type-change="(v) => loginType = v"
          @pass-change="(v) => confirmPassword = v" />
      </bk-form-item>
      <!-- 操作系统 -->
      <bk-form-item
        :label="$t('cluster.create.label.system')"
        property="imageID"
        error-display-type="normal"
        v-if="curCluster.provider === 'tencentCloud' && !formData.advance.isGPUNode">
        <ImageList
          class="max-w-[500px]"
          v-model="formData.advance.nodeOs"
          :region="curCluster.region"
          :cloud-i-d="curCluster.provider"
          :cluster-id="clusterId"
          init-data
          @init="handleOs" />
      </bk-form-item>
      <!-- IP选择器 -->
      <IpSelector
        :cloud-id="curCluster.provider"
        :region="curCluster.region"
        :vpc="{ vpcID: curCluster.vpcID }"
        :show-dialog="showIpSelector"
        :ip-list="formData.ipList"
        :validate-vpc-and-region="curCluster.provider !== 'bluekingCloud'"
        :account-i-d="curCluster.cloudAccountID"
        validate-vpc
        :validate-agent-status="curCluster.provider === 'tencentCloud'"
        :validate-data-disk="curCluster.provider === 'tencentCloud'"
        :validate-node-type="curCluster.provider === 'tencentCloud' ? validateNodeType : undefined"
        @confirm="chooseServer"
        @cancel="showIpSelector = false">
      </IpSelector>
    </bk-form>
    <ConfirmDialog
      v-model="showConfirmDialog"
      :title="$t('cluster.nodeList.title.confirmAddNode')"
      :sub-title="$t('generic.subTitle.confirmConfig')"
      :tips="checkList"
      :ok-text="$t('cluster.nodeList.create.button.confirmAdd.text')"
      :cancel-text="$t('generic.button.cancel')"
      :confirm="handleConfirm"
      theme="primary">
    </ConfirmDialog>
    <div class="mt25">
      <!-- 添加节点 -->
      <bk-button
        class="mw88"
        theme="primary"
        @click="handleShowConfirmDialog">
        {{$t('generic.button.confirm')}}
      </bk-button>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
    </div>
  </div>
</template>
<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, ref } from 'vue';

import TemplateSelector from '../components/template-selector.vue';

import ImageList from './tencent-image-list.vue';
import useNode from './use-node';

import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import IpSelector from '@/components/ip-selector/ip-selector.vue';
import StatusIcon from '@/components/status-icon';
import { ICluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import LoginType from '@/views/cluster-manage/add/components/login-type.vue';

export default defineComponent({
  components: { IpSelector, StatusIcon, ConfirmDialog, TemplateSelector, LoginType, ImageList },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const tableKey = ref('');
    const formRef = ref();
    const formData = ref<{
      currentTemplate: Record<string, string>
      ipList: Array<{
        ip: string
        isGpuNode: boolean
        cloudArea: {
          id: string
        }
      }>
      advance: {
        nodeOs: string
        isGPUNode: boolean
      }
    }>({
      currentTemplate: {},
      ipList: [],
      advance: {
        nodeOs: '',
        isGPUNode: false,
      },
    });
    const formRules = ref({
      ip: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => formData.value.ipList.length > 0,
      }],
      masterLogin: [
        {
          trigger: 'custom',
          message: $i18n.t('generic.validate.required'),
          validator() {
            if (loginType.value === 'password') {
              return !!workerLogin.value.initLoginPassword;
            }
            return !!workerLogin.value.keyPair.keyID && !!workerLogin.value.keyPair.keySecret;
          },
        },
        {
          trigger: 'custom',
          message: $i18n.t('cluster.ca.nodePool.create.validate.password'),
          validator() {
            if (loginType.value === 'password') {
              const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/;
              return regex.test(workerLogin.value.initLoginPassword);
            }
            return true;
          },
        },
        {
          trigger: 'custom',
          message: $i18n.t('tke.validate.passwordNotSame'),
          validator() {
            if (loginType.value === 'password' && confirmPassword.value) {
              return workerLogin.value.initLoginPassword === confirmPassword.value;
            }
            return true;
          },
        },
        {
          trigger: 'blur',
          message: $i18n.t('tke.validate.passwordNotSame'),
          validator() {
            if (loginType.value === 'password') {
              return workerLogin.value.initLoginPassword === confirmPassword.value;
            }
            return true;
          },
        },
      ],
      imageID: [
        {
          trigger: 'custom',
          message: $i18n.t('generic.validate.required'),
          validator() {
            return !!formData.value.advance.nodeOs;
          },
        },
      ],
    });

    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});

    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });

    // IP选择器
    const showIpSelector = ref(false);
    const validateNodeType = ref<'cvm'|'gpu'|undefined>();
    const handleAddNode = () => {
      if (curCluster.value.provider === 'tencentCloud') {
        validateNodeType.value = formData.value.advance.isGPUNode ? 'gpu' : 'cvm';
      }
      showIpSelector.value = true;
    };
    const handleRemoveIp = (row) => {
      const index = formData.value.ipList
        .findIndex(item => item?.cloudArea?.id === row?.cloudArea?.id && item.ip === row.ip);
      if (index > -1) {
        formData.value.ipList.splice(index, 1);
        tableKey.value = `${Math.random() * 10}`;
      }
    };
    const chooseServer = (data) => {
      formData.value.ipList = data;
      showIpSelector.value = false;
      formRef.value?.$refs?.ipSelectorRef?.validate();
    };
    const checkList = computed(() => [
      $i18n.t('cluster.nodeList.create.button.confirmAdd.article1', {
        ip: formData.value.ipList[0]?.ip,
        num: formData.value.ipList.length,
      }),
      $i18n.t('cluster.nodeList.create.button.confirmAdd.article2'),
    ]);
    function handleNodeTypeChange(val: boolean) {
      // gpu节点不需要传操作系统
      if (val) {
        formData.value.advance.nodeOs = '';
      }
      formData.value.ipList = [];
      validateNodeType.value = undefined;
    }

    // 节点模板
    const handleTemplateChange = (item) => {
      formData.value.currentTemplate = item;
    };
    // 登录信息
    const loginType = ref<'password'|'ssh'>('password');
    const confirmPassword = ref('');
    const workerLogin = ref({
      initLoginUsername: '',
      initLoginPassword: '',
      keyPair: {
        keyID: '',
        keySecret: '',
        keyPublic: '',
      },
    });
    const handleLoginValueChange = (data) => {
      workerLogin.value = data;
    };
    const validateLogin = (trigger = '') => {
      formRef.value?.$refs?.loginTypeRef?.validate(trigger);
    };

    // 确认节点模板
    const showConfirmDialog = ref(false);
    const handleShowConfirmDialog = async () => {
      const result = await formRef.value?.validate().catch(() => false);
      if (!result) return;
      formData.value.currentTemplate.nodeTemplateID
        ? $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('cluster.nodeList.title.confirmUseTemplate', { name: formData.value.currentTemplate.name }),
          defaultInfo: true,
          confirmFn: () => {
            showConfirmDialog.value = true;
          },
        })
        : showConfirmDialog.value = true;
    };
    // 确认添加节点
    const { addNode } = useNode();
    const confirmLoading = ref(false);
    const handleConfirm = async () => {
      confirmLoading.value = true;
      const { ipList, currentTemplate, advance } = formData.value;
      // 腾讯自研云 若为当前集群使用镜像，则不传
      const cloneAdvance = cloneDeep(advance);
      if (cloneAdvance.nodeOs === defaultOs.value) {
        cloneAdvance.nodeOs = '';
      }
      const result = await addNode({
        clusterId: props.clusterId,
        nodeIps: ipList.map(item => item.ip),
        nodeTemplateID: currentTemplate.nodeTemplateID,
        login: workerLogin.value,
        advance: cloneAdvance,
      });
      confirmLoading.value = false;

      if (result) {
        $router.push({
          name: 'clusterMain',
          query: {
            active: 'node',
            clusterId: props.clusterId,
          },
        });
      }
    };

    // 腾讯自研云 获取当前集群使用镜像
    const defaultOs = ref('');
    function handleOs(data) {
      defaultOs.value = data?.imageID;
    }

    const handleCancel = () => {
      $router.back();
    };

    return {
      tableKey,
      formRef,
      formRules,
      curCluster,
      workerLogin,
      loginType,
      confirmPassword,
      confirmLoading,
      showConfirmDialog,
      checkList,
      showIpSelector,
      formData,
      statusColorMap,
      validateNodeType,
      handleRemoveIp,
      chooseServer,
      handleCancel,
      handleShowConfirmDialog,
      handleConfirm,
      handleAddNode,
      handleTemplateChange,
      handleLoginValueChange,
      validateLogin,
      handleOs,
      handleNodeTypeChange,
    };
  },
});
</script>
