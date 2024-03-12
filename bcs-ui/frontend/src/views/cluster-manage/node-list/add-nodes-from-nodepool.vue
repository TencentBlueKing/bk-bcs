<template>
  <div>
    <bk-form :rules="formRules" ref="formRef" class="bg-[#fff] pb-[20px]">
      <bk-form-item :label="$t('manualNode.title.nodePool.text')" property="nodePool" error-display-type="normal">
        <div class="flex items-center">
          <bcs-select class="max-w-[600px] flex-1" :loading="nodeGroupLoading" v-model="formData.nodePoolID">
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
                    content: $t('manualNode.tips.insufficientNode'),
                    placement: 'right'
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
                    content: $t('manualNode.tips.insufficientNode'),
                    placement: 'right'
                  }">
                  {{ item.name }}
                </div>
              </bcs-option>
            </bcs-option-group>
            <SelectExtension
              slot="extension"
              :link-text="$t('manualNode.button.gotoNodePool')"
              @link="handleGotoAutoScaler"
              @refresh="handleGetNodeGroupList" />
          </bcs-select>
          <span class="text-[12px] cursor-pointer ml15" v-if="formData.nodePoolID">
            <i
              class="bcs-icon bcs-icon-yulan"
              v-bk-tooltips.top="$t('generic.title.preview')"
              @click="handleGotoNodePool(formData.nodePoolID)">
            </i>
          </span>
        </div>
      </bk-form-item>
      <bk-form-item :label="$t('manualNode.title.desiredSize')" property="desiredSize" error-display-type="normal">
        <bcs-input
          type="number"
          class="w-[74px]"
          :max="nodesCounts.maxCount"
          :min="0"
          :precision="0"
          v-model="formData.desiredSize"></bcs-input>
        <i18n path="manualNode.tips.desiredSize" tag="span" class="text-[#979BA5] ml-[8px]">
          <span place="remaining" class="text-[#313238]">{{ nodesCounts.maxCount }}</span>
          <span place="maxSize">{{ nodesCounts.maxSize }}</span>
          <span place="used">{{ nodesCounts.desiredSize }}</span>
          <span place="total">{{ nodesCounts.total }}</span>
        </i18n>
      </bk-form-item>
    </bk-form>
    <div class="mt25">
      <!-- 从节点池添加节点 -->
      <bk-button
        class="mw88"
        theme="primary"
        :loading="saving"
        @click="handleAddDesiredSize">
        {{$t('generic.button.confirm')}}
      </bk-button>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import { useClusterInfo } from '../cluster/use-cluster';

import { desirednode, nodeGroups } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { ICluster, useConfig } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';

export default defineComponent({
  components: {  SelectExtension },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
    nodePool: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { clusterData, getClusterDetail } = useClusterInfo();// clusterData和curCluster一样，就是多了云上的数据信息

    const formRef = ref();
    const formData = ref({
      nodePoolID: props.nodePool,
      desiredSize: 0,
    });
    const formRules = ref({
      nodePool: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => !!formData.value.nodePoolID,
      }],
      desiredSize: [{
        message: $i18n.t('generic.validate.required'),
        trigger: 'blur',
        validator: () => !!formData.value.desiredSize,
      }],
    });
    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});

    // nodeGroups
    const nodeGroupsList = ref<INodePool[]>([]);
    const curNodePool = computed(() => nodeGroupsList.value
      .find(item => item.nodeGroupID === formData.value.nodePoolID));

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
    const nodesCounts = computed(() => {
      const maxSize = curNodePool.value?.autoScaling?.maxSize || 0;
      const desiredSize = curNodePool.value?.autoScaling?.desiredSize || 0;
      return {
        maxCount: maxSize - desiredSize,
        maxSize,
        desiredSize,
        total: 1000,
      };
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
      if (!clusterData.value?.clusterBasicSettings?.module?.workerModuleID) {
        $bkInfo({
          type: 'warning',
          clsName: 'custom-info-confirm',
          title: $i18n.t('cluster.ca.tips.noModule2'),
          defaultInfo: true,
          okText: $i18n.t('cluster.ca.button.edit'),
          confirmFn: () => {
            const { href } = $router.resolve({
              name: 'clusterMain',
              query: {
                clusterId: props.clusterId,
                active: 'node',
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
            const desiredNode = Number(formData.value.desiredSize)
              + Number(curNodePool.value?.autoScaling?.desiredSize || 0);
            const result = await desirednode({
              $id: formData.value.nodePoolID,
              desiredNode,
              manual: true,
              operator: user.value.username,
            }).catch(() => false);
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
    // 前往CA列表
    const handleGotoAutoScaler = () => {
      const { href } = $router.resolve({
        name: 'clusterMain',
        query: {
          active: 'autoscaler',
          clusterId: props.clusterId,
        },
      });
      window.open(href);
    };
    // 前往节点池详情
    const handleGotoNodePool = (id: string) => {
      let name = '';
      switch (clusterData.value?.provider) {
        case 'gcpCloud':
          name = 'googleNodePoolDetail';
          break;
        default:
          name = 'nodePoolDetail';
      }
      const { href } = $router.resolve({
        name,
        params: {
          clusterId: props.clusterId,
          nodeGroupID: id,
        },
      });
      window.open(href);
    };

    const handleCancel = () => {
      $router.back();
    };

    onBeforeMount(() => {
      getClusterDetail(props.clusterId);
      handleGetNodeGroupList();
    });

    return {
      formData,
      saving,
      formRef,
      formRules,
      curCluster,
      nodeGroupData,
      nodeGroupLoading,
      nodesCounts,
      handleGetNodeGroupList,
      handleAddDesiredSize,
      handleGotoNodePool,
      handleGotoAutoScaler,
      handleCancel,
    };
  },
});
</script>
