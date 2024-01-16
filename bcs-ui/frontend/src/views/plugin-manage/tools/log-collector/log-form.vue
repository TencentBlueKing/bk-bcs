<template>
  <bk-form class="log-form" :label-width="labelWidth" :rules="formDataRules" :model="formData" ref="formRef">
    <bk-form-item
      :label="$t('logCollector.label.displayName')"
      property="display_name"
      error-display-type="normal"
      required>
      <bcs-input
        v-model="formData.display_name"
        :maxlength="64"
        show-word-limit>
      </bcs-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('logCollector.label.name')"
      property="name"
      error-display-type="normal"
      required>
      <bcs-input v-model="formData.name" :disabled="isEdit"></bcs-input>
      <div class="flex items-center text-[#979BA5] h-[20px]">
        <i class="bk-icon icon-info-circle text-[#979BA5] text-[14px]"></i>
        <i18n path="logCollector.tips.name" class="ml-[4px]">
          <span class="text-[#FF9C01]">{{ $t('logCollector.tips.disabledEdit') }}</span>
        </i18n>
      </div>
    </bk-form-item>
    <bk-form-item :label="$t('logCollector.label.logType.text')" required>
      <div class="bk-button-group">
        <bk-button
          :class="['min-w-[106px]', { 'is-selected': logType === '' }]"
          @click="handleChangeLogType('')">
          {{ $t('logCollector.label.logType.line') }}
        </bk-button>
        <bk-button
          :class="['min-w-[106px]', { 'is-selected': logType === 'multiline' }]"
          @click="handleChangeLogType('multiline')">
          {{ $t('logCollector.label.logType.multiline') }}
        </bk-button>
      </div>
    </bk-form-item>
    <bk-form-item :label="$t('logCollector.label.configInfo')">
      <div class="border border-[#DCDEE5] border-solid p-[16px] bg-[#FAFBFD]">
        <div class="border border-[#EAEBF0] border-solid px-[12px] py-[12px] bg-[#fff] rounded-sm">
          <Row class="mb-[6px] px-[12px]">
            <template #left>
              <span class="font-bold">{{ $t('logCollector.title.selectRange.text') }}</span>
              <span class="ml-[12px] flex-1 bcs-ellipsis">
                <i
                  class="bk-icon icon-info-circle text-[#979BA5] text-[14px]"
                  v-bk-tooltips="$t('logCollector.title.selectRange.desc')">
                </i>
              </span>
            </template>
          </Row>
          <bk-form-item
            :label="$t('k8s.namespace')"
            property="namespaces"
            error-display-type="normal"
            class="vertical-form-item px-[12px]"
            required>
            <bcs-select
              v-model="formData.rule.config.namespaces"
              :loading="namespaceLoading"
              searchable
              :clearable="false"
              multiple
              :disabled="fromOldRule"
              selected-style="checkbox"
              @selected="handleNsChange">
              <bcs-option key="all" id="" :name="$t('plugin.tools.all')" v-if="!curCluster.is_shared"></bcs-option>
              <bcs-option
                v-for="option in namespaceList"
                :key="option.name"
                :id="option.name"
                :name="option.name">
              </bcs-option>
            </bcs-select>
          </bk-form-item>
          <LogPanel
            v-if="configRangeMap['label']"
            :disabled="fromOldRule"
            :title="$t('logCollector.button.addRange.label.text')"
            :desc="$t('logCollector.button.addRange.label.tips')"
            @delete="handleDeleteRange('label')">
            <bk-form-item
              :label-width="0.1"
              property="match_labels"
              error-display-type="normal"
              class="hide-form-label">
              <ContainerLabel
                :value="formData.rule.config.label_selector.match_labels"
                :disabled="fromOldRule"
                @change="handleMatchLabelsChange" />
            </bk-form-item>
          </LogPanel>
          <LogPanel
            :disabled="fromOldRule"
            :title="$t('logCollector.button.addRange.workload')"
            v-if="configRangeMap['workload']"
            @delete="handleDeleteRange('workload')">
            <div class="flex items-center">
              <div class="w-[200px] flex">
                <span class="bcs-form-prefix bg-[#f5f7fa]">{{$t('plugin.tools.appType')}}</span>
                <bcs-select
                  class="flex-1 bg-[#fff]"
                  :clearable="false"
                  v-model="formData.rule.config.container.workload_type">
                  <bcs-option v-for="item in kindList" :key="item" :id="item" :name="item"></bcs-option>
                </bcs-select>
              </div>
              <div class="flex-1 flex">
                <span class="bcs-form-prefix bg-[#f5f7fa] ml-[12px]">{{$t('plugin.tools.appName')}}</span>
                <bcs-select
                  class="flex-1"
                  allow-create
                  clearable
                  v-model="formData.rule.config.container.workload_name"
                  ref="workloadNameRef">
                  <bcs-option
                    v-for="item in workloadList"
                    :key="item.metadata.name"
                    :id="item.metadata.name"
                    :name="item.metadata.name" />
                </bcs-select>
              </div>
            </div>
          </LogPanel>
          <LogPanel
            :disabled="fromOldRule"
            :title="$t('logCollector.button.addRange.container')"
            v-if="configRangeMap['container']"
            @delete="handleDeleteRange('container')">
            <bcs-tag-input
              allow-create
              clearable
              has-delete-icon
              :value="containerName"
              @change="handleContainerNameChange">
            </bcs-tag-input>
          </LogPanel>
          <PopoverSelector
            class="mt-[16px] px-[12px]"
            offset="0, 6"
            ref="popoverSelectRef"
            :disabled="fromOldRule"
            v-if="typeList.length">
            <bcs-button
              theme="primary"
              :disabled="fromOldRule"
              outline
              icon="plus">
              {{ $t('logCollector.button.addRange.text') }}
            </bcs-button>
            <template #content>
              <li
                class="bcs-dropdown-item"
                v-for="item in typeList"
                :key="item.id"
                @click="handleAddRange(item)">
                {{ item.title }}
              </li>
            </template>
          </PopoverSelector>
          <p
            class="text-[#EA3636] px-[12px] mt-[8px] flex items-center h-[20px]"
            v-if="hasNoConfig">
            {{ $t('logCollector.title.selectRange.desc') }}
          </p>
        </div>
        <bk-form-item
          :label="$t('logCollector.label.collectorType.text')"
          property="collectorType"
          error-display-type="normal"
          class="vertical-form-item !mt-[16px]"
          required>
          <bcs-checkbox :disabled="fromOldRule" v-model="isContainerFile">
            <span class="text-[12px]">{{ $t('logCollector.label.collectorType.file') }}</span>
          </bcs-checkbox>
          <bcs-checkbox :disabled="fromOldRule" class="ml-[24px]" v-model="formData.rule.config.enable_stdout">
            <span class="text-[12px]">{{ $t('logCollector.label.collectorType.stdout') }}</span>
          </bcs-checkbox>
        </bk-form-item>
        <bk-form-item
          :label="$t('logCollector.label.logPath.text')"
          :desc="$t('logCollector.label.logPath.tips')"
          property="paths"
          error-display-type="normal"
          class="vertical-form-item !mt-[16px]"
          required
          v-if="isContainerFile">
          <div class="flex items-center mb-[8px]" v-for="item, index in formData.rule.config.paths" :key="index">
            <bcs-input :disabled="fromOldRule" v-model="formData.rule.config.paths[index]"></bcs-input>
            <i
              :class="[
                'text-[16px] text-[#C4C6CC] bk-icon icon-plus-circle-shape ml-[16px] cursor-pointer',
                { '!cursor-not-allowed !text-[#EAEBF0]': fromOldRule },
              ]"
              @click="handleAddPath"></i>
            <i
              :class="[
                { '!cursor-not-allowed !text-[#EAEBF0]':
                  (formData.rule.config.paths && formData.rule.config.paths.length === 1) || fromOldRule },
                'text-[16px] text-[#C4C6CC] bk-icon icon-minus-circle-shape ml-[10px] cursor-pointer'
              ]"
              @click="handleDeletePath(index)"></i>
          </div>
        </bk-form-item>
        <bk-form-item :label="$t('logCollector.label.encoding')" class="vertical-form-item !mt-[8px]" required>
          <bcs-select
            :clearable="false"
            :disabled="fromOldRule"
            searchable
            class="w-[270px] bg-[#fff]"
            v-model="formData.rule.config.data_encoding">
            <bcs-option v-for="item in ENCODE_LIST" :key="item.id" :id="item.id" :name="item.name" />
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('logCollector.label.matchContent.text')"
          class="vertical-form-item !mt-[16px]"
          :desc="$t('logCollector.label.matchContent.tips')">
          <bk-radio-group v-model="formData.rule.config.conditions.type">
            <bk-radio
              :disabled="fromOldRule"
              value=""
              class="text-[12px]">{{ $t('logCollector.label.matchContent.none') }}</bk-radio>
            <bk-radio
              :disabled="fromOldRule"
              value="match"
              class="text-[12px]">{{ $t('logCollector.label.matchContent.match.text') }}</bk-radio>
            <bk-radio
              :disabled="fromOldRule"
              value="separator"
              class="text-[12px]">{{ $t('logCollector.label.matchContent.separator.text') }}</bk-radio>
          </bk-radio-group>
          <div class="flex mt-[10px]" v-if="formData.rule.config.conditions.type === 'match'">
            <bcs-select
              class="w-[190px] bg-[#fff]"
              :clearable="false"
              :disabled="fromOldRule"
              v-model="formData.rule.config.conditions.match_type">
              <bcs-option id="include" :name="$t('logCollector.label.matchContent.match.include')"></bcs-option>
              <bcs-option
                v-bk-tooltips="{
                  placement: 'left',
                  content: $t('generic.msg.info.development1')
                }"
                id="exclude"
                :name="$t('logCollector.label.matchContent.match.exclude')"
                disabled>
              </bcs-option>
            </bcs-select>
            <bk-form-item
              class="flex-1"
              :label-width="0.1"
              property="type">
              <bcs-input
                :disabled="fromOldRule"
                class="ml-[8px] flex-1"
                v-model="formData.rule.config.conditions.match_content">
              </bcs-input>
            </bk-form-item>
          </div>
          <div v-else-if="formData.rule.config.conditions.type === 'separator'">
            <bcs-select
              class="w-[268px] bg-[#fff] mt-[10px]"
              :clearable="false"
              :disabled="fromOldRule"
              v-model="formData.rule.config.conditions.separator">
              <bcs-option
                v-for="item, index in separatorList"
                :key="index"
                :id="item.id"
                :name="item.name">
              </bcs-option>
            </bcs-select>
            <div class="flex items-center py-[4px] h-[28px] text-[#aeb0b7]">
              {{ $t('logCollector.label.matchContent.separator.desc') }}
            </div>
            <bk-form-item :label-width="0.1" property="type" error-display-type="normal">
              <SeparatorConfig
                :from-old-rule="fromOldRule"
                v-model="formData.rule.config.conditions.separator_filters" />
            </bk-form-item>
          </div>
        </bk-form-item>
        <bk-form-item
          :label="$t('logCollector.label.beginOfLineRegex')"
          class="vertical-form-item !mt-[16px]"
          property="multiline_pattern"
          required
          v-if="logType === 'multiline'">
          <bcs-input clearable v-model="formData.rule.config.multiline.multiline_pattern"></bcs-input>
          <i18n path="logCollector.label.multilineMaxLineAndTimeout" tag="div" class="flex items-center mt-[10px]">
            <span class="px-[4px]">
              <bk-input
                type="number"
                :min="1"
                :max="1000"
                class="w-[88px]"
                v-model="formData.rule.config.multiline.multiline_max_lines">
              </bk-input>
            </span>
            <span class="px-[4px]">
              <bk-input
                type="number"
                :min="1"
                :max="10"
                class="w-[88px]"
                v-model="formData.rule.config.multiline.multiline_timeout">
              </bk-input>
            </span>
          </i18n>
        </bk-form-item>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('logCollector.label.extraLabels')"
      property="extra_labels"
      error-display-type="normal">
      <KeyValue
        :value="formData.rule.extra_labels"
        :disabled="fromOldRule"
        @change="handleExtraLabelChange" />
      <bcs-checkbox :disabled="fromOldRule" v-model="formData.rule.add_pod_label">
        <span class="text-[12px]">{{ $t('logCollector.label.addPodLabel') }}</span>
      </bcs-checkbox>
    </bk-form-item>
    <bk-form-item :label="$t('generic.label.memo')">
      <bcs-input
        :disabled="fromOldRule"
        type="textarea"
        :maxlength="100"
        :rows="3"
        v-model="formData.description">
      </bcs-input>
    </bk-form-item>
  </bk-form>
