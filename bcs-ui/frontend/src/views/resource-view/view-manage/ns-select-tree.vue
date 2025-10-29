<template>
  <bcs-select
    :loading="nsLoading"
    multiple
    :popover-min-width="360"
    searchable
    :clearable="false"
    :show-empty="false"
    :tag-fixed-height="false"
    :display-tag="displayTag"
    allow-create
    ext-popover-cls="custom-select-popover"
    selected-style="checkbox"
    v-model="nsData"
    @toggle="handleToggle"
    @selected="handleNsChange"
    @tab-remove="handleRemove">
    <template v-if="nsgroup" #trigger>
      <div class="relative">
        <div
          class="pr-[36px] pl-[10px] overflow-hidden text-nowrap h-[30px]">
          <div class="flex items-center">
            <svg width="16px" height="16px" fill="#3A84FF" class="shrink-0">
              <use xlink:href="#bcs-icon-all"></use>
            </svg>
            <span class="ml-[4px] overflow-hidden text-ellipsis">{{ typeMap.all.id === nsgroup
              ? $t('dashboard.ns.label.allNamespaces') : (typeMap.project.id === nsgroup
                ? $t('dashboard.ns.label.projectNS', [''])
                : $t('dashboard.ns.label.systemNS', [''])) }}
            </span>
          </div>
        </div>
        <span class="absolute top-0 right-[2px] h-[30px] flex flex-col justify-center">
          <LoadingIcon v-if="nsLoading" />
          <i
            v-else
            :class="[
              'bk-icon icon-angle-down transition duration-300 text-[#979ba5] text-[22px]',
              { 'rotate-[-180deg]': isToggle }
            ]"></i>
        </span>
      </div>
    </template>
    <template #search>
      <div
        v-for="(item, key) in typeMap"
        :key="key"
        :class="[
          'flex items-center cursor-pointer hover:bg-[#eaf3ff] px-[10px]',
          item.id === nsgroup ? 'bg-[#eaf3ff] text-[#3a84ff]' : ''
        ]"
        @click.stop="handleSelectAll(item)">
        <svg width="16px" height="16px" :fill="item.id === nsgroup ? '#3A84FF' : '#63656e'">
          <use xlink:href="#bcs-icon-all"></use>
        </svg>
        <span class="ml-[4px]">{{ item.name }}</span>
      </div>
      <bcs-divider class="!my-[4px] !border-b-[#c4c6cc]"></bcs-divider>
      <bcs-input
        clearable
        behavior="simplicity"
        left-icon="bk-icon icon-search"
        v-model="searchKey">
      </bcs-input>
    </template>
    <div v-show="false">
      <bcs-option
        v-for="item in nsList"
        :key="item.name"
        :id="item.name"
        :name="item.name">
      </bcs-option>
    </div>

    <bcs-big-tree
      show-checkbox
      ref="treeRef"
      default-expand-all
      @check-change="handleCheckChange">
    </bcs-big-tree>
    <template #extension>
      <SelectExtension
        :link-text="$t('dashboard.ns.create.title')"
        @link="handleGotoNs"
        @refresh="handleGetNsData" />
    </template>
  </bcs-select>
</template>
<script setup lang="ts">
import { isEqual } from 'lodash';
import { computed, PropType, ref, watch } from 'vue';

import LoadingIcon from '@/components/loading-icon.vue';
import SelectExtension from '@/components/select-extension.vue';
import { useCluster } from '@/composables/use-app';
import useDebouncedRef from '@/composables/use-debounce';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import { INamespace, useNamespace } from '@/views/cluster-manage/namespace/use-namespace';

interface IEmitsParams {
  value: string[];
  group: IGroup;
}

