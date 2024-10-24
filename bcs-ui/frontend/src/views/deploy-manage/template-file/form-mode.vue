<template>
  <bcs-resize-layout
    :collapsible="true"
    v-bkloading="{ isLoading }"
    disabled
    :border="false"
    :initial-divide="formToJson && formToJson.length ? '230px' : '0px'"
    @collapse-change="handleCollapseChange"
    class="flex h-full">
    <div slot="aside" class="bg-[#fff] h-full overflow-y-auto overflow-x-hidden">
      <div
        :class="[
          'h-full py-[0 8px] border-r-[1px] border-solid bg-[#fff] w-[230px] pt-[12px]'
        ]">
        <div
          v-for="item, index in formToJson"
          :key="index"
          :class="[
            'flex items-center cursor-pointer leading-[20px]',
            'h-[32px] text-[12px] px-[12px]',
            activeFormIndex === index ? 'bg-[#e1ecff] text-[#3a84ff] border-r-2 border-[#3a84ff]' : ''
          ]"
          @click="handleAnchor(index)">
          <span
            :class="[
              'rounded-full w-2.5 h-2.5 bg-[red] border-2 border-white flex-shrink-0',
              validArray[index] ? 'visible' : 'invisible'
            ]"></span>
          <span
            :class="[
              'rounded-full w-4 h-4 leading-[1rem] text-center text-[#fff] mx-2 flex-shrink-0',
              activeFormIndex === index ? 'bg-[#3a84ff] text-[#fff]' : 'bg-[#979ba5]'
            ]">{{ index + 1 }}</span>
          <span class="bcs-ellipsis" v-bk-overflow-tips>{{ item || $t('templateFile.label.untitled') }}</span>
        </div>
      </div>
    </div>
    <div slot="main" class="h-full">
      <div
        :class="[
          'h-full flex flex-col gap-[16px] text-[12px] min-h-[200px] overflow-auto form-content transition-all',
          isCollapse ? '' : 'ml-[16px]'
        ]"
        @scroll="throttledOnScroll">
        <!-- 表单 -->
        <div
          class="bcs-border border-t-0 bg-[#fff] form-section"
          v-for="form, index in schemaFormData"
          :key="index"
          :id="`template-file-form-${index}`"
          ref="sectionRefs">
          <div
            :class="[
              'bcs-border-bottom bcs-border-top overflow-x-hidden',
              'sticky top-0 z-[1]',
              'flex items-center justify-between h-[56px] bg-[#FAFBFD] px-[24px]'
            ]">
            <div class="flex items-center gap-[24px]">
              <span class="inline-flex items-center min-w-[240px]">
                <span :class="['bcs-prefix mr-[-1px]', { 'bcs-required': !isEdit }]">
                  {{ $t('templateFile.label.kindType') }}
                </span>
                <bcs-select
                  class="flex-1 bg-[#fff]" searchable :clearable="false" :readonly="isEdit"
                  v-model="form.kind" @selected="handleKindChange(index)">
                  <bcs-option-group
                    v-for="item, i in kindList" :key="item.id"
                    :is-collapse="collapseList.includes(item.id)" :name="item.id" :class="[
                      'mt-[8px]',
                      i === (kindList.length - 1) ? 'mb-[4px]' : ''
                    ]">
                    <template #group-name>
                      <CollapseTitle
                        :title="item.name" :collapse="collapseList.includes(item.id)"
                        @click="handleToggleCollapse(item.id)" />
                    </template>
                    <bcs-option v-for="kind in item.children" :key="kind" :id="kind" :name="kind">
                    </bcs-option>
                  </bcs-option-group>
                </bcs-select>
              </span>
              <span class="inline-flex items-center min-w-[240px]">
                <span :class="['bcs-prefix mr-[-1px]', { 'bcs-required': !isEdit }]">apiVersion</span>
                <bcs-select
                  class="flex-1 bg-[#fff]" searchable :clearable="false" :readonly="isEdit"
                  v-model="form.apiVersion">
                  <bcs-option
                    v-for="item in kindApiVersionMap[form.kind]" :key="item.value" :id="item.value"
                    :name="item.label" />
                </bcs-select>
              </span>
              <!-- 资源名称 -->
              <span class="inline-flex items-center min-w-[240px]">
                <span :class="['bcs-prefix mr-[-1px]', { 'bcs-required': !isEdit }]">
                  {{ $t('generic.label.name')}}
                </span>
                <Validate
                  :value="form.formData.metadata.name"
                  required
                  v-if="form.formData.metadata"
                  :rules="[
                    {
                      validator: () => handleValidatorName(index),
                      message: $t('generic.validate.fieldRepeat', [$t('generic.label.resourceName')])
                    }
                  ]"
                  trigger="blur"
                  ref="validateRefs">
                  <bcs-input
                    class="flex-1"
                    :readonly="isEdit"
                    v-model="form.formData.metadata.name"
                    @blur="handleBlur(index, form.formData.metadata.name)" />
                </Validate>
              </span>
            </div>
            <span class="text-[#979BA5] text-[14px] gap-[10px]" v-if="!isEdit">
              <i class="bk-icon icon-plus-circle-shape cursor-pointer" @click="addKind"></i>
              <i
                :class="[
                  'bk-icon icon-minus-circle-shape cursor-pointer',
                  schemaFormData.length <= 1 ? '!cursor-not-allowed !text-[#c4c6cc]' : ''
                ]" @click="removeKind(index)"></i>
            </span>
          </div>
          <BKSchemaForm
            v-model="form.formData"
            ref="bkuiFormRef"
            :schema="kindVersionSchemaMap[`${form.kind}-${form.apiVersion}`]?.schema"
            :layout="kindVersionSchemaMap[`${form.kind}-${form.apiVersion}`]?.layout"
            :rules="kindVersionSchemaMap[`${form.kind}-${form.apiVersion}`]?.rules" :context="context"
            :http-adapter="{
              request: requestAdapter
            }"
            readonly-mode="custom"
            form-type="vertical"
            :readonly="isEdit"
            :disabled="isEdit"
            class="px-[14px] py-[10px]" />
        </div>
        <!-- 切换锚点 -->
        <div
          :class="[
            'fixed right-[16px] bottom-[64px] z-10 bg-[#fff]',
            'flex flex-col items-center justify-center gap-[16px] w-[52px] h-[104px] rounded-[26px]',
            'shadow-[0_2px_8px_0_rgba(0,0,0,0.16)] hover:shadow-[0_2px_12px_0_rgba(0,0,0,0.2)]'
          ]" v-if="schemaFormData.length > 1">
          <bcs-icon
            :class="[
              '!text-[24px] cursor-pointer hover:text-[#3a84ff]',
              activeFormIndex === 0 ? '!text-[#DCDEE5] cursor-not-allowed' : ''
            ]" type="angle-up" v-bk-tooltips="$t('templateFile.tips.preKind')" @click="preKind" />
          <bcs-icon
            :class="[
              '!text-[24px] cursor-pointer hover:text-[#3a84ff]',
              activeFormIndex === (schemaFormData.length - 1) ? '!text-[#DCDEE5] cursor-not-allowed' : ''
            ]" type="angle-down" v-bk-tooltips="$t('templateFile.tips.nextKind')" @click="nextKind" />
        </div>
      </div>
    </div>
  </bcs-resize-layout>