</template>
<script setup lang="ts">
import { cloneDeep, isEqual, merge } from 'lodash';
import { computed, onBeforeMount, onMounted, PropType, ref, watch } from 'vue';

import ContainerLabel from './container-label.vue';
import KeyValue from './key-value.vue';
import LogPanel from './log-panel.vue';
import SeparatorConfig from './separator-config.vue';
import useLog, { IRuleData } from './use-log';

import { ENCODE_LIST } from '@/common/constant';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import { useCluster } from '@/composables/use-app';
import useFormLabel from '@/composables/use-form-label';
import $i18n from '@/i18n/i18n-setup';
import { useSelectItemsNamespace } from '@/views/resource-view/namespace/use-namespace';

const props = defineProps({
  clusterId: {
    type: String,
    required: true,
  },
  data: {
    type: Object as PropType<IRuleData>,
    default: () => null,
  },
  fromOldRule: {
    type: Boolean,
    default: false,
  },
});

const { clusterList } = useCluster();
const curCluster = computed(() => clusterList.value?.find(item => item.clusterID === props.clusterId) || {});

watch(() => props.data, () => {
  handleSetFormData();
}, { deep: true });

const { getWorkloadList } = useLog();
// 表单默认数据
const defaultData = {
  display_name: '',
  name: '',
  description: '',
  rule: {
    add_pod_label: false,
    extra_labels: [{
      key: '',
      value: '',
    }],
    config: {
      namespaces: [],
      paths: [''],
      data_encoding: 'UTF-8',
      enable_stdout: false,
      label_selector: {
        match_labels: [],
      },
      conditions: {
        type: '',
        match_type: 'include',
        match_content: '',
        separator: '|',
        separator_filters: [],
      },
      container: {
        workload_type: '',
        workload_name: '',
        container_name: '',
      },
      multiline: {
        multiline_pattern: '',
        multiline_max_lines: 50,
        multiline_timeout: 2,
      },
    },
  },
  from_rule: '',
};
const formData = ref<IRuleData>(cloneDeep(defaultData));
const formDataRules = ref({
  display_name: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  name: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('logCollector.validate.name'),
      trigger: 'blur',
      validator: v => v.length >= 5 && v.length <= 30 && /^[A-Za-z0-9_]+$/.test(v),
    },
  ],
  namespaces: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => {
        if (curCluster.value.is_shared) {
          // 共享集群不能选择全部命名空间
          return formData.value.rule.config.namespaces.filter(item => !!item).length;
        }
        return !!formData.value.rule.config.namespaces.length;
      },
    },
  ],
  collectorType: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => isContainerFile.value || formData.value.rule.config.enable_stdout,
    },
  ],
  paths: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => !!formData.value.rule.config.paths.filter(item => !!item).length,
    },
    {
      message: $i18n.t('logCollector.validate.logPath'),
      trigger: 'blur',
      validator: () => {
        const regex = new RegExp('^\\/.*$');
        return formData.value.rule.config.paths.every(path => regex.test(path));
      },
    },
  ],
  // 过滤内容
  type: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
      validator: () => {
        if (props.fromOldRule) return true;
        if (formData.value.rule.config.conditions.type === 'match') {
          return !!formData.value.rule.config.conditions.match_content;
        }
        if (formData.value.rule.config.conditions.type === 'separator') {
          return !!formData.value.rule.config.conditions.separator_filters.length
          && formData.value.rule.config.conditions.separator_filters.every(item => item.fieldindex && item.word);
        }
        return true;
      },
    },
  ],
  match_labels: [
    {
      message: $i18n.t('generic.validate.labelKey'),
      trigger: 'custom',
      validator: () => {
        const regex = /^[A-Za-z0-9._/-]+$/;
        const valueReg = /^[A-Za-z0-9,._/-]+$/;
        return formData.value.rule.config.label_selector.match_labels
          .every(item => (!item.key || regex.test(item.key))
          && (!item.value || valueReg.test(item.value)));
      },
    },
  ],
  multiline_pattern: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'custom',
      validator: () => {
        if (logType.value === 'multiline') {
          return !!formData.value.rule.config.multiline.multiline_pattern;
        }
        return true;
      },
    },
  ],
  // 额外标签
  extra_labels: [
    {
      message: $i18n.t('generic.validate.labelKey'),
      trigger: 'custom',
      validator: () => {
        const regex = /^[A-Za-z0-9._/-]+$/;
        return formData.value.rule.extra_labels
          .filter(item => !!item.key)
          .every(item => regex.test(item.key) && regex.test(item.value));
      },
    },
  ],
});
const isEdit = computed(() => !!formData.value.id && !props.fromOldRule);
const containerName = computed(() => formData.value.rule.config.container.container_name.split(',').filter(item => !!item));
const nsList = computed(() => formData.value.rule.config.namespaces?.filter(item => !!item));

