<template>
  <!-- root dom loading有点问题 -->
  <div>
    <div
      class="flex border border-[#DCDEE5] border-solid bg-[#fff]"
      v-bkloading="{ isLoading: loading }">
      <div class="w-[280px] log-border-r relative">
        <Row class="log-border-b px-[16px] h-[42px] text-[12px] bg-[#FAFBFD]">
          <template #left>
            <div class="text-[#313238]">{{ $t('plugin.tools.ruleName') }}</div>
          </template>
          <template #right>
            <bcs-button text class="text-[12px]" @click="handleShowList">
              {{ $t('logCollector.button.expandList') }}
              <i class="bcs-icon bcs-icon-angle-double-right"></i>
            </bcs-button>
          </template>
        </Row>
        <div
          :class="[
            'log-border-b',
            'flex items-center justify-between cursor-pointer px-[16px]',
            'text-[12px] h-[42px] text-[#3A84FF] hover:bg-[#E1ECFF] bg-[#E1ECFF]',
          ]"
          v-if="fromOldRule || !activeID"
          @click="handleChangeID()">
          <span>{{ fromOldRule ? $t('logCollector.action.newRule') : $t('plugin.tools.create') }}</span>
        </div>
        <!-- 规则列表 -->
        <div class="overflow-auto" ref="listRef">
          <div
            v-for="row in data"
            :key="row.id"
            :class="[
              'log-border-b',
              'flex items-center justify-between cursor-pointer px-[16px]',
              'text-[12px] h-[42px] text-[#3A84FF] hover:bg-[#E1ECFF]',
              { 'bg-[#E1ECFF]': activeID === row.id && !fromOldRule }
            ]"
            @click="handleChangeID(row.id)">
            <div class="flex items-center">
              <!-- 是否旧规则 -->
              <span
                class="inline-flex items-center justify-center w-[24px] relative
                  h-[24px] bg-[#F0F1F5] rounded-sm text-[#979BA5] mr-[8px]"
                v-bk-tooltips="{
                  content: $t('logCollector.tips.oldRuleHasNewRule', [
                    row.new_rule_name,
                    moment(row.new_rule_created_at).add(30, 'days').format('YYYY-MM-DD')
                  ]),
                  disabled: !row.new_rule_id
                }"
                v-if="row.old">
                {{ $t('logCollector.msg.oldRuleFlag') }}
                <!-- 是否转换过 -->
                <i
                  class="absolute bottom-[0px] right-[-6px] text-[#87cfab] bcs-icon bcs-icon-check-circle-shape"
                  v-if="row.new_rule_id">
                </i>
              </span>
              <span class="bcs-ellipsis flex-1" v-bk-overflow-tips>{{ row.name }}</span>
            </div>
            <StatusIcon :status-color-map="statusColorMap" :status="row.status" hide-text />
          </div>
        </div>
        <!-- 分页 -->
        <!-- <bcs-pagination
          class="flex items-center justify-center absolute bottom-0 w-full h-[42px] log-border-t"
          small
          :current="pagination.current"
          :count="data.length"
          :limit="pagination.limit"
          :show-limit="false"
          @change="pageChange" /> -->
      </div>
      <div class="flex-1 text-[12px] overflow-hidden">
        <Row class="log-border-b h-[42px] px-[16px] bg-[#FAFBFD]">
          <template #left>
            <bcs-tag :theme="statusTheme" class="mr-[8px]" v-if="activeID">
              {{ statusTextMap[activeRow.status || ''] }}
            </bcs-tag>
            <span class="font-bold">
              {{ activeID
                ? activeRow.name
                : (fromOldRule ? $t('logCollector.action.newRule') : $t('plugin.tools.create')) }}
            </span>
          </template>
          <template #right>
            <!-- 编辑规则 -->
            <template v-if="activeID && !editable">
              <bcs-button
                text
                class="text-[12px]"
                v-if="activeRow
                  && !activeRow.old
                  && !['PENDING', 'RUNNING', 'TERMINATED', 'DELETED'].includes(activeRow.status || '')"
                @click="setStatusOfEditRule">
                <span class="flex items-center">
                  <span class="text-[14px] relative top-[-2px]"><i class="bk-icon icon-edit-line"></i></span>
                  <span class="ml-[5px]">{{ $t('generic.button.edit') }}</span>
                </span>
              </bcs-button>
              <bcs-button
                text
                class="text-[12px]"
                v-else-if="activeRow && activeRow.old"
                @click="setStatusOfCreateRuleFromOld">
                <span class="flex items-center">
                  <span class="text-[14px] relative top-[-1px]">
                    <i class="bcs-icon bcs-icon-arrows-up-circle"></i>
                  </span>
                  <span class="ml-[5px]">{{ $t('logCollector.action.newRule') }}</span>
                </span>
              </bcs-button>
              <bcs-button
                text
                class="text-[12px] ml-[16px]"
                v-if="['SUCCESS', 'TERMINATED'].includes(activeRow.status || '') && !activeRow.old"
                @click="handleToggleRule">
                <span class="flex items-center">
                  <span class="text-[14px] relative top-[-1px]"><i class="bcs-icon bcs-icon-switch"></i></span>
                  <span class="ml-[5px]">
                    {{ activeRow.status === 'TERMINATED'
                      ? $t('logCollector.action.enable') : $t('logCollector.action.stop') }}
                  </span>
                </span>
              </bcs-button>
              <bcs-button text class="text-[12px] ml-[16px]" @click="handleDeleteRule">
                <span class="flex items-center">
                  <span class="text-[14px] relative top-[-2px]"><i class="bk-icon icon-close3-shape"></i></span>
                  <span class="ml-[5px]">{{ $t('generic.button.delete') }}</span>
                </span>
              </bcs-button>
            </template>
            <!-- 新建规则 -->
            <template v-else>
              <bcs-button
                theme="primary"
                class="w-[72px]"
                @click="handleSaveData">{{ $t('generic.button.save') }}</bcs-button>
              <bcs-button @click="handleCancel">{{ $t('generic.button.cancel') }}</bcs-button>
            </template>
          </template>
        </Row>
        <bcs-alert
          type="warning"
          :title="$t('logCollector.msg.oldRuleToNewRule', [date])"
          v-if="fromOldRule" />
        <div class="py-[16px] pr-[16px] overflow-auto" ref="formWrapperRef">
          <LogConllectorDetail :data="logData" v-if="!editable && activeID" />
          <LogForm
            :cluster-id="clusterId"
            :data="logData"
            :from-old-rule="fromOldRule"
            ref="logForm"
            :key="activeID"
            @refresh="handleRefreshList"
            v-else />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import moment from 'moment';
