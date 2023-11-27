<template>
  <div class="choose-node-template bcs-content-wrapper">
    <bk-form :rules="formRules" ref="formRef">
      <FormGroup :allow-toggle="false" class="choose-node">
        <bk-form-item :label="$t('manualNode.title.source.text')">
          <bk-radio-group v-model="nodeSource" @change="handleNodeSourceChange">
            <bk-radio
              value="ip"
              :disabled="isImportCluster">
              <span
                v-bk-tooltips="{
                  disabled: !isImportCluster,
                  content: $t('cluster.nodeList.tips.disableImportClusterAddNode')
                }">
                {{ $t('manualNode.title.source.existingServer') }}
              </span>
            </bk-radio>
            <bk-radio
              value="nodePool"
              v-if="curCluster && curCluster.autoScale">
              {{ $t('manualNode.title.source.addFromNodePool') }}
            </bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <template v-if="nodeSource === 'ip'">
          <bk-form-item
            :label="$t('cluster.nodeList.label.selectNode')"
            property="ip"
            error-display-type="normal">
            <bcs-button
              theme="primary"
              icon="plus"
              @click="handleAddNode">
              {{$t('cluster.nodeList.create.text')}}
            </bcs-button>
            <bcs-table class="mt15" :data="ipList" :key="tableKey">
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
              <!-- <bcs-table-column
                :label="$t('generic.ipSelector.label.idc')"
                prop="idc_unit_name">
              </bcs-table-column> -->
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
          <bk-form-item :label="$t('cluster.create.label.initNodeTemplate')" v-if="$INTERNAL">
            <TemplateSelector
              :is-tke-cluster="isTkeCluster"
              :cluster-id="clusterId"
              @template-change="handleTemplateChange" />
          </bk-form-item>
        </template>
        <template v-else-if="nodeSource === 'nodePool'">
          <bk-form-item :label="$t('manualNode.title.nodePool.text')" property="nodePool" error-display-type="normal">
            <div class="flex items-center">
              <bcs-select class="max-w-[600px] flex-1" :loading="nodeGroupLoading" v-model="nodePoolID">
                <bcs-option-group
                  :name="$t('manualNode.title.nodePool.disabledGroup')"
                  v-if="nodeGroupData.disabled.length">
                  <bcs-option
                    v-for="item in nodeGroupData.disabled"
                    :key="item.nodeGroupID"
                    :id="item.nodeGroupID"
                    :name="item.name"
                    :disabled="(item.autoScaling.maxSize - item.autoScaling.desiredSize) <= 0">
                    <div
                      v-bk-tooltips="{
                        disabled: (item.autoScaling.maxSize - item.autoScaling.desiredSize) > 0,
                        content: $t('manualNode.tips.insufficientNode')
                      }">
                      {{ item.name }}
                    </div>
                  </bcs-option>
                </bcs-option-group>
                <bcs-option-group
                  :name="$t('manualNode.title.nodePool.enableGroup')"
                  v-if="nodeGroupData.enabled.length">
                  <bcs-option
                    v-for="item in nodeGroupData.enabled"
                    :key="item.nodeGroupID"
                    :id="item.nodeGroupID"
                    :name="item.name"
                    :disabled="(item.autoScaling.maxSize - item.autoScaling.desiredSize) <= 0">
                    <div
                      v-bk-tooltips="{
                        disabled: (item.autoScaling.maxSize - item.autoScaling.desiredSize) > 0,
                        content: $t('manualNode.tips.insufficientNode')
                      }">
                      {{ item.name }}
                    </div>
                  </bcs-option>
                </bcs-option-group>
                <span slot="extension" class="cursor-pointer" @click="handleGotoNodePool('')">
                  <i class="bcs-icon bcs-icon-fenxiang mr5 !text-[12px]"></i>
                  {{ $t('manualNode.button.gotoNodePool') }}
                </span>
              </bcs-select>
              <span
                class="ml10 text-[12px] cursor-pointer"
                v-bk-tooltips.top="$t('generic.button.refresh')"
                @click="handleGetNodeGroupList">
                <i class="bcs-icon bcs-icon-reset"></i>
              </span>
              <span class="text-[12px] cursor-pointer ml15" v-if="nodePoolID">
                <i
                  class="bcs-icon bcs-icon-yulan"
                  v-bk-tooltips.top="$t('generic.title.preview')"
                  @click="handleGotoNodePool(nodePoolID)">
                </i>
              </span>
            </div>
          </bk-form-item>
          <bk-form-item :label="$t('manualNode.title.desiredSize')" property="desiredSize" error-display-type="normal">
            <bcs-input
              type="number"
              class="w-[74px]"
              :max="maxCount"
              :min="1"
              :precision="0"
              v-model="desiredSize"></bcs-input>
            <span class="text-[#979BA5] ml-[8px]">
              {{ $t('manualNode.tips.desiredSize', [maxCount]) }}
            </span>
          </bk-form-item>
        </template>
      </FormGroup>
    </bk-form>
    <div class="mt25">
      <!-- 添加节点 -->
      <bk-button
        class="mw88"
        theme="primary"
        @click="handleShowConfirmDialog"
        v-if="nodeSource === 'ip'">
        {{$t('generic.button.confirm')}}
      </bk-button>
      <!-- 从节点池添加节点 -->
      <bk-button
        class="mw88"
        theme="primary"
        :loading="saving"
        @click="handleAddDesiredSize"
        v-else-if="nodeSource === 'nodePool'">
        {{$t('generic.button.confirm')}}
      </bk-button>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
    </div>
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
    <!-- IP选择器 -->
    <IpSelector
      :cloud-id="curCluster.provider"
      :region="curCluster.region"
      :vpc="{ vpcID: curCluster.vpcID }"
      :show-dialog="showIpSelector"
      :ip-list="ipList"
      :validate-vpc-and-region="curCluster.provider !== 'bluekingCloud'"
      :account-i-d="curCluster.cloudAccountID"
      @confirm="chooseServer"
      @cancel="showIpSelector = false">
    </IpSelector>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import TemplateSelector from '../components/template-selector.vue';

