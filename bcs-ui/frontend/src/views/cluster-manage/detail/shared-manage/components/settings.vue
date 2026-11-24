<template>
  <div v-bkloading="{ isLoading: initLoading, opacity: 1 }" class="py-[16px] px-[20px] border-[#dcdee5] border border-t-0">
    <!-- 共享状态（对齐原型 .form-sec：margin 8px 0 16px / line-height 16px / flex 居中） -->
    <div class="flex items-center mt-[8px] mb-[16px] border-l-[3px] border-[#3A84FF] pl-[8px] text-[14px] font-semibold leading-[16px] text-[#313238]">
      {{ $t('cluster.sharedManage.summary.sharedStatus') }}
    </div>
    <bcs-form class="bcs-small-form" :label-width="136">
      <bcs-form-item :label="$t('cluster.sharedManage.settings.enableShared')">
        <div class="flex items-start">
          <bcs-switcher
            :key="switcherKey"
            class="mt-[2px]"
            :value="shared"
            @change="handleToggleShared" />
          <span class="ml-[10px] text-[12px] leading-[20px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.settings.enableSharedTips') }}
          </span>
        </div>
      </bcs-form-item>

      <!-- 可见范围 -->
      <bcs-form-item
        v-if="shared"
        :label="$t('cluster.sharedManage.summary.visibleRange')"
        :required="true">
        <!-- 存量：全平台 / 按业务 只读，需转为指定项目后才能编辑 -->
        <template v-if="rangeMode === 'legacy'">
          <bcs-alert
            type="warning"
            class="mb-[8px]"
            :title="$t('cluster.sharedManage.settings.legacyAlert', { range: legacyRangeText })">
          </bcs-alert>
          <div class="flex items-center gap-[8px]">
            <span>{{ legacyRangeText }}</span>
            <bcs-tag class="m-0 px-[5px]" theme="info">{{ $t('cluster.sharedManage.settings.readonly') }}</bcs-tag>
          </div>
          <div class="mt-[8px]">
            <bcs-button @click="handleConvertRange">
              {{ $t('cluster.sharedManage.settings.convert') }}
            </bcs-button>
          </div>
        </template>
        <!-- 指定项目 -->
        <template v-else>
          <div class="flex flex-wrap items-center gap-[6px]">
            <bcs-tag
              v-for="id in rangeIds"
              :key="id"
              class="m-[0px] pl-[8px] h-[26px] leading-[26px] bg-[#f0f1f5]"
              closable
              @close="handleRemoveRange(id)">
              <span class="text-[12px]">{{ projectName(id) }}</span>
              <em
                v-if="projectNsCount(id)"
                class="not-italic ml-[6px] text-[11px] leading-[16px] px-[4px] rounded-[2px] bg-[#E1ECFF] text-[#3A84FF]">
                {{ $t('cluster.sharedManage.settings.projectNsCount', { num: projectNsCount(id) }) }}
              </em>
            </bcs-tag>
            <bcs-select
              :key="selectKey"
              class="w-[200px]"
              :clearable="false"
              searchable
              :placeholder="$t('cluster.sharedManage.settings.addProject')"
              :value="''"
              :loading="projectLoading"
              size="small"
              @change="handleAddRange">
              <bcs-option
                v-for="option in addableProjects"
                :key="option.projectID"
                :id="option.projectID"
                :name="option.name" />
            </bcs-select>
          </div>
          <p v-if="rangeErr" class="mt-[4px] text-[12px] text-[#EA3636]">{{ rangeErr }}</p>
          <p v-if="legacyFrom" class="mt-[8px] text-[12px] text-[#FF9C01]">
            {{ $t('cluster.sharedManage.settings.legacyConvertTips', { range: legacyRangeText }) }}
          </p>
          <p class="mt-[8px] text-[12px] text-[#979BA5]">
            {{ $t('cluster.sharedManage.settings.rangeTips') }}
          </p>
        </template>
      </bcs-form-item>
    </bcs-form>

    <!-- 底部操作 -->
    <div class="pt-[12px] pl-[136px] border-t border-[#DCDEE5] flex items-center gap-[8px]">
      <bcs-button
        theme="primary"
        :loading="saving"
        :disabled="!dirty"
        @click="handleSave">
        {{ $t('generic.button.save') }}
      </bcs-button>
      <bcs-button :disabled="!dirty" @click="handleCancel">
        {{ $t('generic.button.cancel') }}
      </bcs-button>
      <span v-if="dirty" class="text-[12px] text-[#FF9C01]">
        {{ $t('cluster.sharedManage.settings.unsaved') }}
      </span>
    </div>

    <!-- 关闭共享二次确认（仍持有归属命名空间时） -->
    <SharedConfirmDialog
      :value="closeGuardShow"
      :title="$t('cluster.sharedManage.settings.closeSharedTitle')"
      :width="560"
      :desc="$t('cluster.sharedManage.settings.closeSharedBody', { num: heldNamespaces.length })"
      :confirm-text="$t('cluster.sharedManage.settings.forceClose')"
      :secondary-text="$t('cluster.sharedManage.settings.goRecycle')"
      :namespaces="heldNamespacesItems"
      @input="handleCloseDialog('closeGuardShow', $event)"
      @confirm="confirmCloseShared"
      @secondary="handleGoRecycle" />

    <!-- 关闭共享二次确认（无归属命名空间） -->
    <SharedConfirmDialog
      :value="confirmCloseShow"
      :title="$t('cluster.sharedManage.settings.closeSharedTitle')"
      :width="420"
      :confirm-text="$t('cluster.sharedManage.settings.forceClose')"
      :desc="$t('cluster.sharedManage.settings.closeSharedSimpleBody')"
      @input="handleCloseDialog('confirmCloseShow', $event)"
      @confirm="confirmCloseShared" />

    <!-- 移出仍持有命名空间的项目二次确认 -->
    <SharedConfirmDialog
      :value="removeGuardShow"
      :title="$t('cluster.sharedManage.settings.removeTitle', { name: projectName(pendingRemoveId) })"
      :width="560"
      :confirm-text="$t('cluster.sharedManage.settings.forceRemove')"
      :secondary-text="$t('cluster.sharedManage.settings.goRecycle')"
      :desc="$t('cluster.sharedManage.settings.removeBody', { num: projectNsCount(pendingRemoveId) })"
      :namespaces="pendingRemoveNsItems"
      @input="handleCloseDialog('removeGuardShow', $event)"
      @confirm="confirmRemoveRange"
      @secondary="handleGoRecycle" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import SharedConfirmDialog from './confirm-dialog.vue';

