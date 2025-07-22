<template>
  <bk-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    form-type="vertical"
    class="px-[20px] text-[14px]">
    <div class="font-bold mb-[10px]">{{ $t('serviceMesh.label.highAvailability') }}</div>
    <bk-form-item :label="$t('serviceMesh.label.replicas')">
      <bk-input
        class="w-[120px]"
        v-model="formData.highAvailability.replicaCount"
        type="number"
        :min="1"
        :precision="0"></bk-input>
    </bk-form-item>
    <!-- 资源 -->
    <div class="mt-[24px] mb-[6px]">{{ $t('serviceMesh.label.resource') }}</div>
    <div class="bg-[#F5F7FA] flex flex-wrap rounded-sm py-[12px] px-[16px]">
      <bk-form-item
        class="mr-[12px]"
        property="cpuRequest"
        error-display-type="normal"
        :label-width="120"
        :label="$t('serviceMesh.label.cpuRequest')"
        required>
        <div class="flex">
          <bk-input
            class="w-[84px]"
            v-model="resourceData.cpuRequest"
            type="number"
            :min="0"
            :precision="0"></bk-input>
          <span
            class="px-[4px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center text-[12px]">
            mCPUs</span>
        </div>
      </bk-form-item>
      <bk-form-item
        class="mr-[12px] !mt-0"
        property="cpuLimit"
        error-display-type="normal"
        :label-width="120"
        :label="$t('serviceMesh.label.cpuLimit')">
        <div class="flex">
          <bk-input
            class="w-[84px]"
            v-model="resourceData.cpuLimit"
            type="number"
            :min="0"
            :precision="0"></bk-input>
          <span
            class="px-[4px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center text-[12px]">
            mCPUs</span>
        </div>
      </bk-form-item>
      <bk-form-item
        class="mr-[12px] !mt-0"
        property="memoryRequest"
        error-display-type="normal"
        :label-width="120"
        :label="$t('serviceMesh.label.memoryRequest')"
        required>
        <div class="flex">
          <bk-input
            class="w-[84px]"
            v-model="resourceData.memoryRequest"
            type="number"
            :min="0"
            :precision="0"></bk-input>
          <span
            class="w-[36px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center text-[12px]">
            MB</span>
        </div>
      </bk-form-item>
      <bk-form-item
        :label-width="120"
        class="!mt-0"
        property="memoryLimit"
        error-display-type="normal"
        :label="$t('serviceMesh.label.memoryLimit')">
        <div class="flex">
          <bk-input
            class="w-[84px]"
            v-model="resourceData.memoryLimit"
            type="number"
            :min="0"
            :precision="0"></bk-input>
          <span
            class="w-[36px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center text-[12px]">
            MB</span>
        </div>
      </bk-form-item>
    </div>
    <!-- HPA -->
    <ContentSwitcher
      class="mt-[16px]"
      label="HPA"
      v-model="formData.highAvailability.autoscaleEnabled">
      <div class="flex">
        <bk-form-item
          class="mr-[24px] !mt-0"
          :label-width="250"
          :label="$t('serviceMesh.label.cpuPercent')">
          <div class="flex">
            <bk-input
              class="w-[220px]"
              v-model="formData.highAvailability.targetCPUAverageUtilizationPercent"
              type="number"
              :min="0"
              :max="100"></bk-input>
            <span class="w-[28px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center">%</span>
          </div>
        </bk-form-item>
        <bk-form-item
          class="mr-[24px] !mt-0"
          :label-width="120"
          :label="$t('serviceMesh.label.replicaMin')">
          <bk-input
            v-model="formData.highAvailability.autoscaleMin"
            type="number"
            :min="1"
            :max="100"
            :precision="0"></bk-input>
        </bk-form-item>
        <bk-form-item
          class="!mt-0"
          property="autoscaleMax"
          error-display-type="normal"
          :label-width="120"
          :label="$t('serviceMesh.label.replicaMax')">
          <bk-input
            v-model="formData.highAvailability.autoscaleMax"
            type="number"
            :min="1"
            :max="100"
            :precision="0"></bk-input>
        </bk-form-item>
      </div>
    </ContentSwitcher>
    <!-- 专属节点 -->
    <ContentSwitcher
      class="mt-[16px]"
      v-model="formData.highAvailability.dedicatedNode.enabled">
      <template #label>
        <span
          class="underline decoration-dashed underline-offset-4 decoration-[#979ba5] cursor-pointer"
          v-bk-tooltips="{
            content: $t('serviceMesh.tips.dedicatedNodeDesc'),
            placement: 'top',
          }">{{ $t('serviceMesh.label.dedicatedNode') }}</span>
      </template>
      <bk-form-item class="!mt-0" :label="$t('generic.label.tag')">
        <KeyValue v-model="formData.highAvailability.dedicatedNode.nodeLabels" />
      </bk-form-item>
    </ContentSwitcher>
    <bk-form-item
      :class="[
        'absolute bottom-0 left-0 w-full bg-white py-[10px] px-[40px] z-[999] border-t border-t-[#e6e6ec]'
      ]">
      <template v-if="!isEdit">
        <bk-button theme="default" @click="handlePre">{{ $t('generic.button.pre') }}</bk-button>
        <bk-button theme="primary" @click="handleCreate">{{ $t('generic.button.create') }}</bk-button>
      </template>
      <bk-button v-else theme="primary" @click="handleSave">{{ $t('generic.button.save') }}</bk-button>
      <bk-button theme="default" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
