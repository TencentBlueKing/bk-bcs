<template>
  <div>
    <!-- 集群和命名空间 -->
    <template v-if="showClusterField">
      <div class="h-[20px]">
        {{ $t('view.labels.clusterAndNs') }}
        <PopoverSelector offset="12, 6" placement="bottom-start">
          <i
            :class="[
              'bcs-icon bcs-icon-pilianggouxuan',
              'ml-[16px] text-[16px] text-[#3A84FF] cursor-pointer'
            ]"
            v-bk-tooltips="$t('view.tips.selectMultiCluster')">
          </i>
          <template #content>
            <BatchClusterSelect :clusters="clusters" @change="handleBatchAddCluster" />
          </template>
        </PopoverSelector>
      </div>
      <div
        v-for="(item, index) in displayClusterNamespaces"
        :key="item.clusterID"
        class="flex items-center mt-[8px]">
        <bcs-select
          class="flex-1 w-0 mr-[8px]"
          :popover-min-width="360"
          searchable
          :clearable="false"
          :scroll-height="340"
          v-model="item.clusterID"
          @change="(clusterID) => handleClusterChange(index, clusterID)">
          <bcs-option-group
            v-for="group, groupIndex in clusterListByType"
            :key="group.type"
            :name="group.title"
            :is-collapse="collapseList.includes(group.type)"
            :class="[
              'bcs-select-group mt-[8px]',
              groupIndex === (clusterListByType.length - 1) ? 'mb-[4px]' : ''
            ]">
            <template #group-name>
              <CollapseTitle
                :title="`${group.title} (${group.list.length})`"
                :collapse="collapseList.includes(group.type)"
                @click="handleToggleCollapse(group.type)" />
            </template>
            <bcs-option
              v-for="cluster in group.list"
              :key="cluster.clusterID"
              :id="cluster.clusterID"
              :name="`${cluster.clusterName}(${cluster.clusterID})`"
              :disabled="isClusterDisabled(item, cluster)"
              class="!mt-[0px]">
              <div
                class="flex items-center justify-between px-[12px]"
                @mouseenter="hoverClusterID = cluster.clusterID"
                @mouseleave="hoverClusterID = ''">
                <div class="flex-1 flex flex-col justify-center h-[50px]">
                  <span class="leading-[20px] bcs-ellipsis" v-bk-overflow-tips="{ interactive: false }">
                    {{ cluster.clusterName }}
                  </span>
                  <span
                    :class="[
                      'leading-[20px]',
                      {
                        'text-[#979BA5]': !isClusterDisabled(item, cluster),
                        '!text-[#699DF4]': !isClusterDisabled(item, cluster)
                          && (hoverClusterID === cluster.clusterID || item.clusterID === cluster.clusterID),
                      }
                    ]">
                    ({{ cluster.clusterID }})
                  </span>
                </div>
                <bcs-tag
                  theme="danger"
                  v-if="!normalStatusList.includes(cluster.status || '')">
                  {{ $t('generic.label.abnormal') }}
                </bcs-tag>
              </div>
            </bcs-option>
          </bcs-option-group>
        </bcs-select>
        <NSSelect
          :value="{ list: item.namespaces, group: item.nsgroup }"
          :cluster-id="item.clusterID"
          class="flex-1 w-0"
          @change="(val) => handleNsChange(val, item)" />
        <i
          :class="[
            'bk-icon icon-plus-circle-shape cursor-pointer',
            'text-[16px] text-[#C4C6CC] ml-[16px]',
            disabledAddBtn
              ? '!cursor-not-allowed !text-[#EAEBF0]'
              : ''
          ]"
          @click="handleAddClusterAndNs">
        </i>
        <i
          :class="[
            'bk-icon icon-minus-circle-shape',
            'text-[16px] text-[#C4C6CC] ml-[10px] cursor-pointer',
            viewData.clusterNamespaces.length === 1 ? '!cursor-not-allowed !text-[#EAEBF0]' : ''
          ]"
          @click="handleMinusClusterAndNs(index)">
        </i>
      </div>
      <p class="mt-[2px] text-[#ea3636]" v-if="unknownClusterID">{{ $t('generic.validate.required') }}</p>
    </template>
    <!-- 视图中的条件 -->
    <Field
      v-for="item in fieldList.filter(item => item.status === 'added')"
      :title="item.title"
      :key="item.id"
      :top="true"
      :movable="false"
      class="mt-[24px]"
      @delete="handleDeleteField(item)">
      <template v-if="item.id === 'creator'">
        <BkUserSelector
          v-model="viewData.filter[item.id]"
          class="w-full"
          :api="userSelectorAPI"
          :placeholder="$t('view.placeholder.creator')"
          ref="inputRef" />
      </template>
      <template v-else-if="item.id === 'labelSelector'">
        <LabelSelector
          v-model="viewData.filter.labelSelector"
          :key="viewData.id"
          :cluster-namespaces="viewData.clusterNamespaces" />
      </template>
      <template v-else-if="item.id === 'createSource.source'">
        <bk-radio-group
          class="mb-[8px]"
          v-if="viewData.filter.createSource"
          v-model="viewData.filter.createSource.source"
          @change="handleSourceTypeChange">
          <bk-radio-button value="Template">
            <i class="bcs-icon bcs-icon-templete"></i>
            <span class="text-[12px]">Template</span>
          </bk-radio-button>
          <bk-radio-button value="Helm">
            <i class="bcs-icon bcs-icon-helm"></i>
            <span class="text-[12px]">Helm</span>
          </bk-radio-button>
          <bk-radio-button class="!ml-[-2px]" value="Client">
            <i class="bcs-icon bcs-icon-client"></i>
            <span class="text-[12px]">Client</span>
          </bk-radio-button>
          <bk-radio-button value="Web">
            <i class="bcs-icon bcs-icon-web"></i>
            <span class="text-[12px]">Web</span>
          </bk-radio-button>
        </bk-radio-group>
        <bcs-input
          v-if="viewData.filter.createSource?.source
            && ['Template', 'Helm'].includes(viewData.filter.createSource.source)"
          :value="sourceValue"
          clearable
          :placeholder="
            viewData.filter.createSource.source === 'Template'
              ? $t('view.placeholder.searchTemplate')
              : $t('view.placeholder.searchHelm')"
          @change="handleSourceChange">
        </bcs-input>
      </template>
      <template v-else>
        <bcs-input v-model.trim="viewData.filter[item.id]" clearable :placeholder="item.placeholder"></bcs-input>
      </template>
    </Field>
    <!-- 添加条件 -->
    <PopoverSelector offset="0, 8" ref="addFieldPopoverRef">
      <span
        :class="[
          'flex items-center cursor-pointer text-[14px] mt-[24px]',
          !!fieldList.filter(item => !item.status)?.length? 'text-[#3A84FF]' : 'text-[#dcdee5] !cursor-not-allowed'
        ]">
        <i class="bk-icon icon-plus-circle-shape text-[16px] mr-[4px]"></i>
        {{ addFieldText }}
      </span>
      <template #content>
        <li
          class="bcs-dropdown-item"
          v-for="item in filterFieldList"
          :key="item.id"
          @click="handleAddField(item)">
          {{ item.title }}
        </li>
      </template>
    </PopoverSelector>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep, get, isEqual, merge, set } from 'lodash';