// 处理数据
const handleSetFormData = () => {
  formData.value = merge({}, defaultData, props.data || {});

  // 采集对象是否为容器文件
  isContainerFile.value = !!formData.value.rule.config.paths?.length; // 要在空日志路径前面

  // 如果过滤内容和分割符不存在就默认显示“不过滤”选项（纯前端逻辑）
  if (!formData.value.rule.config?.conditions?.match_content
    && !formData.value.rule.config?.conditions?.separator_filters?.length) {
    formData.value.rule.config.conditions.type = '';
  }
  // 命名空间全部逻辑
  if (!formData.value.rule.config.namespaces.length && (isEdit.value || props.fromOldRule)) {
    formData.value.rule.config.namespaces = [''];
  }

  // 展示一组空的附加标签
  if (!formData.value.rule.extra_labels?.length) {
    formData.value.rule.extra_labels.push({
      key: '',
      value: '',
    });
  }

  // 展示一组空的日志路径
  if (!formData.value.rule.config.paths?.length) {
    formData.value.rule.config.paths = [''];
  }

  // 选中的方式
  configRangeList.value = [
    {
      id: 'label',
      title: $i18n.t('logCollector.button.addRange.label.text'),
      hasAdd: !!formData.value.rule?.config?.label_selector?.match_labels?.length,
    },
    {
      id: 'workload',
      title: $i18n.t('logCollector.button.addRange.workload'),
      hasAdd: !!formData.value.rule?.config?.container?.workload_type,
    },
    {
      id: 'container',
      title: $i18n.t('logCollector.button.addRange.container'),
      hasAdd: !!containerName.value.length,
    },
  ];

  // 日志类型
  logType.value = !!formData.value?.rule?.config?.multiline?.multiline_pattern ? 'multiline' : '';
};