type IEmits = (e: 'change', params: IEmitsParams) => void;
const props = defineProps({
  clusterId: {
    type: String,
    default: '',
  },
  value: {
    type: Object as PropType<{
      list: string[];
      group: IGroup;
    }>,
    default: () => ({
      list: [],
      group: '',
    }),
  },
  displayTag: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits<IEmits>();

const { getNamespaceData } = useNamespace();

const nsData = ref<string[]>([]);
const curNsData = computed(() => nsData.value.filter(item => !!item && !groupIds.includes(item)));// 过滤全部命名空间
const searchKey = useDebouncedRef('', 300);
const treeRef = ref<any>(null);

const typeMap = ref({
  all: {
    id: 'all',
    name: $i18n.t('dashboard.ns.label.allNamespaces'),
  },
  project: {
    id: 'all-user',
    name: $i18n.t('dashboard.ns.label.projectNS', []),
  },
  system: {
    id: 'all-system',
    name: $i18n.t('dashboard.ns.label.systemNS', []),
  },
});
// 全部项目命名空间
const allProjectNs = ref<string[]>([]);
// 全部系统命名空间
const allSystemNs = ref<string[]>([]);
// 分类id，大写字母开头，避免与命名空间同名（命名空间不允许大写字母开头）
const groupIds = ['GROUP-1', 'GROUP-2'];
// const selectValue = ref([]);
const treeData = ref<{
  id: string;
  name: string;
  level: number;
  children: {
    id: string;
    name: string;
    level: number;
  }[];
}[]>([
  {
    id: 'GROUP-1',
    name: '',
    level: 0,
    children: [],
  },
  {
    id: 'GROUP-2',
    name: '',
    level: 0,
    children: [],
  },
]);
const nsgroup = ref<IGroup>('');

function handleCheckChange(ids: string[]) {
  nsgroup.value = '';
  nsData.value = [...ids];
}
// 全选操作
function handleSelectAll(item) {
  nsgroup.value = item.id;
  // todo，直接调用 removeChecked 无效，需要先调用 setChecked，可能是响应式问题
  treeRef.value?.setChecked?.(nsData.value);
  treeRef.value?.removeChecked?.({ emitEvent: false });
  nsData.value = [];
}

function setData() {
  if (isEqual(props.value.list, curNsData.value) && nsgroup.value) return;
  if (!props.value?.list?.length) {
    nsData.value = [];// 全部命名空间逻辑
  } else {
    nsData.value = JSON.parse(JSON.stringify(props.value.list));
  }
  nsgroup.value = props.value.group;
}

const handleNsChange = (nsList) => {
  const last = nsList[nsList.length - 1];
  // 移除全选
  if (last) {
    nsData.value = nsData.value.filter(item => !!item);
  } else {
    nsData.value = [];
  }
};

// 失去焦点才触发
const isToggle = ref(false);
function handleToggle(val: boolean) {
  isToggle.value = val;
  if (!isToggle.value) {
    emits('change', {
      value: curNsData.value,
      group: !curNsData.value.length && !nsgroup.value ? 'all-user' : nsgroup.value,
    });
  }
}

// 清除
function handleRemove({ name }) {
  const index = nsData.value.findIndex(v => v === name);
  nsData.value.splice(index, 1);
  // todo，直接调用 setChecked 无效，需要先调用 removeChecked
  treeRef.value.setData(treeData.value);
  treeRef.value?.removeChecked?.({ emitEvent: false });
  treeRef.value?.setChecked?.(nsData.value);
  if (!curNsData.value.length) {
    nsgroup.value = 'all-user';
  }
  if (isToggle.value) return;
  emits('change', {
    value: curNsData.value,
    group: nsgroup.value,
  });
}

// 组件使用的数据
const nsList = ref<Array<INamespace>>([]);
const nsLoading = ref(false);
const { clusterList } = useCluster();
const handleGetNsData = async () => {
  const exist = clusterList.value.find(item => item.clusterID === props.clusterId);
  if (!exist) return;
  nsLoading.value = true;
  nsList.value = await getNamespaceData({ $clusterId: props.clusterId });
  nsList.value.forEach((item) => {
    const obj = {
      id: item.name,
      name: item.name,
      level: 1,
    };
    if (item.isSystem) {
      treeData.value[1].children.push(obj);
    } else {
      treeData.value[0].children.push(obj);
    }
  });
  treeData.value[0].name = $i18n.t('dashboard.ns.label.projectNS', [` (${treeData.value[0].children.length})`]);
  treeData.value[1].name = $i18n.t('dashboard.ns.label.systemNS', [` (${treeData.value[1].children.length})`]);
  allProjectNs.value = treeData.value[0].children.map(item => item.id);
  allSystemNs.value = treeData.value[1].children.map(item => item.id);
  nsLoading.value = false;
  // 保证有数据后再设置值，否则 collapse-tag 数字不显示
  setData();
  treeRef.value?.setData?.(treeData.value);

  if (!nsgroup.value) {
    treeRef.value?.setChecked?.(nsData.value);
  }
};

// 跳转命名空间
const handleGotoNs = () => {
  const { href } = $router.resolve({
    name: 'createNamespace',
    params: {
      clusterId: props.clusterId,
    },
  });
  window.open(href);
};

const watchOnce = watch(() => props.value, () => {
  if (!nsList.value.length) return;
  setData();
  watchOnce();
});

watch(() => props.clusterId, () => {
  handleGetNsData();
}, { immediate: true });

// 前端搜索
watch(searchKey, () => {
  treeRef.value?.filter?.(searchKey.value);
});
</script>
<style lang="postcss" scoped>
:deep(.bk-select-tag-container) {
  max-height: 400px !important;
}
:deep(.bk-big-tree-node .node-options .node-checkbox.is-checked:after) {
  box-sizing: content-box;
}
:deep(.bk-big-tree-node .node-content) {
  font-size: 12px;
}
</style>