import { computed, PropType, ref, watch } from 'vue';

import BkUserSelector from '@blueking/user-selector';

import BatchClusterSelect from './batch-cluster-select.vue';
import LabelSelector from './label-selector.vue';
import NSSelect from './ns-select-tree.vue';
import Field from './view-field.vue';

import CollapseTitle from '@/components/cluster-selector/collapse-title.vue';
import PopoverSelector from '@/components/popover-selector.vue';
import useClusterGroup from '@/composables/use-cluster-group';
import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  data: {
    type: Object as PropType<IViewData>,
    default: () => ({
      name: '',
      filter: {},
      clusterNamespaces: [],
    }),
  },
  showClusterField: {
    type: Boolean,
    default: true,
  },
  addFieldText: {
    type: String,
    default: '',
  },
});
const emits = defineEmits(['change', 'field-status-change']);

const normalStatusList = ['RUNNING'];

// 用户管理API
const userSelectorAPI = `${window.BK_USER_HOST}/api/c/compapi/v2/usermanage/fs_list_users/?app_code=bk-magicbox&page_size=100&page=1&callback=USER_LIST_CALLBACK_0`;
// popoverRef
const addFieldPopoverRef = ref();
// 视图数据
const viewData = ref<IViewData>({
  name: '',
  filter: {
    createSource: {
      source: '',
    },
  },
  clusterNamespaces: [],
});
// 来源输入框
const sourceValue = computed(() => {
  const source = viewData.value?.filter?.createSource?.source;

  // helm来源
  if (source === 'Helm') {
    return viewData.value.filter?.createSource?.chart?.chartName;
  }
  // template来源，拼接版本数据
  if (source === 'Template') {
    const template = viewData.value.filter?.createSource?.template;
    return `${template?.templateName}${template?.templateVersion ? `:${template?.templateVersion}` : ''}`;
  }
  // 其他来源
  return '';
});
const displayClusterNamespaces = computed<IClusterNamespace[]>(() => {
  if (!viewData.value.clusterNamespaces?.length) {
    return [{
      clusterID: '',
      namespaces: [],
      nsgroup: 'all-user',
    }];
  }
  return viewData.value.clusterNamespaces;
});
// 校验集群ID正确性
const unknownClusterID = computed(() => viewData.value?.clusterNamespaces?.some((item) => {
  const { clusterID } = item;
  return !clusterNameMap.value[clusterID];
}));
// 筛选字段
const fieldList = ref<Array<IFieldItem>>([
  {
    title: $i18n.t('view.labels.creator'),
    id: 'creator',
    status: '',
  },
  {
    title: $i18n.t('k8s.label'),
    id: 'labelSelector',
    status: '',
  },
  {
    title: $i18n.t('view.labels.resourceName'),
    id: 'name',
    status: '',
    placeholder: $i18n.t('view.placeholder.resourceName'),
  },
  {
    title: $i18n.t('generic.label.source'),
    id: 'createSource.source',
    status: '',
  },
]);
const filterFieldList = computed(() => fieldList.value.filter(item => !item.status).sort((a, b) => {
  if (a.id < b.id) {
    return -1;
  }

  if (a.id > b.id) {
    return 1;
  }
  return 0;
}));
watch(fieldList, () => {
  emits('field-status-change', fieldList.value);
}, { deep: true });