</template>
<script setup lang="ts">
import { cloneDeep, isEqual, throttle } from 'lodash';
import { computed, nextTick, onBeforeMount, ref, set, watch } from 'vue';

import createForm from '@blueking/bkui-form/dist/bkui-form-umd';

import BcsVarDatasourceInput from './bcs-variable-datasource-input.vue';
import { updateVarList } from './use-store';

import '@blueking/bkui-form/dist/bkui-form.css';
import { ISchemaData } from '@/@types/cluster-resource-patch';
import { ResourceService } from '@/api/modules/new-cluster-resource';
import request from '@/api/request';
import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import Validate from '@/components/validate.vue';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';

const props = defineProps({
  isEdit: {
    type: Boolean,
    default: false,
  },
  // yaml数据
  value: {
    type: String,
    default: '',
  },
});

const activeFormIndex = ref(0);

// schema form上下文
const projectID = computed(() => $store.state.curProject?.projectID);
const context = {
  projectID: projectID.value,
  baseUrl: `${process.env.NODE_ENV === 'development' ? '' : window.BCS_API_HOST}/bcsapi/v4/clusterresources/v1`,
};
// http 适配器
async function requestAdapter(url, config) {
  const requestMethods = request(config.method || 'get', url);
  const data = await requestMethods(config.params);
  return data?.selectItems || [];
}
const BKSchemaForm = createForm({
  namespace: 'bcs',
  components: {
    input: BcsVarDatasourceInput,
  },
  baseWidgets: {
    radio: 'bk-radio',
    'radio-group': 'bk-radio-group',
  },
});
const isLoading = ref(false);
// 模板文件表单数据
const initFormData = {
  kind: 'Deployment',
  formData: {},
  apiVersion: '',
};
const schemaFormData = ref<ClusterResource.FormData[]>([]);
const validArray = computed(() => schemaFormData.value.reduce<Array<boolean>>((pre, cur) => {
  const name = cur?.formData?.metadata?.name || '';
  pre.push(!name);
  return pre;
}, []));
const formToJson = computed(() => schemaFormData.value.reduce<Array<string>>((pre, cur) => {
  const name = cur?.formData?.metadata?.name || '';
  pre.push(name);
  return pre;
}, []));

