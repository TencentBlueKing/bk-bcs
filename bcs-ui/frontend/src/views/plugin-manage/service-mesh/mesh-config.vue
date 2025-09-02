<template>
  <bk-form
    ref="formRef"
    :model="formData"
    :rules="rules"
    form-type="vertical"
    class="px-[20px] text-[14px]">
    <div class="font-bold mb-[10px]">{{ $t('serviceMesh.label.featureConfig') }}</div>
    <div class="flex">
      <bk-form-item
        class="mr-[10px]"
        :label="$t('serviceMesh.label.egressMode')"
        :desc="{
          allowHTML: true,
          content: '#egressMode',
        }">
        <bk-radio-group v-model="formData.featureConfigs.outboundTrafficPolicy.value">
          <bk-radio value="ALLOW_ANY" class="text-[12px]">ALLOW_ANY</bk-radio>
          <bk-radio value="REGISTRY_ONLY" class="text-[12px] !ml-[20px]">REGISTRY_ONLY</bk-radio>
        </bk-radio-group>
        <div id="egressMode" class="max-w-[380px] flex flex-col">
          <span>{{ $t('serviceMesh.tips.egressModeDesc.title') }}</span>
          <span>{{ $t('serviceMesh.tips.egressModeDesc.registryOnly') }}</span>
          <span>{{ $t('serviceMesh.tips.egressModeDesc.allowAny') }}</span>
        </div>
      </bk-form-item>
      <bk-form-item
        class="!mt-0"
        :label="$t('serviceMesh.label.sidecarOn')"
        :desc="{
          allowHTML: true,
          content: '#sidecarOn',
        }">
        <bcs-switcher
          size="large"
          v-model="formData.featureConfigs.holdApplicationUntilProxyStarts.value"
          true-value="true"
          false-value="false"></bcs-switcher>
        <div id="sidecarOn" class="max-w-[280px]">
          {{ $t('serviceMesh.tips.sidecarOnDesc') }}
        </div>
      </bk-form-item>
      <bk-form-item
        class="!mt-0"
        :label="$t('serviceMesh.label.sidecarOff')"
        :desc="{
          allowHTML: true,
          content: '#sidecarOff',
        }">
        <bcs-switcher
          size="large"
          v-model="formData.featureConfigs.exitOnZeroActiveConnections.value"
          true-value="true"
          false-value="false"></bcs-switcher>
        <div id="sidecarOff" class="max-w-[280px]">
          {{ $t('serviceMesh.tips.sidecarOffDesc') }}
        </div>
      </bk-form-item>
    </div>
    <!-- Sidecar 资源 -->
    <div class="mt-[14px] mb-[6px] text-[12px]">{{ $t('serviceMesh.label.sidecarResource') }}</div>
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
            v-model="sidecarConfig.cpuRequest"
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
            v-model="sidecarConfig.cpuLimit"
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
            v-model="sidecarConfig.memoryRequest"
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
            v-model="sidecarConfig.memoryLimit"
            type="number"
            :min="0"
            :precision="0"></bk-input>
          <span
            class="w-[36px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center text-[12px]">
            MB</span>
        </div>
      </bk-form-item>
    </div>
    <!-- DNS 代理转发 -->
    <div class="flex mt-[14px]">
      <bk-form-item
        class="!mt-0"
        :label="$t('serviceMesh.label.dnsCapture')"
        :desc="{
          allowHTML: true,
          content: '#dnsOn',
        }">
        <bcs-switcher
          size="large"
          v-model="formData.featureConfigs.istioMetaDnsCapture.value"
          true-value="true"
          false-value="false"></bcs-switcher>
        <div id="dnsOn" class="max-w-[280px]">
          {{ $t('serviceMesh.tips.dnsCaptureDesc') }}
        </div>
      </bk-form-item>
      <bk-form-item
        class="!mt-0"
        :label="$t('serviceMesh.label.dnsAutoAllocate')"
        :desc="{
          allowHTML: true,
          content: '#IPOn',
        }">
        <bcs-switcher
          size="large"
          v-model="formData.featureConfigs.istioMetaDnsAutoAllocate.value"
          true-value="true"
          false-value="false"></bcs-switcher>
        <div id="IPOn" class="max-w-[280px]">
          <div>{{ $t('serviceMesh.tips.dnsAutoAllocateDesc') }}</div>
          <bk-button text theme="primary">{{ $t('generic.button.detail1') }}</bk-button>
        </div>
      </bk-form-item>
    </div>
    <!-- 排除 IP 范围 -->
    <bk-form-item
      class="mt-[10px]"
      :label="$t('serviceMesh.label.excludeIPs')"
      :desc="$t('serviceMesh.tips.excludeIPsDesc')">
      <bcs-tag-input
        allow-create
        has-delete-icon
        free-paste
        allow-auto-match
        v-model="excludeIPs"
        :placeholder="$t('serviceMesh.placeholder.ips')">
      </bcs-tag-input>
    </bk-form-item>
    <!-- HTTP/1.0 -->
    <bk-form-item
      class="!mt-[14px]"
      :label="$t('serviceMesh.label.http10')"
      :desc="$t('serviceMesh.tips.http10Desc')">
      <bcs-switcher
        size="large"
        v-model="formData.featureConfigs.istioMetaHttp10.value"
        true-value="true"
        false-value="false"></bcs-switcher>
    </bk-form-item>
    <bcs-divider class="!mt-[24px] !mb-[8px]"></bcs-divider>
    <div class="font-bold mb-[10px] text-[14px]">{{ $t('serviceMesh.label.observability') }}</div>
    <!-- 指标采集 -->
    <ContentSwitcher
      class="mt-[16px]"
      :label="$t('serviceMesh.label.metrics')"
      v-model="formData.observabilityConfig.metricsConfig.metricsEnabled">
      <div class="grid grid-cols-2">
        <bk-form-item
          class="!mt-0"
          :label="$t('serviceMesh.label.controlPlane')">
          <bcs-switcher
            size="large"
            v-model="formData.observabilityConfig.metricsConfig.controlPlaneMetricsEnabled"></bcs-switcher>
        </bk-form-item>
        <bk-form-item
          class="!mt-0"
          :label="$t('serviceMesh.label.dataPlane')"
          :desc="{
            allowHTML: true,
            content: '#dataOn',
          }">
          <bcs-switcher
            size="large"
            v-model="formData.observabilityConfig.metricsConfig.dataPlaneMetricsEnabled"></bcs-switcher>
          <div id="dataOn" class="max-w-[222px]">
            {{ $t('serviceMesh.tips.dataPlaneDesc') }}
          </div>
        </bk-form-item>
      </div>
    </ContentSwitcher>
    <!-- 日志输出 -->
    <ContentSwitcher
      class="mt-[16px]"
      :label="$t('serviceMesh.label.logOutput')"
      v-model="formData.observabilityConfig.logCollectorConfig.enabled">
      <bk-form-item :label="$t('serviceMesh.label.logType')">
        <bk-radio-group v-model="formData.observabilityConfig.logCollectorConfig.accessLogEncoding">
          <bk-radio value="TEXT" class="text-[12px]">TEXT</bk-radio>
          <bk-radio value="JSON" class="text-[12px]">JSON</bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item :label="$t('serviceMesh.label.logFormat')" error-display-type="normal" property="accessLogFormat">
        <bk-input
          type="textarea"
          v-model="formData.observabilityConfig.logCollectorConfig.accessLogFormat"
          :show-word-limit="false"
          :maxlength="2000"></bk-input>
      </bk-form-item>
    </ContentSwitcher>
    <!-- 全链路追踪 -->
    <ContentSwitcher
      class="mt-[16px]"
      :label="$t('serviceMesh.label.tracing')"
      v-model="formData.observabilityConfig.tracingConfig.enabled">
      <bk-form-item
        :label="$t('serviceMesh.label.sampleRate')"
        :desc="{
          allowHTML: true,
          content: '#sampleRate',
        }"
        property="traceSamplingPercent"
        error-display-type="normal"
        required>
        <div class="flex w-[100px]">
          <bk-input
            class="w-[92px]"
            v-model="formData.observabilityConfig.tracingConfig.traceSamplingPercent"
            type="number"
            :min="0"
            :max="100"></bk-input>
          <span class="w-[28px] h-full bg-[#FAFBFD] border-[#c4c6cc] border rounded-sm ml-[-2px] text-center">%</span>
        </div>
        <div id="sampleRate" class="max-w-[350px]">
          {{ $t('serviceMesh.tips.sampleRateDesc') }}
        </div>
      </bk-form-item>
      <bk-form-item
        :label="$t('serviceMesh.label.endpoint')"
        :desc="$t('serviceMesh.tips.endpointDesc')"
        property="endpoint"
        error-display-type="normal"
        required>
        <bk-input v-model="formData.observabilityConfig.tracingConfig.endpoint"></bk-input>
      </bk-form-item>
      <bk-form-item
        :label="$t('serviceMesh.label.bkToken')"
        :desc="$t('serviceMesh.tips.bkTokenDesc')"
        property="bkToken"
        error-display-type="normal"
        required>
        <bk-input v-model="formData.observabilityConfig.tracingConfig.bkToken"></bk-input>
      </bk-form-item>
    </ContentSwitcher>
    <div
      :class="[
        'absolute bottom-0 left-0 w-full bg-white py-[10px] px-[40px] z-[999] border-t border-t-[#e6e6ec]'
      ]">
      <template v-if="!isEdit">
        <bk-button theme="default" @click="handlePre">{{ $t('generic.button.pre') }}</bk-button>
        <bk-button theme="primary" @click="handleNext">{{ $t('generic.button.next') }}</bk-button>
      </template>
      <bk-button v-else theme="primary" @click="handleSave">{{ $t('generic.button.save') }}</bk-button>
      <bk-button theme="default" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { ref, watch } from 'vue';