// 重置临时条件
const resetFieldListStatus = () => {
  fieldList.value.forEach((item) => {
    if (get(viewData.value.filter, item.id)?.length) {
      item.status = 'added';
    } else {
      item.status = '';
    }
  });
};

// 集群分组
const hoverClusterID = ref<string>();
const { clusterList, clusterNameMap, clusterListByType, collapseList, handleToggleCollapse } = useClusterGroup();
const clusters = computed(() => viewData.value.clusterNamespaces?.map(item => item.clusterID));
const clusterNsMap = computed(() => viewData.value.clusterNamespaces?.reduce((pre, item) => {
  pre[item.clusterID] = {
    namespaces: item.namespaces,
    nsgroup: item.nsgroup,
  };
  return pre;
}, {}));

// 批量添加集群
const handleBatchAddCluster = (clusters: string[]) => {
  viewData.value.clusterNamespaces = clusters.map(clusterID => ({
    clusterID,
    namespaces: clusterNsMap.value[clusterID]?.namespaces || [],
    nsgroup: clusterNsMap.value[clusterID]?.nsgroup || 'all-user',
  }));
};

// 当前集群是否能选择（一个集群只能选一次）
const isClusterDisabled = (item, cluster) => (viewData.value.clusterNamespaces
  .some(data => data.clusterID === cluster.clusterID) && item.clusterID !== cluster.clusterID)
  || !normalStatusList.includes(cluster.status || '');
// 添加集群和命名空间
const disabledAddBtn = computed(() => viewData.value.clusterNamespaces?.length === clusterList.value.length
  || !clusterList.value.find(item => !clusterNsMap.value[item.clusterID])?.clusterID);