import { computed, onBeforeMount, PropType, ref, watch } from 'vue';

import LogConllectorDetail from './log-conllector-detail.vue';
import LogForm from './log-form.vue';
import useLog, { IRuleData } from './use-log';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import Row from '@/components/layout/Row.vue';
import StatusIcon from '@/components/status-icon';
import $i18n from '@/i18n/i18n-setup';
// import $router from '@/router';

const props = defineProps({
  // 列表数据
  data: {
    type: Array as PropType<Array<IRuleData>>,
    default: () => [],
  },
  statusColorMap: {
    type: Object,
    default: () => ({}),
  },
  statusTextMap: {
    type: Object,
    default: () => ({}),
  },
  clusterId: {
    type: String,
    required: true,
  },
  // 当前规则ID
  id: {
    type: String,
  },
  // 是否进入编辑态
  edit: {
    type: Boolean,
    default: false,
  },
});

watch(() => props.id, () => {
  handleChangeID(props.id || '');
});

const emits = defineEmits(['show-list', 'delete', 'refresh', 'toggle-rule', 'update-create-id']);

// const pagination = ref({
//   current: 1,
//   limit: 10,
// });
// const curPageData = computed(() => {
//   const { limit, current } = pagination.value;
//   return props.data.slice(limit * (current - 1), limit * current);
// });
const date = moment(new Date()).add(30, 'days')
  .format('YYYY-MM-DD');
const { logCollectorDetail, createLogCollectorRule, modifyLogCollectorRule } = useLog();
const editable = ref(props.edit);
const fromOldRule = ref(false); // 是否从旧规则中生成
// 详情
const loading = ref(false);
const logData = ref<IRuleData>();
const handleGetDetail = async () => {
  if (!activeID.value) return;
  loading.value = true;
  logData.value = await logCollectorDetail({ $clusterId: props.clusterId, $ID: activeID.value });
  loading.value = false;
  // 旧规则不能编辑，只能生成新的
  if (editable.value && logData.value?.old) {
    setNewRuleStatus();
  }
};