import ContentSwitcher from './content-switcher.vue';
import useMesh, { IMesh, ISidecar } from './use-mesh';

import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';

type IParams = Pick<IMesh, 'sidecarResourceConfig' | 'featureConfigs' | 'observabilityConfig'>;

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

const emits = defineEmits(['pre', 'next', 'cancel', 'save', 'change']);

const {
  extractNumbers,
  configData,
} = useMesh();
const { focusOnErrorField } = useFocusOnErrorField();

const excludeIPs = ref<string[]>([]);
const sidecarConfig = ref<Record<keyof ISidecar, number>>({
  cpuRequest: 0,
  cpuLimit: 0,
  memoryRequest: 0,
  memoryLimit: 0,
});
const formData = ref<IParams>({
  sidecarResourceConfig: { // Sidecar 资源配置
    cpuRequest: '',
    cpuLimit: '',
    memoryRequest: '',
    memoryLimit: '',
  },
  featureConfigs: { // 特性
    outboundTrafficPolicy: { // 出站流量策略
      value: 'ALLOW_ANY',
      defaultValue: 'ALLOW_ANY',
    },
    holdApplicationUntilProxyStarts: { // 应用等待 sidecar 启动
      value: 'false',
      defaultValue: 'false',
    },
    exitOnZeroActiveConnections: { // 无活动连接时退出
      value: 'false',
      defaultValue: 'false',
    },
    excludeIPRanges: {
      value: '',
      defaultValue: '',
    }, // 排除IP范围
    istioMetaDnsCapture: { // DNS转发
      value: 'true',
      defaultValue: 'true',
    },
    istioMetaDnsAutoAllocate: { // 自动分配IP
      value: 'true',
      defaultValue: 'true',
    },
    istioMetaHttp10: { // 是否支持HTTP/1.0
      value: 'false',
      defaultValue: 'false',
    },
  },
  observabilityConfig: { // 可观测性
    metricsConfig: {
      metricsEnabled: true,
      controlPlaneMetricsEnabled: true,
      dataPlaneMetricsEnabled: true,
    },
    logCollectorConfig: {
      enabled: true,
      accessLogEncoding: 'JSON',
      accessLogFormat: '',
    },
    tracingConfig: {
      enabled: true,
      traceSamplingPercent: 1,
      endpoint: '',
      bkToken: '',
    },
  },
});
const endpointReg = new RegExp('^((https?://)?([a-zA-Z0-9.-]+)(:[0-9]+)?(/[a-zA-Z0-9._-]*)*)$');
const rules = ref({
  traceSamplingPercent: [
    {
      validator: () => !formData.value.observabilityConfig.tracingConfig.enabled
          || !!formData.value.observabilityConfig.tracingConfig.traceSamplingPercent,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  endpoint: [
    {
      validator: () => !formData.value.observabilityConfig.tracingConfig.enabled
        || formData.value.observabilityConfig.tracingConfig.endpoint,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      validator: () => !formData.value.observabilityConfig.tracingConfig.enabled
          || (formData.value.observabilityConfig.tracingConfig.endpoint
            && endpointReg.test(formData.value.observabilityConfig.tracingConfig.endpoint)),
      message: $i18n.t('serviceMesh.tips.endpointValidate'),
      trigger: 'blur',
    },
  ],
  bkToken: [
    {
      validator: () => !formData.value.observabilityConfig.tracingConfig.enabled
          || !!formData.value.observabilityConfig.tracingConfig.bkToken,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  accessLogFormat: [
    {
      validator: () => !formData.value.observabilityConfig.logCollectorConfig.enabled
          || !formData.value?.observabilityConfig?.logCollectorConfig?.accessLogFormat
          || (formData.value?.observabilityConfig?.logCollectorConfig?.accessLogFormat?.length
            && formData.value.observabilityConfig.logCollectorConfig.accessLogFormat.length <= 2000
          ),
      message: $i18n.t('serviceMesh.tips.maxlength'),
      trigger: 'change',
    },
  ],
  cpuRequest: [
    {
      validator: () => sidecarConfig.value.cpuRequest > 0,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
  cpuLimit: [
    {
      validator: () => sidecarConfig.value.cpuLimit === 0
        || sidecarConfig.value.cpuLimit >= sidecarConfig.value.cpuRequest,
      message: $i18n.t('serviceMesh.tips.cpuValidate'),
      trigger: 'change',
    },
  ],
  memoryRequest: [
    {
      validator: () => sidecarConfig.value.memoryRequest > 0,
      message: $i18n.t('generic.validate.required'),
      trigger: 'change',
    },
  ],
  memoryLimit: [
    {
      validator: () => sidecarConfig.value.memoryLimit === 0
        || sidecarConfig.value.memoryLimit >= sidecarConfig.value.memoryRequest,
      message: $i18n.t('serviceMesh.tips.memoryValidate'),
      trigger: 'change',
    },
  ],
});

function handlePre() {
  emits('pre');
}
const formRef = ref();
async function handleNext() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    focusOnErrorField();
    return;
  };

  formData.value.featureConfigs.excludeIPRanges.value = excludeIPs.value.join(',');
  emits('next', formData.value);
}

async function handleSave() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    focusOnErrorField();
    return;
  };

  formData.value.featureConfigs.excludeIPRanges.value = excludeIPs.value.join(',');
  emits('save', formData.value);
}

function handleCancel() {
  emits('cancel');
}

watch(sidecarConfig, () => {
  const config = {
    cpuRequest: `${sidecarConfig.value.cpuRequest}m`,
    cpuLimit: `${sidecarConfig.value.cpuLimit}m`,
    memoryRequest: `${sidecarConfig.value.memoryRequest}Mi`,
    memoryLimit: `${sidecarConfig.value.memoryLimit}Mi`,
  };
  formData.value.sidecarResourceConfig = config;
}, { deep: true });

watch(() => props.active, () => {
  if (props.isEdit) {
    formData.value = cloneDeep(props.data) as IParams;
    // excludeIPs
    excludeIPs.value = formData.value.featureConfigs?.excludeIPRanges?.value?.split(',')?.filter(item => !!item) || [];
    // 避免空异常
    if (formData.value.observabilityConfig.metricsConfig?.metricsEnabled === undefined) {
      formData.value.observabilityConfig.metricsConfig = {};
    }
    if (formData.value.observabilityConfig.logCollectorConfig?.enabled === undefined) {
      formData.value.observabilityConfig.logCollectorConfig = {};
    }
    if (formData.value.observabilityConfig.tracingConfig?.enabled === undefined) {
      formData.value.observabilityConfig.tracingConfig = {};
    }
  }
}, { immediate: true });

watch(formData, () => {
  if (!props.isEdit) return;
  emits('change');
}, { deep: true });

// 默认数据
watch(configData, () => {
  let defaultData: Partial<ISidecar> = {};
  const { sidecarResourceConfig, featureConfigs, observabilityConfig } = configData.value;
  if (props.isEdit) {
    defaultData = formData.value.sidecarResourceConfig;
  } else {
    defaultData = sidecarResourceConfig || {};
  }
  sidecarConfig.value = {
    cpuRequest: extractNumbers(defaultData?.cpuRequest),
    cpuLimit: extractNumbers(defaultData?.cpuLimit),
    memoryRequest: extractNumbers(defaultData?.memoryRequest),
    memoryLimit: extractNumbers(defaultData?.memoryLimit),
  };

  if (props.isEdit) return;
  // 等效于 for in，但更安全
  Object.entries(featureConfigs || {}).forEach(([key, value]) => {
    formData.value.featureConfigs[key].value = value.defaultValue || '';
  });
  Object.entries(observabilityConfig || {}).forEach(([key, value]) => {
    formData.value.observabilityConfig[key] = {
      ...(value || {}),
    };
  });
}, { immediate: true });
</script>
