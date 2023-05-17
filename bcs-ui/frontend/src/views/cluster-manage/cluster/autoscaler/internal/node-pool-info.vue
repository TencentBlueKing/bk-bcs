<template>
  <div class="node-pool-wrapper">
    <bcs-resize-layout
      placement="right"
      :class="collapse ? '' : 'node-pool'"
      collapsible
      :initial-divide="400"
      :border="false"
      :min="3"
      disabled
      @collapse-change="handleCollapseChange">
      <template #aside>
        <ActionDoc type="autoscaler" class="aside" :title="$t('初始化配置说明')" />
      </template>
      <template #main>
        <div class="main" ref="nodePoolInfoRef">
          <FormGroup :title="$t('基本信息')" :allow-toggle="false">
            <BasicPoolInfo
              :schema="schema"
              :default-values="defaultValues"
              :is-edit="isEdit"
              :cluster="cluster"
              ref="basicInfoRef">
            </BasicPoolInfo>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('Kubelet组件参数配置')" :allow-toggle="false">
            <KubeletParams v-model="nodePoolInfoData.nodeTemplate.extraArgs.kubelet" ref="kubeletRef"></KubeletParams>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('扩容节点初始化配置')" :allow-toggle="false">
            <p>{{$t('前置初始化')}}</p>
            <bcs-input
              type="textarea"
              class="mt10"
              :rows="6"
              :placeholder="$t('请输入 bash 脚本')"
              v-model="nodePoolInfoData.nodeTemplate.preStartUserScript">
            </bcs-input>
            <p class="mt-[32px]">{{$t('后置初始化')}}</p>
            <div class="mt-[10px]">
              <bcs-select class="max-w-[524px]" :clearable="false" v-model="scaleOutPostActionType">
                <bcs-option id="simple" :name="$t('简单脚本执行')"></bcs-option>
                <bcs-option id="complex" :name="$t('标准运维流程执行')"></bcs-option>
              </bcs-select>
              <bcs-input
                type="textarea"
                class="mt10"
                :rows="6"
                :placeholder="$t('请输入 bash 脚本')"
                v-if="scaleOutPostActionType === 'simple'"
                v-model="nodePoolInfoData.nodeTemplate.userScript">
              </bcs-input>
              <BkSops
                class="mt10"
                actions-key="postActions"
                :cluster-id="cluster.clusterID"
                :addons="nodePoolInfoData.nodeTemplate.scaleOutExtraAddons"
                ref="scaleOutRef"
                v-else>
              </BkSops>
            </div>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('节点回收前清理配置')" :allow-toggle="false">
            <bcs-select class="max-w-[524px]" :clearable="false" v-model="scaleInPreActionType">
              <bcs-option id="simple" :name="$t('简单脚本执行')"></bcs-option>
              <bcs-option id="complex" :name="$t('标准运维流程执行')"></bcs-option>
            </bcs-select>
            <bcs-input
              type="textarea"
              class="mt10"
              :rows="6"
              :placeholder="$t('请输入 bash 脚本')"
              v-if="scaleInPreActionType === 'simple'"
              v-model="nodePoolInfoData.nodeTemplate.scaleInPreScript">
            </bcs-input>
            <BkSops
              class="mt10"
              actions-key="preActions"
              :addons="nodePoolInfoData.nodeTemplate.scaleInExtraAddons"
              :cluster-id="cluster.clusterID"
              ref="scaleInRef"
              v-else>
            </BkSops>
          </FormGroup>
        </div>
      </template>
    </bcs-resize-layout>
    <div class="bcs-fixed-footer" v-if="showFooter">
      <bcs-button @click="handlePre">{{$t('上一步')}}</bcs-button>
      <bcs-button
        theme="primary"
        :loading="saveLoading"
        class="ml10"
        @click="handleSaveNodePoolData">
        {{isEdit ? $t('保存节点规格') : $t('创建节点规格')}}
      </bcs-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('取消') }}</bk-button>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, toRefs } from 'vue';
import FormGroup from '@/views/cluster-manage/cluster/create/form-group.vue';
import BasicPoolInfo from './basic-pool-info.vue';
import KubeletParams from './kubelet-params.vue';
import BkSops from './bk-sops.vue';
import $router from '@/router';
import { mergeDeep } from '@/common/util';
import ActionDoc from '@/views/cluster-manage/components/action-doc.vue';