// 获取数据
const formRef = ref();
const hasNoConfig = ref(false);
const validate = async () => {
  hasNoConfig.value = !formData.value.rule?.config?.label_selector?.match_labels?.length
  && !formData.value.rule?.config?.container?.workload_type
  && !containerName.value.length && !props.fromOldRule
  && !nsList.value.length; // 选择全部命名空间时, 三种范围必须选择一个
  const result = await formRef.value.validate().catch(() => false);

  return result && !hasNoConfig.value;
};
const handleGetFormData = async () => {
  const isValidate = await validate();
  if (!isValidate) return null;

  const data: IRuleData = cloneDeep(formData.value);
  // 处理一些空数据逻辑
  if (data.rule.config.conditions.type === 'separator') {
    data.rule.config.conditions.match_content = '';
  } else if (data.rule.config.conditions.type === 'match') {
    data.rule.config.conditions.separator_filters = [];
  } else {
    data.rule.config.conditions.match_content = '';
    data.rule.config.conditions.separator_filters = [];
  }
  data.rule.config.conditions.type = data.rule.config.conditions.type || 'match';
  data.rule.config.namespaces = data.rule.config.namespaces.filter(item => !!item);
  data.rule.extra_labels = data.rule.extra_labels.filter(item => !!item.key);
  data.rule.config.paths = data.rule.config.paths.filter(item => !!item);
  if (!isContainerFile.value) {
    data.rule.config.paths = [];
  }
  return data;
};

