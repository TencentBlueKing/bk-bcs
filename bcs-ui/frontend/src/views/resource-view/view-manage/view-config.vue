<template>
  <div class="view-config text-[12px]" ref="panelRef" v-bkloading="{ isLoading }">
    <div class="resize-line" ref="resizeRef" @mousedown="onMousedownEvent"></div>
    <!-- 查看模式 -->
    <div class="flex flex-col max-h-[100%] pt-[24px]" v-if="!isEdit">
      <div class="flex-1 overflow-auto px-[24px] pb-[24px]">
        <!-- 视图详情 -->
        <div class="bg-[#F5F7FA] rounded-sm px-[8px] py-[6px]">
          <div
            class="flex items-center justify-between cursor-pointer select-none h-[20px]"
            @click="toggleCollapse">
            <!-- 视图名称 -->
            <div class="flex items-center flex-1 pr-[16px]">
              <span
                class="transition-all relative top-[-1px] text-[#979BA5] mr-[6px]"
                :style="collapse ? 'transform: rotate(-90deg);' : 'transform: rotate(0deg);'">
                <i class="bcs-icon bcs-icon-down-shape"></i>
              </span>
              <span class="bcs-ellipsis font-bold">{{ $t('view.labels.viewDataRange') }}</span>
              <i
                class="bcs-icon bcs-icon-alarm-insufficient text-[14px] text-[#FFB848] ml-[8px]"
                v-if="unknownClusterID"
                v-bk-tooltips="$t('view.tips.invalidate')">
              </i>
            </div>
            <!-- 视图克隆、删除和编辑 -->
            <div class="flex items-center" v-if="originViewData.id" @click.stop>
              <span
                :class="[
                  'flex items-center text-[12px]',
                  isEditHover ? 'text-[#3a84ff]' : ''
                ]"
                @click="handleEditView"
                @mouseenter="isEditHover = true"
                @mouseleave="isEditHover = false">
                <i
                  :class="[
                    'bk-icon icon-edit-line text-[#979BA5] mr-[6px]',
                    isEditHover ? '!text-[#3a84ff]' : ''
                  ]">
                </i>
                {{ $t('view.button.edit') }}
              </span>
              <PopoverSelector offset="0, 10">
                <span class="bcs-icon-more-btn w-[20px] h-[20px] ml-[16px]">
                  <i class="bcs-icon bcs-icon-more"></i>
                </span>
                <template #content>
                  <ul class="bg-[#fff]">
                    <li class="bcs-dropdown-item" @click="handleCloneView">{{ $t('view.button.clone') }}</li>
                    <li class="bcs-dropdown-item" @click="handleDeleteView">{{ $t('view.button.delete') }}</li>
                  </ul>
                </template>
              </PopoverSelector>
            </div>
          </div>
          <div v-if="!collapse" class="mt-[8px] px-[8px]">
            <ViewField
              :title="$t('view.labels.clusterAndNs')"
              :deletable="false"
              :active="activeField === 'clusterNamespaces'"
              class="rounded-sm p-[8px]">
              <template v-if="!isClusterMode">
                <bcs-tag
                  :class="['m-[0px]', index > 0 ? 'mt-[6px]' : '']"
                  v-for="(item, index) in originViewData?.clusterNamespaces"
                  :key="index">
                  <span
                    :class="[
                      'bcs-ellipsis',
                      !clusterNameMap[item.clusterID] ? '!text-[#979BA5] line-through': ''
                    ]"
                    v-bk-tooltips="{
                      content: !clusterNameMap[item.clusterID]
                        ? $t('view.tips.invalidate')
                        : `${clusterNameMap[item.clusterID] || item.clusterID} / ${item.namespaces.join(', ')
                          || $t('view.labels.allNs')}`
                    }">
                    {{
                      `${clusterNameMap[item.clusterID] || item.clusterID} / ${item.namespaces.join(', ')
                        || $t('view.labels.allNs')}`
                    }}
                  </span>
                </bcs-tag>
              </template>
              <template v-else>
                <bcs-tag class="m-[0px]">
                  <span class="bcs-ellipsis" v-bk-overflow-tips>
                    {{`${curViewName} / ${$t('view.labels.allNs')}`}}
                  </span>
                </bcs-tag>
              </template>
            </ViewField>
            <ViewField
              :title="$t('view.labels.creator')"
              :deletable="false"
              :active="activeField === 'creator'"
              class="rounded-sm p-[8px] mt-[8px]"
              v-if="originViewData?.filter?.creator?.length">
              <bcs-tag
                v-for="(item, index) in originViewData?.filter.creator"
                :key="index"
                :class="['m-[0px] mr-[8px]', index > 0 ? 'mt-[6px]' : '']">
                {{ item }}
              </bcs-tag>
            </ViewField>
            <ViewField
              :title="$t('k8s.label')"
              :deletable="false"
              :active="activeField === 'labelSelector'"
              class="rounded-sm p-[8px] mt-[8px]"
              v-if="originViewData.filter?.labelSelector?.length">
              <div
                :class="['flex items-center', index > 0 ? 'mt-[6px]' : '']"
                v-for="(item, index) in originViewData.filter.labelSelector"
                :key="index">
                <span
                  class="flex items-center justify-center w-[26px] h-[22px] text-[#3A84FF] mr-[4px] bcs-border"
                  v-if="index > 0">
                  &
                </span>
                <bcs-tag class="m-[0px]">
                  <span class="bcs-ellipsis" v-bk-overflow-tips>
                    {{ item.key }}
                    <span class="text-[#FF9C01]">{{ item.op }}</span>
                    {{ item.values.join(',') }}
                  </span>
                </bcs-tag>
              </div>
            </ViewField>
            <ViewField
              :title="$t('view.labels.resourceName')"
              :deletable="false"
              :active="activeField === 'name'"
              class="rounded-sm p-[8px] mt-[8px]"
              v-if="originViewData.filter?.name">
              <bcs-tag class="m-[0px]">
                <span class="bcs-ellipsis" v-bk-overflow-tips>
                  {{ originViewData.filter?.name }}
                </span>
              </bcs-tag>
            </ViewField>
            <ViewField
              :title="$t('generic.label.source')"
              :deletable="false"
              :active="activeField === 'source'"
              class="rounded-sm p-[8px] mt-[8px]"
              v-if="originViewData.filter?.createSource?.source">
              <bcs-tag class="m-[0px]">
                <span class="bcs-ellipsis" v-bk-overflow-tips>
                  <span>{{ originViewData.filter?.createSource?.source || '--' }}</span>
                  <span v-if="originViewData.filter?.createSource?.source === 'Helm'">
                    <span v-if="originViewData.filter?.createSource?.chart?.chartName">/ </span>
                    {{ originViewData.filter?.createSource?.chart?.chartName }}
                  </span>
                  <span v-else-if="originViewData.filter?.createSource?.source === 'Template'">
                    <span v-if="originViewData.filter?.createSource?.template?.templateName">/ </span>
                    <span>{{ `${originViewData.filter?.createSource?.template?.templateName}${
                      originViewData.filter?.createSource?.template?.templateVersion
                        ? `:${originViewData.filter?.createSource?.template?.templateVersion}`
                        : ''}` }}</span>
                  </span>
                </span>
              </bcs-tag>
            </ViewField>
          </div>
        </div>
        <!-- 临时条件 -->
        <ViewForm
          :data="parseCurTmpViewData"
          :show-cluster-field="false"
          :add-field-text="$t('view.button.addQueryField')"
          @change="handleUpdateTmpViewData"
          @field-status-change="handleFieldStatusChange" />
      </div>
      <!-- 视图操作 -->
      <div class="flex items-center justify-between sticky bottom-0 bg-[#fff] py-[8px] px-[24px]">
        <div class="flex items-center">
          <bcs-button
            theme="primary"
            class="min-w-[88px]"
            v-bk-trace.click="{
              module: 'view',
              operation: 'query',
              desc: '视图查询操作',
              username: $store.state.user.username,
              projectCode: $store.getters.curProjectCode,
            }"
            @click="handleQuery">
            {{ $t('view.button.query') }}
          </bcs-button>
          <bcs-button
            class="min-w-[88px]"
            v-if="!originViewData.id"
            @click="handleShowSaveAsDialog">
            {{ $t('view.button.saveAsView') }}
          </bcs-button>
        </div>
        <bcs-button
          class="min-w-[88px]"
          v-if="showResetBtn"
          @click="() => handleUpdateTmpViewData()">
          {{ $t('view.button.reset') }}
        </bcs-button>
      </div>
    </div>
    <!-- 编辑模式 -->
    <div class="flex flex-col max-h-[100%]" v-else>
      <div class="flex items-center px-[24px] pt-[16px] pb-[24px]">
        <i
          class="bcs-icon bcs-icon-arrows-left text-[#3A84FF] text-[16px] font-bold cursor-pointer"
          @click="() => isEdit = false">
        </i>
        <span class="text-[#313238] text-[16px] ml-[8px]">
          {{ newViewData?.id ? $t('view.labels.editView') : $t('view.labels.newView') }}
        </span>
      </div>
      <div class="flex-1 overflow-auto px-[24px] pb-[24px]">
        <ViewField
          :title="$t('view.labels.name')"
          :deletable="false">
          <bcs-input v-model.trim="viewName" :maxlength="64" clearable></bcs-input>
        </ViewField>
        <bcs-divider class="!mt-[24px] !mb-[4px]"></bcs-divider>
        <div class="font-bold leading-[20px]">{{ $t('view.labels.viewDataRange') }}</div>
        <ViewForm
          :data="newViewData"
          :add-field-text="$t('view.button.addField')"
          class="mt-[16px]"
          @change="handleViewDataChange" />
      </div>
      <!-- 视图操作 -->
      <div class="flex items-center sticky bottom-0 bg-[#fff] py-[8px] px-[24px]">
        <bcs-button
          theme="primary"
          class="min-w-[88px]"
          :loading="saving"
          :disabled="disabledSaveBtn"
          v-bk-trace.click="{
            module: 'view',
            operation: 'save',
            desc: '视图保存操作',
            username: $store.state.user.username,
            projectCode: $store.getters.curProjectCode,
          }"
          @click="handleSaveView">
          {{ $t('generic.button.save') }}
        </bcs-button>
        <bcs-button
          class="min-w-[88px]"
          @click="() => isEdit = false">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </div>
    </div>
    <!-- 另存为 -->
    <bcs-dialog
      v-model="showSaveAs"
      header-position="left"
      :title="$t('view.labels.saveAs')">
      <bcs-form form-type="vertical" class="mt-[-16px]">
        <bcs-form-item :label="$t('view.labels.name')" required>
          <bcs-input v-model.trim="viewName" :maxlength="64"></bcs-input>
        </bcs-form-item>
      </bcs-form>
      <template #footer>
        <div>
          <bcs-button
            :loading="saving"
            :disabled="!viewName"
            theme="primary"
            v-bk-trace.click="{
              module: 'view',
              operation: 'save',
              desc: '视图保存操作',
              username: $store.state.user.username,
              projectCode: $store.getters.curProjectCode,
            }"
            @click="handleSaveAs(false)">
            {{ $t('generic.button.save') }}
          </bcs-button>
          <bcs-button
            :loading="saving"
            :disabled="!viewName"
            v-bk-trace.click="{
              module: 'view',
              operation: 'save',
              desc: '视图保存操作',
              username: $store.state.user.username,
              projectCode: $store.getters.curProjectCode,
            }"
            @click="handleSaveAs">
            {{ $t('view.button.confirmAndChangeView') }}
          </bcs-button>
          <bcs-button @click="showSaveAs = false">{{ $t('generic.button.cancel') }}</bcs-button>
        </div>
      </template>
    </bcs-dialog>
  </div>