import { cloneDeep, mergeWith } from 'lodash';
import { ref, watch } from 'vue';

import ContentSwitcher from './content-switcher.vue';
import useMesh, { IMesh, ISidecar } from './use-mesh';

import $i18n from '@/i18n/i18n-setup';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';

type IParams = Pick<IMesh, 'highAvailability'>;


const props = defineProps({
  isEdit: {
    type: Boolean,
    default: false,
  },
  data: {
    type: Object,
    default: () => ({}),
  },
  active: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['pre', 'submit', 'cancel', 'save', 'change']);

const {
  extractNumbers,
  configData,
} = useMesh();

const resourceData = ref({
  cpuRequest: 0,
  cpuLimit: 0,
  memoryRequest: 0,
  memoryLimit: 0,
});
const formData = ref<IParams>({
  highAvailability: {
    autoscaleEnabled: true,
    autoscaleMin: 2,
    autoscaleMax: 5,
    replicaCount: 3,
    targetCPUAverageUtilizationPercent: 80,
    resourceConfig: {
      cpuRequest: '',
      cpuLimit: '',
      memoryRequest: '',
      memoryLimit: '',
    },
    dedicatedNode: {
      enabled: false,
      nodeLabels: {},
    },
  },
});
const rules = ref({
  cpuRequest: [
    {
      validator: () => resourceData.value.cpuRequest > 0,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
  cpuLimit: [
    {
      validator: () => resourceData.value.cpuLimit === 0
        || resourceData.value.cpuLimit >= resourceData.value.cpuRequest,
      message: $i18n.t('serviceMesh.tips.cpuValidate'),
      trigger: 'change',
    },
  ],
  memoryRequest: [
    {
      validator: () => resourceData.value.memoryRequest > 0,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
  memoryLimit: [
    {
      validator: () => resourceData.value.memoryLimit === 0
        || resourceData.value.memoryLimit >= resourceData.value.memoryRequest,
      message: $i18n.t('serviceMesh.tips.memoryValidate'),
      trigger: 'change',
    },
  ],
  autoscaleMax: [
    {
      validator: () => !formData.value.highAvailability.autoscaleEnabled
        || formData.value.highAvailability.autoscaleMax >= formData.value.highAvailability.autoscaleMin,
      message: $i18n.t('serviceMesh.tips.autoscaleMax'),
      trigger: 'change',
    },
  ],
});

function handlePre() {
  emits('pre');
}

const formRef = ref();
async function handleCreate() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    return;
  };
  if (!formData.value.highAvailability.dedicatedNode.enabled) {
    formData.value.highAvailability.dedicatedNode = {};
  }

  emits('submit', formData.value);
}
async function handleSave() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    return;
  };
  if (!formData.value.highAvailability.dedicatedNode.enabled) {
    formData.value.highAvailability.dedicatedNode = {};
  }

  emits('save', formData.value);
}
function handleCancel() {
  emits('cancel');
}

watch(resourceData, () => {
  const config = {
    cpuRequest: `${resourceData.value.cpuRequest}m`,
    cpuLimit: `${resourceData.value.cpuLimit}m`,
    memoryRequest: `${resourceData.value.memoryRequest}Mi`,
    memoryLimit: `${resourceData.value.memoryLimit}Mi`,
  };
  formData.value.highAvailability.resourceConfig = config;
}, { deep: true });

watch(() => props.active, () => {
  if (props.isEdit) {
    formData.value = cloneDeep(props.data) as IParams;
  }
}, { immediate: true });

watch(formData, () => {
  if (!props.isEdit) return;
  emits('change');
}, { deep: true });

// 默认数据
watch(configData, () => {
  let defaultData: Partial<ISidecar> = {};
  const { highAvailability } = configData.value;
  if (props.isEdit) {
    defaultData = formData.value.highAvailability.resourceConfig;
  } else {
    defaultData = highAvailability?.resourceConfig || {};
    formData.value.highAvailability = mergeWith(formData.value.highAvailability, highAvailability);
  }

  resourceData.value = {
    cpuRequest: extractNumbers(defaultData?.cpuRequest),
    cpuLimit: extractNumbers(defaultData?.cpuLimit),
    memoryRequest: extractNumbers(defaultData?.memoryRequest),
    memoryLimit: extractNumbers(defaultData?.memoryLimit),
  };
}, { immediate: true });
</script>