export default defineComponent({
  name: 'NodePoolInfo',
  components: { FormGroup, BasicPoolInfo, KubeletParams, BkSops, ActionDoc },
  props: {
    schema: {
      type: Object,
      default: () => ({}),
    },
    // 详情数据或者默认值
    defaultValues: {
      type: Object,
      default: () => ({}),
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    cluster: {
      type: Object,
      default: () => ({}),
    },
    showFooter: {
      type: Boolean,
      default: true,
    },
    saveLoading: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { defaultValues } = toRefs(props);
    const nodePoolInfoData = ref({
      nodeTemplate: {
        extraArgs: {
          kubelet: defaultValues.value?.nodeTemplate?.extraArgs?.kubelet || '',
        },
        preStartUserScript: defaultValues.value?.nodeTemplate?.preStartUserScript || '', // 扩容前置脚本
        userScript: defaultValues.value?.nodeTemplate?.userScript || '', // 扩容后置脚本
        scaleOutExtraAddons: defaultValues.value?.nodeTemplate?.scaleOutExtraAddons || {}, // 扩容后置流程
        scaleInPreScript: defaultValues.value?.nodeTemplate?.scaleInPreScript || '', // 缩容前置脚本
        scaleInExtraAddons: defaultValues.value?.nodeTemplate?.scaleInExtraAddons || {}, // 缩容后置流程
        labels: {}, // basic-pool-info里面赋值
      },
      labels: {},
    });

    const scaleOutPostActionType = ref<'complex' | 'simple'>('simple');
    const scaleInPreActionType = ref<'complex' | 'simple'>('simple');

    const nodePoolInfoRef = ref<any>(null);
    const basicInfoRef = ref<any>(null);
    const kubeletRef = ref<any>(null);
    const scaleInRef = ref<any>(null);
    const scaleOutRef = ref<any>(null);

    const getNodePoolData = () => {
      // 处理基本参数
      nodePoolInfoData.value = mergeDeep(nodePoolInfoData.value, basicInfoRef.value?.nodePoolInfo || {});
      // 处理扩容脚本参数
      if (scaleOutPostActionType.value === 'complex') {
        nodePoolInfoData.value.nodeTemplate.userScript = '';
        nodePoolInfoData.value.nodeTemplate.scaleOutExtraAddons = scaleOutRef.value?.bkSopsData;
      } else {
        nodePoolInfoData.value.nodeTemplate.scaleOutExtraAddons = {};
      }

      // 处理缩容前置脚本参数
      if (scaleInPreActionType.value === 'complex') {
        nodePoolInfoData.value.nodeTemplate.scaleInPreScript = '';
        nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons = scaleInRef.value?.bkSopsData;
      } else {
        nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons = {};
      }

      // 处理label参数 后端label放两地方
      nodePoolInfoData.value.labels = nodePoolInfoData.value.nodeTemplate.labels;
      return nodePoolInfoData.value;
    };
    const validate = async () => {
      const basicFormValidate = await basicInfoRef.value?.validate().catch(() => false);
      if (!basicFormValidate) {
        // 滚动到顶部
        nodePoolInfoRef.value.scrollTop = 0;
        return false;
      }
      const kubeletValidate = kubeletRef.value?.validateKubeletParams;
      if (!kubeletValidate) return false;

      return true;
    };
    const handlePre = () => {
      ctx.emit('pre');
    };
    const handleSaveNodePoolData = async () => {
      const result = await validate();
      if (!result) return;

      ctx.emit('next', getNodePoolData());
      ctx.emit('confirm');
    };
    const handleCancel = () => {
      $router.back();
    };

    const collapse = ref(false);
    const handleCollapseChange = (value) => {
      collapse.value = value;
    };
    onMounted(() => {
      scaleOutPostActionType.value = nodePoolInfoData.value.nodeTemplate.scaleOutExtraAddons?.postActions?.length ? 'complex' : 'simple';
      scaleInPreActionType.value = nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons?.preActions?.length ? 'complex' : 'simple';
    });
    return {
      collapse,
      nodePoolInfoRef,
      basicInfoRef,
      kubeletRef,
      scaleInRef,
      scaleOutRef,
      scaleOutPostActionType,
      scaleInPreActionType,
      nodePoolInfoData,
      getNodePoolData,
      validate,
      handlePre,
      handleCancel,
      handleCollapseChange,
      handleSaveNodePoolData,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-pool-wrapper {
  height: calc(100vh - 104px);
}
>>> .bk-resize-layout>.bk-resize-layout-aside:after {
  content: unset;
}
.node-pool {
  >>> .bk-resize-layout-aside {
    border-left: none !important;
  }
}
.aside {
  border-left: none;
  height: 100%;
  overflow: auto;
  background: #fff;
  >>> .content-wrapper {
    max-height: calc(100vh - 224px);
  }
}
.main {
  max-height: calc(100vh - 164px);
  overflow: auto;
  padding: 24px;
}
</style>