</template>

<script lang="ts" setup>
import { cloneDeep, debounce } from 'lodash';
import { computed, inject, onBeforeMount, onBeforeUnmount, ref, watch } from 'vue';

import useViewConfig from './use-view-config';
import ViewField from './view-field.vue';
import ViewForm from './view-form.vue';

import $bkMessage from '@/common/bkmagic';
import { bus } from '@/common/bus';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import PopoverSelector from '@/components/popover-selector.vue';
import { useCluster } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $store from '@/store';

const emits = defineEmits(['close']);

const { clusterNameMap } = useCluster();

// resize event
const panelRef = ref();
const resizeRef = ref();
const maxWidth = 640;
const minWidth = 400;
const onMousedownEvent = (e: MouseEvent) => {
  // 颜色改变提醒
  resizeRef.value.style.borderRight = '1px solid #3a84ff';
  const startX = e.clientX;
  const { clientWidth } = panelRef.value;
  // 鼠标拖动事件
  document.onmousemove =  (e) => {
    if (!panelRef.value) return;
    document.body.style.userSelect = 'none';
    const endX = e.clientX;
    const moveLen = startX - endX;
    const width = clientWidth - moveLen;
    if (width <= (minWidth - 20)) { // 预留20的操作空间，防止一点击就是关闭
      panelRef.value.style.width = `${minWidth}px`;
      emits('close');
    } else if (width > maxWidth) {
      panelRef.value.style.width = `${maxWidth}px`;
    } else {
      panelRef.value.style.width = `${width}px`;
    }
  };
  // 鼠标松开事件
  document.onmouseup =  () => {
    document.body.style.userSelect = '';
    document.onmousemove = null;
    document.onmouseup = null;
    if (resizeRef.value) {
      resizeRef.value.style.borderRight = '';
      resizeRef.value?.releaseCapture?.();
    }
  };
  resizeRef.value?.setCapture?.();
};