import { getSharedprojects, modifyCluster } from '@/api/modules/cluster-manager';
import { INativeNamespace } from '@/api/modules/shared-manage';
import $bkMessage from '@/common/bkmagic';
import { ICluster, IProject, useAppData } from '@/composables/use-app';
import useProjects from '@/views/project-manage/project/use-project';

// 共享设置：共享状态 + 可见范围（指定项目），存量只读
// 交互与数据校验参考原型 settingsHtml()
const props = defineProps({
  clusterData: {
    type: Object as () => Partial<ICluster>,
    default: () => ({}),
  },
  namespaceList: {
    type: Array as () => INativeNamespace[],
    default: () => [],
  },
});

const emits = defineEmits(['refresh', 'go-recycle']);

const { user } = useAppData();
const { getProjectList } = useProjects();

// 已保存的基线（用于 dirty 判断与取消重置）
const savedShared = computed(() => !!props.clusterData?.is_shared);
const savedRangeIds = computed<string[]>(() => props.clusterData?.sharedRanges?.projectIdOrCodes || []);
const savedBizs = computed<string[]>(() => props.clusterData?.sharedRanges?.bizs || []);

// 编辑态
const shared = ref(false);
const rangeIds = ref<string[]>([]);
const legacyFrom = ref(''); // 存量转为指定项目后记录原范围：all / biz
const rangeErr = ref('');
const saving = ref(false);

watch(() => props.clusterData, () => {
  shared.value = savedShared.value;
  rangeIds.value = [...savedRangeIds.value];
  legacyFrom.value = '';
  rangeErr.value = '';
}, { immediate: true });

const dirty = computed(() => shared.value !== savedShared.value
  || JSON.stringify(rangeIds.value) !== JSON.stringify(savedRangeIds.value));