// 资源名称唯一校验
function handleValidatorName(index: number) {
  const item = schemaFormData.value[index];
  if (!item) return false;

  const name = item?.formData?.metadata?.name || '';
  const kind = item?.kind;

  return !schemaFormData.value
    .filter((_, i) => i !== index) // 排除当前index
    .some(form => form.formData?.metadata?.name === name && form.kind === kind);// 同种kind的name不能重复
}

const curResourceType = computed(() => schemaFormData.value.reduce<ClusterResource.FormResourceType[]>((pre, item) => {
  const exist = pre.find(d => d.apiVersion === item.apiVersion && d.kind === item.kind);
  if (!exist) {
    pre.push({
      apiVersion: item.apiVersion,
      kind: item.kind,
    });
  }
  return pre;
}, []));
// apiVersion
const kindApiVersionMap = ref<Record<string, { label: string; value: string }[]>>({});

// 获取apiVersion数据
const getFormSupportedAPIVersions = () => {
  curResourceType.value.filter(item => !kindApiVersionMap.value[item.kind]).forEach((item) => {
    ResourceService.GetFormSupportedAPIVersions({
      kind: item.kind,
      $clusterId: '-', // todo 修改插件
    }).then(({ selectItems }) => {
      kindApiVersionMap.value[item.kind] = selectItems || [];
      // 设置默认kind
      schemaFormData.value
        .filter(d => !d.apiVersion && d.kind === item.kind)
        .forEach(d => d.apiVersion = selectItems[0]?.value);
    });
  });
};

// 切换类型（select组件默认会触发一次change事件）
const handleKindChange = (index: number) => {
  const data = schemaFormData.value[index];
  if (!data) return;

  data.apiVersion = kindApiVersionMap.value[data.kind]?.at(0)?.value || '';
  data.formData = {};
};

// 新增表单
const addKind = async () => {
  const data = cloneDeep(initFormData);
  data.kind = schemaFormData.value[0]?.kind;
  data.apiVersion = schemaFormData.value[0]?.apiVersion;
  data.formData = {
    // 重置name（不然出现重复的name）
    metadata: {
      name: '',
    },
  };
  schemaFormData.value.push(data);
  nextTick(() => {
    handleAnchor(schemaFormData.value.length - 1);
  });
};

// 删除表单
const removeKind = async (index: number) => {
  if (schemaFormData.value.length <= 1) return;
  schemaFormData.value.splice(index, 1);
};

// kind和version对应的schema数据
const kindVersionSchemaMap = ref<Record<string, ISchemaData>>({});

// 获取资源类型对应的schema数据
const handleChangeSchema = async (data: ClusterResource.FormResourceType[] = []) => {
  // kind不存在、apiVersion字段没设置和已经请求过时就不在发生请求
  const resourceTypes = data.filter(item => item.kind && item.apiVersion && !kindVersionSchemaMap.value[`${item.kind}-${item.apiVersion}`]);
  if (!resourceTypes.length) return;

  isLoading.value = true;
  const schemaDataList = await ResourceService.GetMultiResFormSchema({
    resourceTypes,
  }).catch(() => ([]));
  isLoading.value = false;

  schemaDataList.forEach((item) => {
    set(kindVersionSchemaMap.value, `${item.kind}-${item.apiVersion}`, item);
  });
};

// 资源类型
const collapseList = ref<string[]>([]);
const kindList = ref([
  {
    id: 'workload',
    name: $i18n.t('templateFile.label.workload'),
    children: ['Deployment', 'StatefulSet', 'DaemonSet', 'Job', 'CronJob'],
  },
  {
    id: 'network',
    name: $i18n.t('k8s.networking'),
    children: ['Ingress', 'Service'],
  },
  {
    id: 'config',
    name: $i18n.t('nav.configuration'),
    children: ['ConfigMap', 'Secret'],
  },
  {
    id: 'storage',
    name: $i18n.t('generic.label.storage'),
    children: ['PersistentVolume', 'PersistentVolumeClaim', 'StorageClass'],
  },
  {
    id: 'HPA',
    name: 'HPA',
    children: ['HorizontalPodAutoscaler'],
  },
]);
const handleToggleCollapse = (id: string) => {
  const index = collapseList.value.findIndex(item => item === id);
  if (index > -1) {
    collapseList.value.splice(index, 1);
  } else {
    collapseList.value.push(id);
  }
};