const {
  isClusterMode,
  dashboardViewID,
  curTmpViewData,
  curViewData,
  curViewName,
  getViewConfigDetail,
  createViewConfig,
  deleteViewConfig,
  getViewConfigList,
  updateViewConfig,
  updateViewIDStore,
} = useViewConfig();

const isEdit = ref(false);
const isLoading = ref(false);

// 原始数据
const initViewData = {
  name: '',
  filter: {},
  clusterNamespaces: [],
};
const originViewData = ref<IViewData>(cloneDeep(initViewData));
// 当前视图临时数据
const parseCurTmpViewData = computed(() => {
  const data = cloneDeep(curTmpViewData.value);
  // hack 查看模式时标签接口依赖集群和命名空间
  if (data && !data.clusterNamespaces?.length) {
    if (isClusterMode.value) {
      // 集群模式
      data.clusterNamespaces = curViewData.value?.clusterNamespaces;
    } else {
      // 自定义视图
      data.clusterNamespaces = originViewData.value?.clusterNamespaces;
    }
  }
  return data;
});

// 更新视图临时条件数据
const handleUpdateTmpViewData = debounce((data: IViewData|undefined = undefined) => {
  $store.commit('updateTmpViewData', cloneDeep(data));
  // 数据上报
  window.BkTrace?.startReported({
    module: 'view',
    operation: 'auto-query',
    desc: '视图输入查询',
    username: $store.state.user.username,
    projectCode: $store.getters.curProjectCode,
  });
}, 300);