// 存量共享范围：全平台（无 projectIdOrCodes）/ 按业务（bizs）
const legacyMode = computed(() => {
  if (!savedShared.value || savedRangeIds.value.length) return '';
  return savedBizs.value.length ? 'biz' : 'all';
});
// 编辑模式：legacy 只读，project 可编辑（转换后 legacyFrom 非空）
const rangeMode = computed(() => (legacyMode.value && !legacyFrom.value ? 'legacy' : 'project'));

const legacyRangeText = computed(() => {
  if (legacyMode.value === 'biz') {
    return `${window.i18n.t('cluster.sharedManage.settings.bizRange')}（${savedBizs.value.join('、')}）`;
  }
  return window.i18n.t('cluster.sharedManage.summary.wholePlatform');
});

// 可见范围项目详情（GetClusterSharedProject），用于回显项目名
const sharedProjectList = ref<IProject[]>([]);
// 手动记录新增项目的名称（避免接口列表不全时 tag 回显原始 ID）
const projectNameMap = ref<Record<string, string>>({});
const projectName = (idOrCode: string) => projectNameMap.value[idOrCode]
  || sharedProjectList.value.find(project => (
    project.projectID === idOrCode || project.projectCode === idOrCode
  ))?.name
  || projectList.value.find(project => project.projectID === idOrCode)?.name
  || idOrCode;

const projectLoading = ref(false);
// 初始化聚合 loading（项目列表 + 可见范围），覆盖整个表单
const initLoading = ref(false);
const projectList = ref<IProject[]>([]);
const addableProjects = computed(() => projectList.value.filter(project => (
  project.projectID !== props.clusterData?.projectID
    && !rangeIds.value.includes(project.projectID)
)));

// 分页拉取全量项目，用于选择器选项及项目名回显
const initProjectList = async () => {
  projectLoading.value = true;
  const list: IProject[] = [];
  let offset = 0;
  let total = 0;
  do {
    const { data } = await getProjectList({ offset, limit: 100 })
      .catch(() => ({ data: { results: [] as IProject[], total: 0 } }));
    list.push(...(data?.results || []));
    total = data?.total || 0;
    offset += 100;
  } while (offset < total);
  projectList.value = list;
  projectLoading.value = false;
  // 预填充 id -> name 映射（存量转换 / 已保存范围回显用）
  list.forEach((project) => {
    projectNameMap.value[project.projectID] = project.name;
  });
};
const initSharedProjectList = async () => {
  if (!props.clusterData?.clusterID) return;
  sharedProjectList.value = await getSharedprojects({
    $clusterId: props.clusterData.clusterID,
  }).catch(() => [] as IProject[]);
};

// 命名空间归属统计
const projectNsCount = (idOrCode: string) => props.namespaceList.filter(ns => (
  ns.projectID === idOrCode || ns.projectCode === idOrCode
)).length;
// 已分配给其他项目的命名空间（关闭共享保护校验用）
const heldNamespaces = computed(() => props.namespaceList.filter(ns => (
  ns.projectID && ns.projectID !== props.clusterData?.projectID
)));
// 命名空间明细项（名称 + 归属项目名 + 可选状态标签，对齐原型 nsList）
const toNsItem = (ns: INativeNamespace) => ({
  name: ns.name,
  project: projectName(ns.projectID) || ns.projectCode,
  // vcluster 命名空间 / 后端扩展字段未上线时不展示审批状态标签
  tag: ns.itsmTicketStatus ? window.i18n.t('cluster.sharedManage.settings.pendingReview') : '',
});
// 关闭共享弹窗明细
const heldNamespacesItems = computed(() => heldNamespaces.value.map(toNsItem));
// 待移出项目持有的命名空间明细（移出项目弹窗用）
const pendingRemoveNsItems = computed(() => props.namespaceList
  .filter(ns => ns.projectID === pendingRemoveId.value || ns.projectCode === pendingRemoveId.value)
  .map(toNsItem));