// 显示列表
const handleShowList = () => {
  if (!editable.value && activeID.value) {
    activeID.value = '';
    emits('show-list');
    return;
  }
  // 编辑态返回列表确认
  $bkInfo({
    title: $i18n.t('generic.msg.info.exitTips.text'),
    subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
    clsName: 'custom-info-confirm default-info',
    okText: $i18n.t('generic.button.exit'),
    cancelText: $i18n.t('generic.button.cancel'),
    confirmFn() {
      activeID.value = '';
      emits('show-list');
    },
    cancelFn() {},
  });
};

// 分页
// const pageChange = (page: number) => {
//   pagination.value.current = page;
// };

const activeID = ref(props.id || '');
const activeRow = computed<IRuleData>(() => (props.data.find(item => item.id === activeID.value) || {}) as IRuleData);
const statusTheme = computed(() => {
  if (activeRow.value?.status === 'FAILED') return 'danger';

  if (activeRow.value?.status === 'SUCCESS') return 'success';

  return '';
});

// watch(activeID, () => {
//   if ($router.currentRoute?.query?.id === activeID.value) return;
//   $router.replace({
//     name: 'logCrdcontroller',
//     query: {
//       id: activeID.value,
//     },
//   });
// }, { immediate: true });

const handleChangeID = (id = '') => {
  if (!id) {
    setStatusOfCreateRule();
    return;
  }

  if (activeID.value === id) return;
  fromOldRule.value = false;
  editable.value = false;
  activeID.value = id;
  handleGetDetail();
};

// 编辑规则
const setStatusOfEditRule = () => {
  editable.value = true;
};

// 删除规则
const handleDeleteRule = () => {
  emits('delete', logData.value);
};

// 创建规则
const setStatusOfCreateRule = () => {
  fromOldRule.value = false;
  editable.value = false;
  activeID.value = '';
  logData.value = undefined;
};

// 启用 & 停用
const handleToggleRule = () => {
  emits('toggle-rule', activeRow.value);
};

// 生成新规则
const setNewRuleStatus = () => {
  if (!logData.value) return;
  logData.value.name = `${logData.value?.name.replace(/-/g, '_')}_new`;
  activeID.value = '';
  fromOldRule.value = true;
};
const setStatusOfCreateRuleFromOld = async () => {
  if (logData.value?.new_rule_id) {
    $bkInfo({
      type: 'warning',
      title: $i18n.t('logCollector.title.exitNewRule', [logData.value.name]),
      clsName: 'custom-info-confirm default-info',
      okText: $i18n.t('logCollector.button.continueCreateNewRule'),
      cancelText: $i18n.t('generic.button.cancel'),
      async confirmFn() {
        setNewRuleStatus();
      },
    });
  } else {
    setNewRuleStatus();
  }
};

// 取消
const handleCancel = () => {
  if (activeID.value) {
    editable.value = false;
  } else {
    handleShowList();
  }
};

// 保存
const logForm = ref();
const handleCreateRule = async (data: IRuleData) => {
  loading.value = true;
  const result = await createLogCollectorRule({
    ...data,
    from_rule: data.id || '',
    $clusterId: props.clusterId,
  });
  loading.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('logCollector.msg.success.create'),
    });
    logData.value = undefined;
    emits('update-create-id', result);
    emits('refresh');
    emits('show-list');
  }
};
const handleModifyRule = async (data: IRuleData) => {
  loading.value = true;
  const result = await modifyLogCollectorRule({
    ...data,
    $clusterId: props.clusterId,
    $ID: data.id || '',
  });
  loading.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('logCollector.msg.success.update'),
    });
    logData.value = undefined;
    editable.value = false;
    emits('refresh');
    emits('show-list');
  }
};
const handleSaveData = async () => {
  const data = await logForm.value.handleGetFormData();
  if (!data) return;

  if (!data.id || fromOldRule.value) {
    handleCreateRule(data);
  } else {
    handleModifyRule(data);
  }
};

// 刷新列表
const handleRefreshList = () => {
  emits('refresh');
};

// 更新滚动
const formWrapperRef = ref();
watch([editable, activeID], () => {
  if (formWrapperRef.value) {
    formWrapperRef.value.scrollTop = 0;
  }
});

defineExpose({
  setStatusOfCreateRule,
  handleGetDetail,
});

onBeforeMount(() => {
  handleGetDetail();
});
</script>
<style lang="postcss" scoped>
.log-border-r {
  border-right: 1px solid #DCDEE5;
}
.log-border-b {
  border-bottom: 1px solid #DCDEE5;
}
.log-border-t {
  border-top: 1px solid #DCDEE5;
}
</style>
