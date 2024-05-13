<template>
  <div>
    <!-- 集群和命名空间 -->
    <template v-if="showClusterField">
      <div class="h-[20px]">
        {{ $t('view.labels.clusterAndNs') }}
        <PopoverSelector offset="0, 6">
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
                class="flex flex-col justify-center h-[50px] px-[12px]"
                @mouseenter="hoverClusterID = cluster.clusterID"
                @mouseleave="hoverClusterID = ''">
                <span class="leading-[20px] bcs-ellipsis" v-bk-overflow-tips>{{ cluster.clusterName }}</span>
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
            </bcs-option>
          </bcs-option-group>
        </bcs-select>
        <NSSelect v-model="item.namespaces" :cluster-id="item.clusterID" class="flex-1 w-0" />
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
import { cloneDeep, isEqual, merge } from 'lodash';
import { computed, PropType, ref, watch } from 'vue';

import BkUserSelector from '@blueking/user-selector';

import BatchClusterSelect from './batch-cluster-select.vue';
import LabelSelector from './label-selector.vue';
import NSSelect from './ns-select.vue';
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

// 用户管理API
const userSelectorAPI = `${window.BK_USER_HOST}/api/c/compapi/v2/usermanage/fs_list_users/?app_code=bk-magicbox&page_size=100&page=1&callback=USER_LIST_CALLBACK_0`;
// popoverRef
const addFieldPopoverRef = ref();
// 视图数据
const viewData = ref<IViewData>({
  name: '',
  filter: {},
  clusterNamespaces: [],
});
const displayClusterNamespaces = computed(() => {
  if (!viewData.value.clusterNamespaces?.length) {
    return [{
      clusterID: '',
      namespaces: [],
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
    if (viewData.value.filter?.[item.id]?.length) {
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
  pre[item.clusterID] = item.namespaces;
  return pre;
}, {}));

// 批量添加集群
const handleBatchAddCluster = (clusters: string[]) => {
  viewData.value.clusterNamespaces = clusters.map(clusterID => ({
    clusterID,
    namespaces: clusterNsMap.value[clusterID] || [],
  }));
};

// 当前集群是否能选择（一个集群只能选一次）
const isClusterDisabled = (item, cluster) => viewData.value.clusterNamespaces
  .some(data => data.clusterID === cluster.clusterID) && item.clusterID !== cluster.clusterID;
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
    viewData.value.filter[item.id] = Array.isArray(viewData.value.filter[item.id]) ? [] : '';
  }
};

watch(() => props.data, () => {
  if (isEqual(props.data, viewData.value)) return;

  viewData.value = merge({
    name: '',
    filter: {},
    clusterNamespaces: [],
  }, cloneDeep(props.data));
  resetFieldListStatus();
}, { immediate: true, deep: true });

watch(viewData, () => {
  emits('change', viewData.value);
}, { deep: true });
</script>