const handleAddClusterAndNs = () => {
  if (viewData.value.clusterNamespaces?.length === clusterList.value.length) return;

  const nextClusterID = clusterList.value.find(item => !clusterNsMap.value[item.clusterID])?.clusterID;
  if (!nextClusterID) return;

  viewData.value.clusterNamespaces.push({
    clusterID: nextClusterID,
    namespaces: [],
    nsgroup: 'all-user',
  });
};
// 删除集群和命名空间
const handleMinusClusterAndNs = (index: number) => {
  if (viewData.value.clusterNamespaces?.length <= 1) return;

  viewData.value.clusterNamespaces?.splice(index, 1);
};

// 集群切换
const handleClusterChange = (index: number, clusterID: string) => {
  if (!viewData.value.clusterNamespaces[index]) {
    viewData.value.clusterNamespaces.push({
      clusterID,
      namespaces: [],
      nsgroup: '',
    });
  } else {
    viewData.value.clusterNamespaces[index].namespaces = [];
  }
};

// 添加查询字段
const handleAddField = (item: IFieldItem) => {
  const data = fieldList.value.find(data => data.id === item.id);
  if (data) {
    data.status = 'added';
    // 来源字段设置默认值
    if (data?.id === 'createSource.source' && !viewData.value.filter?.createSource?.source) {
      set(viewData.value.filter, data?.id, 'Template');
    }
    // 排序（新添加的总是排在最后）
    fieldList.value.sort((pre, current) => {
      if (current.id === item.id) {
        return -1;
      }
      return 0;
    });
  }
  addFieldPopoverRef.value?.hide();
};
// 删除查询字段
const handleDeleteField = (item: IFieldItem) => {
  const data = fieldList.value.find(data => data.id === item.id);
  if (data) {
    data.status = '';
    const emptyValue = Array.isArray(get(viewData.value.filter, item.id)) ? [] : '';
    set(viewData.value.filter, item.id, emptyValue);
  }
  if (item.id === 'createSource.source') {
    handleSourceTypeChange();
  };
};

// 来源类型变更时重置数据
function handleSourceTypeChange() {
  if (viewData.value.filter.createSource?.template) {
    viewData.value.filter.createSource.template = {
      templateName: '',
      templateVersion: '',
    };
  }

  if (viewData.value.filter.createSource?.chart) {
    viewData.value.filter.createSource.chart = { chartName: '' };
  }
}

// 来源变更
function handleSourceChange(v: string) {
  if (!viewData.value.filter.createSource?.source) return;

  if (viewData.value.filter.createSource.source === 'Template') {
    // reverse为了方便取出最后分隔作为version
    const [last, ...reset] = v?.split(':')?.reverse() || [];
    if (reset?.length) {
      viewData.value.filter.createSource.template = {
        templateName: reset?.join(':'),
        templateVersion: last,
      };
    } else {
      viewData.value.filter.createSource.template = {
        templateName: last,
        templateVersion: '',
      };
    }
  } else if (viewData.value.filter.createSource.source === 'Helm') {
    viewData.value.filter.createSource.chart = { chartName: v || '' };
  }
}

function handleNsChange({ value, group }: { value: string[], group: IGroup }, item: IClusterNamespace) {
  item.namespaces = value;
  item.nsgroup = group;
};

watch(() => props.data, () => {
  if (isEqual(props.data, viewData.value)) return;

  viewData.value = merge({
    name: '',
    filter: {
      createSource: {
        source: '',
        template: {
          templateName: '',
          templateVersion: '',
        },
        chart: {
          chartName: '',
        },
      },
    },
    clusterNamespaces: [],
  }, cloneDeep(props.data));
  resetFieldListStatus();
}, { immediate: true, deep: true });

watch(viewData, () => {
  emits('change', viewData.value);
}, { deep: true });
</script>
<style lang="postcss" scoped>
/deep/ .bk-form-radio-button {
  .bk-radio-button-text {
    width: 88px;
    height: 26px;
    line-height: 26px;
    padding: 0;
  }
}
</style>