// 日志类型
const logType = ref<'' | 'multiline'>('');
const handleChangeLogType = (type: '' | 'multiline') => {
  logType.value = type;
  formData.value.rule.config.multiline.multiline_pattern = '';
};

// 命名空间
const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();
const handleNsChange = (nsList) => {
  hasNoConfig.value = false;
  const last = nsList[nsList.length - 1];
  // 移除全选
  if (last) {
    formData.value.rule.config.namespaces = formData.value.rule.config.namespaces.filter(item => !!item);
  } else {
    formData.value.rule.config.namespaces = [''];
  }
};

// 添加范围
const configRangeList = ref([
  {
    id: 'label',
    title: $i18n.t('logCollector.button.addRange.label.text'),
    hasAdd: !!formData.value.rule?.config?.label_selector?.match_labels?.length,
  },
  {
    id: 'workload',
    title: $i18n.t('logCollector.button.addRange.workload'),
    hasAdd: !!formData.value.rule?.config?.container?.workload_type,
  },
  {
    id: 'container',
    title: $i18n.t('logCollector.button.addRange.container'),
    hasAdd: !!containerName.value.length,
  },
]);
const configRangeMap = computed(() => configRangeList.value.reduce((pre, item) => {
  pre[item.id] = item.hasAdd;
  return pre;
}, {}));
const typeList = computed(() => configRangeList.value.filter(item => !item.hasAdd));
// 选择添加范围
const popoverSelectRef = ref();
const workloadNameRef = ref();
const handleAddRange = (item) => {
  const data = configRangeList.value.find(data => data.id === item.id);
  data && (data.hasAdd = true);
  hasNoConfig.value = false; // 重置校验
  popoverSelectRef.value?.hide();
  setTimeout(() => {
    // hack 设置工作负载组件的placeholder
    workloadNameRef.value?.$refs?.createInput?.setAttribute('placeholder', $i18n.t('logCollector.placeholder.workloadName'));
  }, 0);
};
// 删除添加范围
const handleDeleteRange = (id) => {
  const data = configRangeList.value.find(data => data.id === id);
  data && (data.hasAdd = false);

  if (id === 'label') {
    formData.value.rule.config.label_selector.match_labels = [];
  } else if (id === 'workload') {
    formData.value.rule.config.container.workload_type = '';
    formData.value.rule.config.container.workload_name = '';
  } else if (id === 'container') {
    formData.value.rule.config.container.container_name = '';
  }
};