const showResetBtn = ref(false);
const handleFieldStatusChange = (data: IFieldItem[]) => {
  showResetBtn.value = data?.some(item => item.status === 'added');
};

// 获取详情
const handleGetDetail = async (id: string) => {
  isLoading.value = true;
  let data: IViewData;
  if (id) {
    data = await getViewConfigDetail(id);
  } else {
    data = cloneDeep(initViewData);
  }
  originViewData.value = cloneDeep(data);
  isLoading.value = false;
};

// 切换视图
const viewConfigPopoverRef = ref();
const viewChange = async (id: string) => {
  await handleGetDetail(id);
  updateViewIDStore(id);// 会触发路由发生变化
  viewConfigPopoverRef.value?.hide();
};

// 视图详情展开和折叠
const collapse = ref(true);
const toggleCollapse = () => {
  collapse.value = !collapse.value;
};

// 编辑视图
const isEditHover = ref(false);
const newViewData = ref<IViewData>();
watch(isEdit, () => {
  if (!isEdit.value) {
    cancelEdit();
  }
  $store.commit('updateViewEditable', isEdit.value);
});
const handleEditView = () => {
  viewName.value = originViewData.value?.name || '';
  newViewData.value = cloneDeep(originViewData.value);
  isEdit.value = true;
};
const handleViewDataChange = (v: IViewData) => {
  newViewData.value = v;
  // 修改临时条件
  $store.commit('updateEditViewData', v);
};
const cancelEdit = () => {
  isEdit.value = false;
  newViewData.value = undefined;
  $store.commit('updateEditViewData', {});
  $store.commit('updateViewEditable', false);
};

// 克隆视图
const handleCloneView = () => {
  viewName.value = `view-${Math.floor(Date.now() / 1000)}`;
  newViewData.value = !isClusterMode.value ? cloneDeep(originViewData.value) : cloneDeep(curViewData.value);
  delete newViewData.value?.id; // 删除ID
  isEdit.value = true;
};

// 校验集群ID正确性
const unknownClusterID = computed(() => {
  let data: IClusterNamespace[] = [];
  if (isEdit.value) {
    data = newViewData.value?.clusterNamespaces || [];
  } else {
    data = originViewData.value?.clusterNamespaces || [];
  }
  return data?.some((item) => {
    const { clusterID } = item;
    return !clusterNameMap.value[clusterID];
  });
});
// 校验保存按钮
const disabledSaveBtn = computed(() => unknownClusterID.value
|| !newViewData.value?.clusterNamespaces?.length || !viewName.value);
const handleSaveView = async () => {
  if (!newViewData.value?.clusterNamespaces?.length || unknownClusterID.value) return;

  let result;
  if (newViewData.value?.id) {
    result = await handleModifyView({
      ...newViewData.value,
      name: viewName.value,
    });
  } else {
    result = await handleCreateView({
      ...newViewData.value,
      name: viewName.value,
    });
  }

  if (result) {
    isEdit.value = false;
    handleUpdateTmpViewData();
  }
};

