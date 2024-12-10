<template>
  <div v-bkloading="{ isLoading }">
    <bk-form
      :model="metricData"
      :rules="metricDataRules"
      form-type="vertical"
      ref="formRef">
      <bk-form-item
        :label="$t('generic.label.name')"
        property="name"
        error-display-type="normal"
        required>
        <bk-input :disabled="isEdit" class="max-w-[50%]" v-model="metricData.name"></bk-input>
      </bk-form-item>
      <bk-form-item
        :label="$t('plugin.metric.label.service')"
        property="service_name"
        error-display-type="normal"
        required>
        <div class="flex">
          <ClusterSelect
            :cluster-type="['independent', 'managed', 'virtual']"
            v-model="clusterID"
            class="flex-1"
            disabled />
          <NamespaceSelect
            :disabled="isEdit"
            :cluster-id="clusterID"
            v-model="metricData.$namespaceId"
            class="flex-1 ml-[15px]" />
          <bcs-select
            class="flex-1 ml-[15px]"
            :loading="serviceLoading"
            :disabled="isEdit"
            v-model="metricData.service_name">
            <bcs-option
              v-for="item in serviceList"
              :key="item.metadata.name"
              :id="item.metadata.name"
              :name="item.metadata.name">
            </bcs-option>
          </bcs-select>
        </div>
      </bk-form-item>
      <template v-if="metricData.service_name">
        <bk-form-item
          :label="$t('plugin.metric.label.matchLabels')"
          property="selector"
          error-display-type="normal"
          required>
          <div v-if="Object.keys(labels).length">
            <keyValueSelector
              :labels="labels"
              :value="metricData.selector"
              @change="handleLabelChange" />
          </div>
          <div class="text-[12px]" v-else>
            {{ $t('plugin.metric.tips.noLabel') }}
            <bk-button
              text
              class="text-[12px]"
              @click="handleGotoService">{{ $t('plugin.metric.action.add') }}</bk-button>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('plugin.metric.label.portName')"
          property="port"
          error-display-type="normal"
          required>
          <bcs-select searchable v-model="metricData.port">
            <bcs-option
              v-for="item in ports"
              :key="item.name"
              :id="item.name"
              :name="item.name">
            </bcs-option>
          </bcs-select>
        </bk-form-item>
      </template>
      <bk-form-item
        :label="$t('plugin.metric.endpoints.path')"
        property="path"
        error-display-type="normal"
        required>
        <bk-input v-model="metricData.path"></bk-input>
      </bk-form-item>
      <bk-form-item :label="$t('plugin.metric.endpoints.params')">
        <KeyValue :min-item="0" v-model="metricData.params" />
      </bk-form-item>
      <div class="flex mt-[8px]">
        <bk-form-item
          :label="$t('plugin.metric.endpoints.interval')"
          property="interval"
          error-display-type="normal"
          required
          class="flex-1">
          <bcs-input type="number" v-model="metricData.interval"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('plugin.metric.sampleLimit')"
          property="sample_limit"
          error-display-type="normal"
          required
          class="!mt-[0px] flex-1 ml-[15px]">
          <bcs-input type="number" :max="100000" v-model="metricData.sample_limit"></bcs-input>
        </bk-form-item>
      </div>
    </bk-form>
    <div class="flex mt-[30px]">
      <bk-button
        :loading="saveLoading"
        theme="primary"
        @click="handleSubmit">{{ $t('generic.button.submit') }}</bk-button>
      <bk-button :disabled="saveLoading" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </div>
</template>
<script lang='ts' setup>
import { computed, onMounted, ref, watch } from 'vue';

import keyValueSelector from './key-value-selector.vue';
import useMetric, { IMetricData } from './use-metric';

import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';

const props = defineProps({
  data: {
    type: Object,
  },
});

const emits = defineEmits(['cancel', 'submit', 'change', 'init-data']);

const isEdit = computed(() => !!props.data && !!Object.keys(props.data).length);
const isLoading = ref(false);
const formRef = ref<any>(null);
const clusterID = ref('');// ClusterID是异步的
const metricData = ref<IMetricData & {
  $namespaceId: string
}>({
  $namespaceId: '',
  service_name: '',
  path: '', // 路径
  selector: {}, // 关联label
  interval: '30', // 采集周期
  port: '', // portName
  sample_limit: 10000, // 允许最大Sample数
  name: '', // 名称
  params: {}, // 参数
});
watch(metricData, () => {
  emits('change', metricData.value);
}, { deep: true });

const metricDataRules = ref({
  name: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('generic.validate.name'),
      trigger: 'blur',
      validator(val) {
        return /^[a-z][-a-z0-9]*$/.test(val);
      },
    },
  ],
  service_name: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  selector: [
    {
      message: $i18n.t('plugin.metric.tips.needLabel'),
      trigger: 'blur',
      validator: () => Object.keys(metricData.value.selector).length >= 1,
    },
  ],
  port: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  path: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  interval: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  sample_limit: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
});

const {
  serviceLoading,
  serviceList,
  handleGetServiceList,
  handleCreateServiceMonitor,
  handleUpdateServiceMonitor,
} = useMetric();

// 关联labels
const labels = computed(() => serviceList.value
  .find(item => item.metadata?.name === metricData.value.service_name)?.metadata?.labels || {});
// PortName列表
const ports = computed(() => serviceList.value
  .find(item => item.metadata?.name === metricData.value.service_name)?.spec?.ports || []);

watch(() => metricData.value.$namespaceId, () => {
  if (isEdit.value) return;
  metricData.value.service_name = '';
  handleGetServiceList(metricData.value.$namespaceId, clusterID.value);
});

// label变更
const handleLabelChange = (value) => {
  metricData.value.selector = value;
};
// 跳转service资源
const handleGotoService = () => {
  $router.push({
    name: 'dashboardNetworkService',
  });
};

// 提交保存
const saveLoading = ref(false);
const handleSubmit = async () => {
  const validate = await formRef.value?.validate().catch(() => false);
  if (!validate) return;

  let result = false;
  const params = {
    ...metricData.value,
    sample_limit: Number(metricData.value.sample_limit),
    interval: `${metricData.value.interval}s`,
    $clusterId: clusterID.value,
  };
  saveLoading.value = true;
  if (isEdit.value) {
    result = await handleUpdateServiceMonitor({
      ...params,
      $name: params.name, // name放在路由上面
    });
  } else {
    result = await handleCreateServiceMonitor(params);
  }
  saveLoading.value = false;
  result && emits('submit');
};
// 取消
const handleCancel = () => {
  emits('cancel');
};

onMounted(async () => {
  // 编辑模式
  if (isEdit.value) {
    const { metadata, spec = {} } = props.data || {};
    const endpoints = spec?.endpoints?.[0] || {};
    metricData.value = {
      name: metadata.name, // 名称
      $namespaceId: metadata.namespace,
      service_name: metadata?.labels?.['io.tencent.bcs.service_name'],
      sample_limit: spec?.sampleLimit, // 允许最大Sample数
      selector: spec?.selector?.matchLabels || {}, // 关联label
      path: endpoints?.path, // 路径
      port: endpoints?.port, // portName
      params: Object.keys(endpoints?.params || {}).reduce((pre, key) => {
        pre[key] = endpoints?.params?.[key]?.[0]; // 去数组第一个作为value
        return pre;
      }, {}), // 参数
      interval: String(parseInt(spec?.endpoints?.[0]?.interval || '0')), // 采集周期
    };
    isLoading.value = true;
    await handleGetServiceList(metadata.namespace, clusterID.value);
    isLoading.value = false;
    emits('init-data');
  }
});
</script>