// 开关共享（关闭时做前端保护确认）
// bcs-switcher 非受控组件：点击后内部状态立即变更，即使不更新 shared 也不会回退。
// 通过自增 switcherKey 强制重挂载，使开关视觉状态与 shared 重新同步（取消弹窗时回退）
const switcherKey = ref(0);
const closeGuardShow = ref(false);
const confirmCloseShow = ref(false);
const handleToggleShared = (value: boolean) => {
  rangeErr.value = '';
  if (value) {
    shared.value = true;
    if (legacyMode.value) {
      // 开启存量共享：转入指定项目模式
      legacyFrom.value = legacyMode.value;
    }
    return;
  }
  if (heldNamespaces.value.length) {
    closeGuardShow.value = true;
  } else {
    confirmCloseShow.value = true;
  }
};
// 弹窗关闭（取消 / 右上角关闭）：重挂载开关，回退其视觉状态
const handleCloseDialog = (key: 'closeGuardShow' | 'confirmCloseShow' | 'removeGuardShow', value: boolean) => {
  if (key === 'closeGuardShow') closeGuardShow.value = value;
  if (key === 'confirmCloseShow') confirmCloseShow.value = value;
  if (key === 'removeGuardShow') removeGuardShow.value = value;
  if (!value) switcherKey.value += 1;
};
const confirmCloseShared = () => {
  shared.value = false;
  switcherKey.value += 1;
  closeGuardShow.value = false;
  confirmCloseShow.value = false;
};
const handleGoRecycle = () => {
  closeGuardShow.value = false;
  confirmCloseShow.value = false;
  removeGuardShow.value = false;
  emits('go-recycle');
};

// 存量转为指定项目：自动带入当前持有命名空间的项目
const handleConvertRange = () => {
  legacyFrom.value = legacyMode.value;
  rangeIds.value = [...new Set(heldNamespaces.value.map(ns => ns.projectID))];
};

// 添加 / 移出可见范围项目
// 选中后自增 selectKey 重建选择器，保持未选择状态（bcs-select 内部会记忆选中项）
const selectKey = ref(0);
const handleAddRange = (projectID: string) => {
  selectKey.value += 1;
  if (!projectID || rangeIds.value.includes(projectID)) return;
  const option = projectList.value.find(project => project.projectID === projectID);
  if (option) projectNameMap.value[projectID] = option.name;
  rangeIds.value.push(projectID);
  rangeErr.value = '';
};
const removeGuardShow = ref(false);
const pendingRemoveId = ref('');
const handleRemoveRange = (id: string) => {
  if (projectNsCount(id)) {
    pendingRemoveId.value = id;
    removeGuardShow.value = true;
    return;
  }
  rangeIds.value = rangeIds.value.filter(item => item !== id);
};
const confirmRemoveRange = () => {
  rangeIds.value = rangeIds.value.filter(item => item !== pendingRemoveId.value);
  removeGuardShow.value = false;
};

// 保存（UpdateCluster：仅传 is_shared 与 sharedRanges）
const handleSave = async () => {
  rangeErr.value = '';
  if (shared.value && !rangeIds.value.length) {
    rangeErr.value = window.i18n.t('cluster.sharedManage.settings.errNoProject');
    return;
  }
  saving.value = true;
  const result = await modifyCluster({
    $clusterId: props.clusterData?.clusterID,
    is_shared: shared.value,
    sharedRanges: shared.value ? { projectIdOrCodes: rangeIds.value } : {},
    updator: user.value?.username,
  }).then(() => true)
    .catch(() => false);
  saving.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: window.i18n.t('generic.msg.success.modify'),
    });
    legacyFrom.value = '';
    emits('refresh');
  }
};

// 取消：恢复基线数据
const handleCancel = () => {
  shared.value = savedShared.value;
  rangeIds.value = [...savedRangeIds.value];
  legacyFrom.value = '';
  rangeErr.value = '';
};

// 初始化聚合 loading（任一请求进行中即展示）
initLoading.value = true;
Promise.all([initProjectList(), initSharedProjectList()]).finally(() => {
  initLoading.value = false;
});
watch(() => props.clusterData?.clusterID, () => initSharedProjectList(), { immediate: true });
</script>

<style lang="postcss" scoped>
:deep(.bcs-small-form) {
  .bk-label-text {
    color: #313238;
  }
}
</style>