// 修改视图
const handleModifyView = async (data: IViewData) => {
  if (!data?.id) return;

  saving.value = true;
  const result = await updateViewConfig({
    ...data,
    $id: data.id,
  });
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.save'),
    });
    await Promise.all([
      handleGetDetail(data.id),
      getViewConfigList(),
    ]);
  }
  saving.value = false;
  return result;
};

// 创建视图
const handleCreateView = async (data: Partial<IViewData>) => {
  if (!data?.name) return;
  saving.value = true;
  const result = await createViewConfig(data);
  if (result?.id) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('view.status.saveAs'),
    });
    showSaveAs.value = false;
    await viewChange(result?.id);
    await getViewConfigList();
  }
  saving.value = false;
  return result?.id;
};

// 删除视图
const handleDeleteView = () => {
  if (!originViewData.value.id) return;

  $bkInfo({
    type: 'warning',
    title: $i18n.t('view.tips.confirmDelete'),
    clsName: 'custom-info-confirm default-info',
    subTitle: originViewData.value.name,
    confirmFn: async () => {
      const result = await deleteViewConfig({ $id: originViewData.value.id || '' });
      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.delete'),
        });
        await viewChange('');
        await getViewConfigList();
      }
    },
  });
};

// 查询
const { reload } = inject<any>('dashboard-view') || {};
const handleQuery = () => {
  reload?.();
};

// 另存为
const viewName = ref('');
const showSaveAs = ref(false);
const saving = ref(false);
const handleShowSaveAsDialog = () => {
  viewName.value = `view-${Math.floor(Date.now() / 1000)}`;
  showSaveAs.value = true;
};
const handleSaveAs = async (changeView = true) => {
  saving.value = true;
  const result = await createViewConfig({
    ...curViewData.value,
    name: viewName.value,
  });
  if (result?.id) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('view.status.saveAs'),
    });
    showSaveAs.value = false;
    await getViewConfigList();
  }
  saving.value = false;
  if (result?.id && changeView) {
    viewChange(result?.id);
  }
};

watch(dashboardViewID, () => {
  viewChange(dashboardViewID.value);
});

// 高亮点击的字段
const activeField = ref<'clusterNamespaces'|'creator'|'labelSelector'|'name'|'source'>();
let timeoutID;
onBeforeMount(() => {
  bus.$on('locate-to-field', (field) => {
    if (!field) return;

    timeoutID && clearTimeout(timeoutID);
    activeField.value = field;
    collapse.value = false;
    timeoutID = setTimeout(() => {
      activeField.value = undefined;
    }, 2000);
  });
});

onBeforeUnmount(() => {
  bus.$off('locate-to-field');
  cancelEdit();
});

onBeforeMount(() => {
  handleGetDetail(dashboardViewID.value);
});

// 创建新视图
const createNewView = () => {
  viewName.value = '';
  newViewData.value = cloneDeep(initViewData);
  isEdit.value = true;
};
// 编辑视图
const editView = async (id: string) => {
  if (!id) return;

  await viewChange(id);
  handleEditView();
};

defineExpose({
  cancelEdit,
  createNewView,
  editView,
});
</script>

<style lang="postcss" scoped>
.view-config {
  background: #fff;
  border-radius: 2px;
  width: 400px;
  position: relative;
  .resize-line {
    width: 8px;
    height: 100%;
    cursor: col-resize;
    position: absolute;
    right: 0px;
    z-index: 2;
    &::after {
      content: "";
      position: absolute;
      width: 2px;
      height: 2px;
      color: #63656E;
      background: #63656E;
      box-shadow: 0 4px 0 0 #63656E,0 8px 0 0 #63656E,0 -4px 0 0 #63656E,0 -8px 0 0 #63656E;
      left: 2px;
      top: 50%;
      transform: translate3d(0, -50%, 0);
    }
  }
  .view-name {
    border: 1px solid #699DF4;
  }
}

.bcs-border-top {
  border-top: 1px solid #DCDEE5;
}
>>> .add-filter .bk-tooltip-ref {
  width: 100%;
}

>>> .label-selector .key-value:last-child {
  margin-bottom: 0px;
}

>>> .name-validate .error-tip {
  top: 4px !important;
}

>>> .view-name-popover .bk-tooltip-ref {
  width: 100%;
}

>>> .bk-dialog-footer {
  padding: 8px 24px;
}
</style>
