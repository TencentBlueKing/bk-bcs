<template>
  <bk-form class="log-form" :rules="formDataRules" :model="formData" ref="formRef">
    <bk-form-item :label="$t('规则名称')" property="name" error-display-type="normal" required v-if="!isEdit">
      <bcs-input v-model="formData.name"></bcs-input>
    </bk-form-item>
    <bk-form-item :label="$t('配置信息')">
      <div class="border border-[#DCDEE5] border-solid p-[16px] bg-[#FAFBFD]">
        <div class="border border-[#EAEBF0] border-solid px-[12px] py-[12px] bg-[#fff] rounded-sm">
          <Row class="mb-[6px] px-[12px]">
            <template #left>
              <span class="font-bold">{{ $t('选择容器范围') }}</span>
              <span class="ml-[12px]">
                <i class="bk-icon icon-info-circle text-[#979BA5] text-[14px]"></i>
                <span class="ml-[4px]">{{ $t('所有选择范围可相互叠加并作用，除命名空间外，至少添加一种范围') }}</span>
              </span>
            </template>
          </Row>
          <bk-form-item
            :label="$t('命名空间')"
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
              @selected="handleNsChange">
              <bcs-option key="all" id="" :name="$t('所有')" v-if="!curCluster.is_shared"></bcs-option>
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
            :title="$t('按标签选择')"
            :desc="$t('如果添加多个标签，只会采集Pod上同时存在多个标签的容器日志')"
            @delete="handleDeleteRange('label')">
            <bk-form-item
              :label-width="0.1"
              property="match_labels"
              error-display-type="normal">
              <KeyValue
                :value="formData.rule.config.label_selector.match_labels"
                :disabled="fromOldRule"
                @change="handleMatchLabelsChange" />
            </bk-form-item>
          </LogPanel>
          <LogPanel
            :disabled="fromOldRule"
            :title="$t('按工作负载选择')"
            v-if="configRangeMap['workload']"
            @delete="handleDeleteRange('workload')">
            <div class="flex items-center">
              <div class="w-[200px] flex">
                <span class="bcs-form-prefix bg-[#f5f7fa]">{{$t('应用类型')}}</span>
                <bcs-select
                  class="flex-1 bg-[#fff]"
                  :clearable="false"
                  v-model="formData.rule.config.container.workload_type">
                  <bcs-option v-for="item in kindList" :key="item" :id="item" :name="item"></bcs-option>
                </bcs-select>
              </div>
              <div class="flex-1 flex">
                <span class="bcs-form-prefix bg-[#f5f7fa] ml-[12px]">{{$t('应用名称')}}</span>
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
            :title="$t('按容器名称选择')"
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
              {{ $t('添加范围') }}
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
            {{ $t('所有选择范围可相互叠加并作用，除命名空间外，至少添加一种范围') }}
          </p>
        </div>
        <bk-form-item
          :label="$t('采集对象')"
          property="collectorType"
          error-display-type="normal"
          class="vertical-form-item !mt-[16px]"
          required>
          <bcs-checkbox :disabled="fromOldRule" v-model="isContainerFile">
            <span class="text-[12px]">{{ $t('容器内文件') }}</span>
          </bcs-checkbox>
          <bcs-checkbox :disabled="fromOldRule" class="ml-[24px]" v-model="formData.rule.config.enable_stdout">
            <span class="text-[12px]">{{ $t('标准输出') }}</span>
          </bcs-checkbox>
        </bk-form-item>
        <bk-form-item
          :label="$t('日志路径')"
          :desc="$t('只支持星号（*）、问号（?）两种通配符')"
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
                { '!cursor-not-allowed !text-[#EAEBF0]': formData.rule.config.paths.length === 1 || fromOldRule },
                'text-[16px] text-[#C4C6CC] bk-icon icon-minus-circle-shape ml-[10px] cursor-pointer'
              ]"
              @click="handleDeletePath(index)"></i>
          </div>
        </bk-form-item>
        <bk-form-item :label="$t('日志字符集')" class="vertical-form-item !mt-[8px]" required>
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
          :label="$t('过滤内容')"
          class="vertical-form-item !mt-[16px]"
          :desc="$t('为减少传输和存储成本，可以过滤掉部分内容，默认不开启过滤')">
          <bk-radio-group v-model="formData.rule.config.conditions.type">
            <bk-radio :disabled="fromOldRule" value="" class="text-[12px]">{{ $t('不过滤') }}</bk-radio>
            <bk-radio :disabled="fromOldRule" value="match" class="text-[12px]">{{ $t('字符串过滤') }}</bk-radio>
            <bk-radio :disabled="fromOldRule" value="separator" class="text-[12px]">{{ $t('分隔符过滤') }}</bk-radio>
          </bk-radio-group>
          <div class="flex mt-[10px]" v-if="formData.rule.config.conditions.type === 'match'">
            <bcs-select
              class="w-[190px] bg-[#fff]"
              :clearable="false"
              :disabled="fromOldRule"
              v-model="formData.rule.config.conditions.match_type">
              <bcs-option id="include" :name="$t('include(保留匹配字符串)')"></bcs-option>
              <bcs-option
                v-bk-tooltips="{
                  placement: 'left',
                  content: $t('功能开发中，暂不开放')
                }"
                id="exclude"
                :name="$t('exclude(过滤匹配字符串)')"
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
              {{ $t('复杂的过滤条件（超过5个）会影响机器性能') }}
            </div>
            <bk-form-item :label-width="0.1" property="type" error-display-type="normal">
              <SeparatorConfig
                :from-old-rule="fromOldRule"
                v-model="formData.rule.config.conditions.separator_filters" />
            </bk-form-item>
          </div>
        </bk-form-item>
      </div>
    </bk-form-item>
    <bk-form-item
      :label="$t('附加日志标签')"
      property="extra_labels"
      error-display-type="normal">
      <KeyValue
        :value="formData.rule.extra_labels"
        :disabled="fromOldRule"
        @change="handleExtraLabelChange" />
      <bcs-checkbox :disabled="fromOldRule" v-model="formData.rule.add_pod_label">
        <span class="text-[12px]">{{ $t('是否自动添加 Pod 中的 labels') }}</span>
      </bcs-checkbox>
    </bk-form-item>
    <bk-form-item :label="$t('备注')">
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
import { PropType, computed, onBeforeMount, ref, watch } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import Row from '@/components/layout/Row.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import LogPanel from './log-panel.vue';
import SeparatorConfig from './separator-config.vue';
import { useSelectItemsNamespace } from '@/views/resource-view/namespace/use-namespace';
import useLog, { IRuleData } from './use-log';
import { ENCODE_LIST } from '@/common/constant';
import { cloneDeep, isEqual } from 'lodash';
import { useCluster } from '@/composables/use-app';
import KeyValue from './key-value.vue';

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
    },
  },
  from_rule: '',
};
const formData = ref<IRuleData>(cloneDeep(defaultData));
const formDataRules = ref({
  name: [
    {
      required: true,
      message: $i18n.t('必填项'),
      trigger: 'blur',
    },
    {
      message: $i18n.t('仅支持英文, 数字和下划线, 且长度5~30字符'),
      trigger: 'blur',
      validator: v => v.length >= 5 && v.length <= 30 && /^[A-Za-z0-9_]+$/.test(v),
    },
  ],
  namespaces: [
    {
      message: $i18n.t('必填项'),
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
      message: $i18n.t('必填项'),
      trigger: 'blur',
      validator: () => isContainerFile.value || formData.value.rule.config.enable_stdout,
    },
  ],
  paths: [
    {
      message: $i18n.t('必填项'),
      trigger: 'blur',
      validator: () => !!formData.value.rule.config.paths.filter(item => !!item).length,
    },
    {
      message: $i18n.t('仅支持绝对路径'),
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
      message: $i18n.t('必填项'),
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
      message: $i18n.t('仅支持字母，数字和字符(-_./)'),
      trigger: 'custom',
      validator: () => {
        const regex = /^[A-Za-z0-9._/-]+$/;
        return formData.value.rule.config.label_selector.match_labels
          .filter(item => !!item.key)
          .every(item => regex.test(item.key) && regex.test(item.value));
      },
    },
  ],
  // 额外标签
  extra_labels: [
    {
      message: $i18n.t('仅支持字母，数字和字符(-_./)'),
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
  formData.value = cloneDeep(props.data || defaultData);

  // 采集对象是否为容器文件
  isContainerFile.value = !!formData.value.rule.config.paths.length; // 要在空日志路径前面

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
  if (!formData.value.rule.config.paths.length) {
    formData.value.rule.config.paths = [''];
  }

  // 选中的方式
  configRangeList.value = [
    {
      id: 'label',
      title: $i18n.t('按标签选择'),
      hasAdd: !!formData.value.rule?.config?.label_selector?.match_labels?.length,
    },
    {
      id: 'workload',
      title: $i18n.t('按工作负载选择'),
      hasAdd: !!formData.value.rule?.config?.container?.workload_type,
    },
    {
      id: 'container',
      title: $i18n.t('按容器名称选择'),
      hasAdd: !!containerName.value.length,
    },
  ];
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
    title: $i18n.t('按标签选择'),
    hasAdd: !!formData.value.rule?.config?.label_selector?.match_labels?.length,
  },
  {
    id: 'workload',
    title: $i18n.t('按工作负载选择'),
    hasAdd: !!formData.value.rule?.config?.container?.workload_type,
  },
  {
    id: 'container',
    title: $i18n.t('按容器名选择'),
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
    workloadNameRef.value?.$refs?.createInput?.setAttribute('placeholder', $i18n.t('请输入应用名称, 支持正则表达式'));
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
  if (formData.value.rule.config.paths.length === 1) return;
  formData.value.rule.config.paths.splice(index, 1);
};

// 分割符
const separatorList = ref([
  {
    id: '|',
    name: $i18n.t('竖线(|)'),
  },
  {
    id: ',',
    name: $i18n.t('逗号(,)'),
  },
  {
    id: '`',
    name: $i18n.t('反引号(`)'),
  },
  {
    id: ';',
    name: $i18n.t('分号(;)'),
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
</style>