// 表单数据转YAML
const getData = async () => {
  isLoading.value = true;
  const { manifest } = await ResourceService.FormToYAML({
    resources: schemaFormData.value,
  }).catch(() => ({ manifest: '' }));
  isLoading.value = false;
  return manifest; // 表单数据
};

// 校验表单
const bkuiFormRef = ref<Array<any>>();
const validateRefs = ref();
const validate = async () => {
  // 校验名称必填
  for (let i = 0; i < validateRefs.value.length; i++) {
    const res = await validateRefs.value[i].validate('blur');
    if (!res) {
      document.getElementById(`template-file-form-${i}`)?.scrollIntoView();
      handleAnchor(i);
      // 出现一个没填直接返回
      return false;
    }
  }

  // 校验表单
  const data = bkuiFormRef.value?.map(item => item.validate()) || [];
  const result = await Promise.all(data).then(() => true)
    .catch(() => false);
  if (!result) {
    // 自动滚动到第一个错误的位置
    const errDom = document?.querySelectorAll('.bk-schema-form-item__error-tips');
    errDom[0]?.scrollIntoView({
      block: 'center',
    });
  }
  return result;
};

// 滚动到上一个kind
const preKind = () => {
  if ((activeFormIndex.value - 1) < 0) return;

  activeFormIndex.value -= 1;
  document.getElementById(`template-file-form-${activeFormIndex.value}`)?.scrollIntoView();
};
// 滚动到下一个kind
const nextKind = () => {
  if ((activeFormIndex.value + 1) >= schemaFormData.value.length) return;

  activeFormIndex.value += 1;
  document.getElementById(`template-file-form-${activeFormIndex.value}`)?.scrollIntoView();
};

// 跳转到对应的form
const isProgrammaticScroll = ref(true);
const handleAnchor = (index: number) => {
  // 防止点击滚动时导致tab高亮显示不准确
  isProgrammaticScroll.value = false;
  activeFormIndex.value = index;
  document.getElementById(`template-file-form-${activeFormIndex.value}`)?.scrollIntoView();
  setTimeout(() => {
    isProgrammaticScroll.value = true;
  }, 300);
};

const handleBlur = (index: number, name: string) => {
  if (formToJson.value.length === 0) return;
  formToJson.value[index] = name ? name : '';
};

const sectionRefs = ref();
const handleScroll = () => {
  if (!isProgrammaticScroll.value) return;
  // 获取当前滚动位置
  const scrollTop = document.querySelector('.form-content')?.scrollTop as number;

  // 遍历所有 section，定位当前滚动位置
  for (let i = 0; i < sectionRefs.value.length; i++) {
    const sectionTop = sectionRefs.value[i].offsetTop;
    const sectionHeight = sectionRefs.value[i].offsetHeight;

    if (scrollTop >= sectionTop && scrollTop < sectionTop + sectionHeight) {
      activeFormIndex.value = i;
      break;
    }
  }
};
// 使用 lodash 的 throttle 创建节流函数，设定每 300 毫秒触发一次
const throttledOnScroll = throttle(handleScroll, 500);

const isCollapse = ref(false);
const handleCollapseChange = (value: boolean) => {
  isCollapse.value = value;
};

watch(() => props.value, async () => {
  if (!props.value && props.isEdit) return;// 编辑态时不初始化表单
  if (!props.value) {
    // 非编辑态时默认初始化一条数据
    schemaFormData.value = [cloneDeep(initFormData)];
    return;
  }
  isLoading.value = true;
  const data = await ResourceService.YAMLToForm({
    yaml: props.value,
  }).catch(() => ({ resources: [] }));
  isLoading.value = false;
  schemaFormData.value = data?.resources || [cloneDeep(initFormData)];
}, { immediate: true });

watch(curResourceType, (newValue, oldValue) => {
  if (isEqual(newValue, oldValue)) return;
  handleChangeSchema(curResourceType.value);
  getFormSupportedAPIVersions();
}, { immediate: true });

defineExpose({
  getData,
  validate,
});

onBeforeMount(() => {
  updateVarList();
});
</script>
<style lang="postcss" scoped>
/deep/ .bk-option-group-name {
  border-bottom: 0 !important;
}
/deep/ .bk-schema-form-group.card {
  padding-bottom: 10px;
}
</style>