import useNode from './use-node';

import { desirednode, nodeGroups } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import ConfirmDialog from '@/components/comfirm-dialog.vue';
import IpSelector from '@/components/ip-selector/ip-selector.vue';
import StatusIcon from '@/components/status-icon';
import { ICluster, useConfig } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import FormGroup from '@/views/cluster-manage/add/common/form-group.vue';

export default defineComponent({
  components: { FormGroup, IpSelector, StatusIcon, ConfirmDialog, TemplateSelector },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const tableKey = ref('');
    const formRef = ref();
    const formRules = ref({
      ip: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => ipList.value.length > 0,
      }],
      nodePool: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => !!nodePoolID.value,
      }],
      desiredSize: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => desiredSize.value > 0,
      }],
    });
    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});
    const isTkeCluster = computed(() => curCluster.value?.provider === 'tencentCloud');
    const isImportCluster = computed(() => curCluster.value.clusterCategory === 'importer');
    const statusColorMap = ref({
      0: 'red',
      1: 'green',
    });

    const showIpSelector = ref(false);
    const ipList = ref<any[]>([]);
    const handleAddNode = () => {
      showIpSelector.value = true;
    };
    const handleRemoveIp = (row) => {
      const index = ipList.value.findIndex(item => item?.cloudArea?.id === row?.cloudArea?.id && item.ip === row.ip);
      if (index > -1) {
        ipList.value.splice(index, 1);
        tableKey.value = `${Math.random() * 10}`;
      }
    };
    const chooseServer = (data) => {
      ipList.value = data;
      showIpSelector.value = false;
      formRef.value?.validate().catch(() => false);
    };
    const { addNode } = useNode();
    const checkList = computed(() => [
      $i18n.t('cluster.nodeList.create.button.confirmAdd.article1', {
        ip: ipList.value[0]?.bk_host_innerip,
        num: ipList.value.length,
      }),
      $i18n.t('cluster.nodeList.create.button.confirmAdd.article2'),
    ]);
    const currentTemplate = ref<Record<string, string>>({});
    const handleTemplateChange = (item) => {
      currentTemplate.value = item;
    };
    const showConfirmDialog = ref(false);
    const handleShowConfirmDialog = async () => {
      const result = await formRef.value?.validate().catch(() => false);
      if (!result) return;
      currentTemplate.value.nodeTemplateID
        ? $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('cluster.nodeList.title.confirmUseTemplate', { name: currentTemplate.value.name }),
          defaultInfo: true,
          confirmFn: () => {
            showConfirmDialog.value = true;
          },
        })
        : showConfirmDialog.value = true;
    };
    const confirmLoading = ref(false);
    const handleConfirm = async () => {
      confirmLoading.value = true;
      const result = await addNode({
        clusterId: props.clusterId,
        nodeIps: ipList.value.map(item => item.bk_host_innerip),
        nodeTemplateID: currentTemplate.value.nodeTemplateID,
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
    const handleCancel = () => {
      $router.back();
    };

    // nodeGroups
    const nodePoolID = ref('');
    const desiredSize = ref(1);
    const nodeSource = ref<'nodePool'|'ip'>(isImportCluster.value ? 'nodePool' : 'ip');
    const handleNodeSourceChange = (value: 'nodePool'|'ip') => {
      if (value === 'nodePool') {
        handleGetNodeGroupList();
      }
    };
    const nodeGroupsList = ref<INodePool[]>([]);
    const curNodePool = computed(() => nodeGroupsList.value.find(item => item.nodeGroupID === nodePoolID.value));
    // 分组数据
    const nodeGroupData = computed(() => {
      const groupData: {
        enabled: INodePool[]
        disabled: INodePool[]
      } = {
        enabled: [],
        disabled: [],
      };
      nodeGroupsList.value.forEach((item) => {
        if (item.enableAutoscale) {
          groupData.enabled.push(item);
        } else {
          groupData.disabled.push(item);
        }
      });
      return groupData;
    });
    // 最大可添加节点数量
    const maxCount = computed(() => {
      const maxSize = curNodePool.value?.autoScaling?.maxSize || 0;
      const desiredSize = curNodePool.value?.autoScaling?.desiredSize || 0;
      return maxSize - desiredSize;
    });
    const nodeGroupLoading = ref(false);
    const handleGetNodeGroupList = async () => {
      nodeGroupLoading.value = true;
      nodeGroupsList.value = await nodeGroups({
        $clusterId: props.clusterId,
        enableFilter: true,
      }).catch(() => []);
      nodeGroupLoading.value = false;
    };
    const user = computed(() => $store.state.user);
    const { _INTERNAL_ } = useConfig();
    const saving = ref(false);
    // 从节点池添加节点
    const handleAddDesiredSize = async () => {
      const result = await formRef.value?.validate().catch(() => false);
      if (!result) return;

      // 判断是否配置了扩容转移模块
      saving.value = true;
      const autoscalerData = await $store.dispatch('clustermanager/clusterAutoScaling', {
        $clusterId: props.clusterId,
        provider: _INTERNAL_.value ? 'selfProvisionCloud' : '',
      });
      saving.value = false;
      if (!autoscalerData.module?.scaleOutModuleID || !autoscalerData.module?.scaleOutBizID) {
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('cluster.ca.tips.noModule2'),
          defaultInfo: true,
          okText: $i18n.t('cluster.ca.button.edit'),
          confirmFn: () => {
            const { href } = $router.resolve({
              name: 'autoScalerConfig',
              params: {
                clusterId: props.clusterId,
              },
            });
            window.open(href);
          },
        });
      } else {
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: curNodePool.value?.enableAutoscale ? $i18n.t('manualNode.msg.confirmText1') : $i18n.t('manualNode.msg.confirmText'),
          defaultInfo: true,
          confirmFn: async () => {
            saving.value = true;
            const result = await desirednode({
              $id: nodePoolID.value,
              desiredNode: Number(desiredSize.value) + Number(curNodePool.value?.autoScaling?.desiredSize || 0),
              manual: true,
              operator: user.value.username,
            });
            saving.value = false;
            if (result) {
              $bkMessage({
                theme: 'success',
                message: $i18n.t('generic.msg.success.deliveryTask'),
              });
              $router.push({
                name: 'clusterMain',
                query: {
                  clusterId: props.clusterId,
                  active: 'node',
                },
              });
            }
          },
        });
      }
    };
    const handleGotoNodePool = (id?: string) => {
      if (id) {
        const { href } = $router.resolve({
          name: 'nodePoolDetail',
          params: {
            clusterId: props.clusterId,
            nodeGroupID: id,
          },
        });
        window.open(href);
      } else {
        const { href } = $router.resolve({
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            clusterId: props.clusterId,
          },
        });
        window.open(href);
      }
    };

    onBeforeMount(() => {
      if (nodeSource.value === 'nodePool') {
        handleGetNodeGroupList();
      }
    });

    return {
      tableKey,
      saving,
      isImportCluster,
      formRef,
      formRules,
      curCluster,
      isTkeCluster,
      confirmLoading,
      showConfirmDialog,
      checkList,
      showIpSelector,
      ipList,
      statusColorMap,
      nodeSource,
      nodePoolID,
      nodeGroupData,
      nodeGroupLoading,
      maxCount,
      desiredSize,
      handleNodeSourceChange,
      handleRemoveIp,
      chooseServer,
      handleCancel,
      handleShowConfirmDialog,
      handleConfirm,
      handleAddNode,
      handleTemplateChange,
      handleGetNodeGroupList,
      handleAddDesiredSize,
      handleGotoNodePool,
    };
  },
});
</script>
<style lang="postcss" scoped>
.choose-node-template {
  padding: 24px;
  >>> .choose-node {
    .form-group-content {
      padding-top: 0;
    }
  }
}
</style>
