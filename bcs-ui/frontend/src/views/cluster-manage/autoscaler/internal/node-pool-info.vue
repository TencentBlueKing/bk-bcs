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
        <ActionDoc type="autoscaler" class="aside" :title="$t('cluster.ca.nodePool.title.initConfig1')" />
      </template>
      <template #main>
        <div class="main" ref="nodePoolInfoRef">
          <FormGroup :title="$t('generic.title.basicInfo1')" :allow-toggle="false">
            <BasicPoolInfo
              :schema="schema"
              :default-values="defaultValues"
              :is-edit="isEdit"
              :cluster="cluster"
              ref="basicInfoRef">
            </BasicPoolInfo>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('cluster.nodeTemplate.kubelet.title.argsConfig1')" :allow-toggle="false">
            <KubeletParams v-model="nodePoolInfoData.nodeTemplate.extraArgs.kubelet" ref="kubeletRef"></KubeletParams>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('cluster.ca.nodePool.create.scaleInitConfig.title')" :allow-toggle="false">
            <p>
              <span
                class="bcs-border-tips"
                v-bk-tooltips="$t('cluster.ca.nodePool.create.scaleInitConfig.preInit.tips')">
                {{$t('cluster.nodeTemplate.label.preInstall.title')}}
              </span>
            </p>
            <bcs-input
              type="textarea"
              class="mt10"
              :rows="6"
              :placeholder="$t('cluster.ca.nodePool.create.scaleInitConfig.placeholder')"
              v-model="nodePoolInfoData.nodeTemplate.preStartUserScript">
            </bcs-input>
            <p class="mt-[32px]">{{$t('cluster.nodeTemplate.label.postInstall.title')}}</p>
            <div class="mt-[10px]">
              <bcs-select class="max-w-[524px]" :clearable="false" v-model="scaleOutPostActionType">
                <bcs-option id="simple" :name="$t('cluster.nodeTemplate.label.postInstall.type.scripts')"></bcs-option>
                <bcs-option id="complex" :name="$t('cluster.nodeTemplate.label.postInstall.type.sops')"></bcs-option>
              </bcs-select>
              <bcs-input
                type="textarea"
                class="mt10"
                :rows="6"
                :placeholder="$t('cluster.ca.nodePool.create.scaleInitConfig.placeholder')"
                v-if="scaleOutPostActionType === 'simple'"
                v-model="nodePoolInfoData.nodeTemplate.userScript">
              </bcs-input>
              <BkSops
                class="mt10"
                actions-key="postActions"
                :cluster-id="cluster.clusterID"
                :addons="nodePoolInfoData.nodeTemplate.scaleOutExtraAddons"
                :allow-skip-when-failed="allowSkipScaleOutWhenFailed"
                ref="scaleOutRef"
                v-else>
              </BkSops>
            </div>
            <p class="mt-[10px]">{{$t('generic.label.whenFailed')}}</p>
            <bk-radio-group class="mt-[8px]" v-model="allowSkipScaleOutWhenFailed">
              <bk-radio :value="false" class="text-[12px]">
                {{ $t('cluster.ca.nodePool.create.scaleInitConfig.postInit.errorHandling.radio1') }}
                <span class="text-[#699DF4]">
                  [{{ $t('cluster.ca.nodePool.create.scaleInitConfig.label.keyStepsRecommended') }}]
                </span>
                <i
                  class="bk-icon icon-info-circle"
                  v-bk-tooltips="$t('cluster.ca.nodePool.create.scaleInitConfig.postInit.errorHandling.tips1')">
                </i>
              </bk-radio>
              <bk-radio :value="true" class="text-[12px]">
                {{ $t('cluster.ca.nodePool.create.scaleInitConfig.postInit.errorHandling.radio2') }}
                <span class="text-[#979BA5]">
                  [{{ $t('cluster.ca.nodePool.create.scaleInitConfig.label.unKeyStepsRecommended') }}]
                </span>
                <i
                  class="bk-icon icon-info-circle"
                  v-bk-tooltips="$t('cluster.ca.nodePool.create.scaleInitConfig.postInit.errorHandling.tips2')">
                </i>
              </bk-radio>
            </bk-radio-group>
          </FormGroup>
          <div class="px-[16px]"><bcs-divider class="!my-[0px]"></bcs-divider></div>
          <FormGroup :title="$t('cluster.ca.nodePool.create.scaleInitConfig.scaleInPreScript')" :allow-toggle="false">
            <bcs-select class="max-w-[524px]" :clearable="false" v-model="scaleInPreActionType">
              <bcs-option id="simple" :name="$t('cluster.nodeTemplate.label.postInstall.type.scripts')"></bcs-option>
              <bcs-option id="complex" :name="$t('cluster.nodeTemplate.label.postInstall.type.sops')"></bcs-option>
            </bcs-select>
            <bcs-input
              type="textarea"
              class="mt10"
              :rows="6"
              :placeholder="$t('cluster.ca.nodePool.create.scaleInitConfig.placeholder')"
              v-if="scaleInPreActionType === 'simple'"
              v-model="nodePoolInfoData.nodeTemplate.scaleInPreScript">
            </bcs-input>
            <BkSops
              class="mt10"
              actions-key="preActions"
              :addons="nodePoolInfoData.nodeTemplate.scaleInExtraAddons"
              :cluster-id="cluster.clusterID"
              :allow-skip-when-failed="allowSkipScaleInWhenFailed"
              ref="scaleInRef"
              v-else>
            </BkSops>
            <p class="mt-[10px]">{{$t('generic.label.whenFailed')}}</p>
            <bk-radio-group class="mt-[8px]" v-model="allowSkipScaleInWhenFailed">
              <bk-radio :value="false" class="text-[12px]">
                {{ $t('cluster.ca.nodePool.create.scaleInitConfig.recycleClean.errorHandling.radio1') }}
                <span class="text-[#699DF4]">
                  [{{ $t('cluster.ca.nodePool.create.scaleInitConfig.label.keyStepsRecommended') }}]
                </span>
                <i
                  class="bk-icon icon-info-circle"
                  v-bk-tooltips="$t('cluster.ca.nodePool.create.scaleInitConfig.recycleClean.errorHandling.tips1')">
                </i>
              </bk-radio>
              <bk-radio :value="true" class="text-[12px]">
                {{ $t('cluster.ca.nodePool.create.scaleInitConfig.recycleClean.errorHandling.radio2') }}
                <span class="text-[#979BA5]">
                  [{{ $t('cluster.ca.nodePool.create.scaleInitConfig.label.unKeyStepsRecommended') }}]
                </span>
                <i
                  class="bk-icon icon-info-circle"
                  v-bk-tooltips="$t('cluster.ca.nodePool.create.scaleInitConfig.recycleClean.errorHandling.tips2')">
                </i>
              </bk-radio>
            </bk-radio-group>
          </FormGroup>
        </div>
      </template>
    </bcs-resize-layout>
    <div class="bcs-fixed-footer" v-if="showFooter">
      <bcs-button @click="handlePre">{{$t('generic.button.pre')}}</bcs-button>
      <bcs-button
        theme="primary"
        :loading="saveLoading"
        class="ml10"
        @click="handleSaveNodePoolData">
        {{isEdit ? $t('cluster.ca.nodePool.create.button.save') : $t('cluster.ca.nodePool.records.taskType.create')}}
      </bcs-button>
      <bk-button class="ml10" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, toRefs } from 'vue';