// 手动创建标签
const handleMatchLabelsChange = (data = []) => {
  formData.value.rule.config.label_selector.match_labels = data;
};
// 工作负载选择
const kindList = ['Deployment', 'StatefulSet', 'DaemonSet', 'Job'];
const kindMap = {
  Deployment: 'deployments',
  StatefulSet: 'statefulsets',
  DaemonSet: 'daemonsets',
  Job: 'jobs',
};
const workloadLoading = ref(false);
const workloadList = ref<any[]>([]);
watch(() => [
  formData.value.rule.config.namespaces,
  formData.value.rule.config.container.workload_type,
], async (newValue, oldValue) => {
  if (isEqual(newValue, oldValue)) return;
  // 目前接口只支持单命名空间的workload查询
  const nsList = formData.value.rule.config.namespaces.filter(item => !!item);
  if (nsList.length !== 1 || !formData.value.rule.config.container.workload_type) {
    workloadList.value = [];
    return;
  }
  workloadLoading.value = true;
  const data = await getWorkloadList({
    $clusterId: props.clusterId,
    $namespaceId: nsList[0],
    $category: kindMap[formData.value.rule.config.container.workload_type],
  });
  workloadList.value = data?.manifest?.items || [];
  workloadLoading.value = false;
}, { deep: true, immediate: true });

// 指定容器
const handleContainerNameChange = (v) => {
  formData.value.rule.config.container.container_name = v.join(',');
};

// 容器内文件
const isContainerFile = ref(true);

// 采集路径
const handleAddPath = () => {
  if (props.fromOldRule) return;
  formData.value.rule.config.paths.push('');
};
const handleDeletePath = (index: number) => {
  if (formData.value.rule.config.paths?.length === 1) return;
  formData.value.rule.config.paths.splice(index, 1);
};

// 分割符
const separatorList = ref([
  {
    id: '|',
    name: $i18n.t('logCollector.label.matchContent.separator.vertical'),
  },
  {
    id: ',',
    name: $i18n.t('logCollector.label.matchContent.separator.comma'),
  },
  {
    id: '`',
    name: $i18n.t('logCollector.label.matchContent.separator.quote'),
  },
  {
    id: ';',
    name: $i18n.t('logCollector.label.matchContent.separator.semicolon'),
  },
]);

// 附加标签
const handleExtraLabelChange = (data) => {
  formData.value.rule.extra_labels = data;
};

onBeforeMount(() => {
  handleSetFormData();
  getNamespaceData({ clusterId: props.clusterId });
});

const { labelWidth, initFormLabelWidth } = useFormLabel();
onMounted(() => {
  initFormLabelWidth(formRef.value);
});

defineExpose({
  handleGetFormData,
});
</script>

<style lang="postcss" scoped>
.log-form {
  >>> .bk-label-text {
    font-size: 12px !important;
  }
  >>> .bk-form-content {
    max-width: 646px;
  }
}
>>> .vertical-form-item {
  display: flex;
  flex-direction: column;
  .bk-label {
    width: auto !important;
    text-align: left;
  }
  .bk-form-content {
    margin-left: 0px !important;
  }
}
>>> .hide-form-label .bk-form-content {
  margin-left: 0px !important;
}
</style>
