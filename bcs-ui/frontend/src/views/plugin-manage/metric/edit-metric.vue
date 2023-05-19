<template>
  <div v-bkloading="{ isLoading }">
    <bk-form
      :model="metricData"
      :rules="metricDataRules"
      form-type="vertical"
      ref="formRef">
      <bk-form-item
        :label="$t('名称')"
        property="name"
        error-display-type="normal"
        required>
        <bk-input :disabled="isEdit" class="max-w-[50%]" v-model="metricData.name"></bk-input>
      </bk-form-item>
      <bk-form-item
        :label="$t('选择Service')"
        property="service_name"
        error-display-type="normal"
        required>
        <div class="flex">
          <ClusterSelect v-model="clusterID" class="flex-1" disabled />
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
          :label="$t('选择关联Label')"
          property="selector"
          error-display-type="normal"
          required>
          <div v-if="Object.keys(labels).length">
            <div
              class="flex items-center mb-[10px]"
              v-for="key in Object.keys(labels)" :key="key">
              <bk-input class="w-[230px]" :value="key" disabled></bk-input>
              <span class="mx-[15px] text-[#c3cdd7]">=</span>
              <bk-input class="w-[230px]" :value="labels[key]" disabled></bk-input>
              <bk-checkbox
                class="ml-[10px]"
                :value="(key in metricData.selector)"
                @change="handleLabelChange(key, labels[key])">
              </bk-checkbox>
            </div>
          </div>
          <div class="text-[12px]" v-else>
            {{ $t('当前Service没有设置Labels') }}
            <bk-button text class="text-[12px]" @click="handleGotoService">{{ $t('前往添加') }}</bk-button>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('选择PortName')"
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
        :label="$t('Metric路径')"
        property="path"
        error-display-type="normal"
        required>
        <bk-input v-model="metricData.path"></bk-input>
      </bk-form-item>
      <bk-form-item :label="$t('Metric参数')">
        <KeyValue :min-item="0" v-model="metricData.params" />
      </bk-form-item>
      <div class="flex">
        <bk-form-item
          :label="$t('采集周期（秒）')"
          property="interval"
          error-display-type="normal"
          required
          class="flex-1">
          <bcs-input type="number" v-model="metricData.interval"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('允许最大Sample')"
          property="sample_limit"
          error-display-type="normal"
          required
          class="!mt-[0px] flex-1 ml-[15px]">
          <bcs-input type="number" :max="100000" v-model="metricData.sample_limit"></bcs-input>
        </bk-form-item>
      </div>
    </bk-form>
    <div class="flex mt-[30px]">
      <bk-button :loading="saveLoading" theme="primary" @click="handleSubmit">{{ $t('提交') }}</bk-button>
      <bk-button :disabled="saveLoading" @click="handleCancel">{{ $t('取消') }}</bk-button>
    </div>
  </div>
</template>
<script lang='ts' setup>
import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import useMetric, { IMetricData } from './use-metric';
import { computed, onMounted, ref, watch } from 'vue';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';

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
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
  service_name: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
  selector: [
    {
      message: $i18n.t('至少关联一个Label'),
      trigger: 'blur',
      validator: () => Object.keys(metricData.value.selector).length >= 1,
    },
  ],
  port: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
  path: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
  interval: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
  ],
  sample_limit: [
    {
      required: true,
      message: $i18n.t('必填项'),
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
const handleLabelChange = (key, value) => {
  if (key in metricData.value.selector) {
    delete metricData.value.selector[key];
  } else {
    metricData.value.selector[key] = value;
  }
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