import BkSops from '../bk-sops.vue';
import KubeletParams from '../kubelet-params.vue';

import BasicPoolInfo from './basic-pool-info.vue';

import { mergeDeep } from '@/common/util';
import $router from '@/router';
import FormGroup from '@/views/cluster-manage/add/common/form-group.vue';
import ActionDoc from '@/views/cluster-manage/components/action-doc.vue';;

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
        allowSkipScaleOutWhenFailed: defaultValues.value?.nodeTemplate?.allowSkipScaleOutWhenFailed,
        allowSkipScaleInWhenFailed: defaultValues.value?.nodeTemplate?.allowSkipScaleInWhenFailed,
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
        nodePoolInfoData.value.nodeTemplate.allowSkipScaleOutWhenFailed = allowSkipScaleOutWhenFailed.value;
      }

      // 处理缩容前置脚本参数
      if (scaleInPreActionType.value === 'complex') {
        nodePoolInfoData.value.nodeTemplate.scaleInPreScript = '';
        nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons = scaleInRef.value?.bkSopsData;
      } else {
        nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons = {};
        nodePoolInfoData.value.nodeTemplate.allowSkipScaleInWhenFailed = allowSkipScaleInWhenFailed.value;
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

    // 脚本失败后能够跳过配置
    const allowSkipScaleOutWhenFailed = ref(false);
    const allowSkipScaleInWhenFailed = ref(true);

    onMounted(() => {
      scaleOutPostActionType.value = nodePoolInfoData.value.nodeTemplate.scaleOutExtraAddons?.postActions?.length ? 'complex' : 'simple';
      scaleInPreActionType.value = nodePoolInfoData.value.nodeTemplate.scaleInExtraAddons?.preActions?.length ? 'complex' : 'simple';
      if (scaleOutPostActionType.value === 'complex') {
        const addons = nodePoolInfoData.value.nodeTemplate?.scaleOutExtraAddons;
        const plugin = addons?.postActions?.[0];// 取postAction第1个（目前只支持一个），但是后端设计为了数组
        allowSkipScaleOutWhenFailed.value = !!addons?.plugins?.[plugin]?.allowSkipWhenFailed;
      } else {
        allowSkipScaleOutWhenFailed.value = !!nodePoolInfoData.value.nodeTemplate.allowSkipScaleOutWhenFailed;
      }

      if (scaleInPreActionType.value === 'complex') {
        const addons = nodePoolInfoData.value.nodeTemplate?.scaleInExtraAddons;
        const plugin = addons?.preActions?.[0];// 取preAction第1个（目前只支持一个），但是后端设计为了数组
        allowSkipScaleInWhenFailed.value = !!addons?.plugins?.[plugin]?.allowSkipWhenFailed;
      } else {
        allowSkipScaleInWhenFailed.value = nodePoolInfoData.value.nodeTemplate.allowSkipScaleInWhenFailed !== undefined
          ? !!nodePoolInfoData.value.nodeTemplate.allowSkipScaleInWhenFailed
          : true;
      }
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
      allowSkipScaleOutWhenFailed,
      allowSkipScaleInWhenFailed,
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
