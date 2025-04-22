<template>
  <bcs-sideslider :is-show.sync="isShow" :width="1000" quick-close @hidden="hideSideslider">
    <template #header>
      <div class="flex items-center">
        <span>{{ rollback ? $t('updateRecord.label.rollbackVersion') :$t('updateRecord.label.versionDiff') }}</span>
        <bcs-divider direction="vertical"></bcs-divider>
        <span class="text-[12px]">{{ rollbackReversion }}</span>
      </div>
    </template>
    <template #content>
      <div class="p-[24px]" v-bkloading="{ isLoading: detailLoading || versionListLoading }">
        <div
          class="flex items-center justify-between h-[42px] bg-[#2E2E2E] shadow pl-[24px] pr-[16px]">
          <!-- 回滚时当前版本不允许选择 -->
          <span v-if="rollback">
            <span class="text-[#B6B6B6] text-[12px]">{{ onlineVersion }}</span>
            <bcs-tag
              theme="success"
              class="bg-[#E4FAF0]">
              {{ $t('updateRecord.label.onlineVersion') }}
            </bcs-tag>
          </span>
          <!-- 当前版本 -->
          <bcs-select
            class="flex-1 !bg-[#2E2E2E] version-select"
            :clearable="false"
            :loading="versionListLoading"
            v-model="currentVersion"
            v-else
            @change="getCurrentVersionData">
            <bcs-option
              v-for="item in tableData"
              :key="item.revision"
              :id="item.revision"
              :name="onlineVersion === item.revision
                ? `${item.revision}(${$t('updateRecord.label.onlineVersion')})`
                : item.revision">
            </bcs-option>
          </bcs-select>
          <!-- 对比/回滚版本 -->
          <div class="w-[50%] ml-[16px] flex items-center">
            <bcs-select
              class="flex-1 !bg-[#2E2E2E] version-select"
              :clearable="false"
              :loading="versionListLoading"
              v-model="rollbackReversion"
              @change="getRollbackVersionData">
              <bcs-option
                v-for="item in tableData"
                :key="item.revision"
                :id="item.revision"
                :name="onlineVersion === item.revision
                  ? `${item.revision}(${$t('updateRecord.label.onlineVersion')})`
                  : item.revision">
              </bcs-option>
            </bcs-select>
            <bcs-button
              theme="primary"
              class="ml-[8px]"
              :disabled="tableData?.length === 1 || !rollbackReversion"
              v-if="rollback"
              @click="handleRollback">{{ $t('updateRecord.button.rolloutThisVersion') }}</bcs-button>
          </div>
        </div>
        <div class="h-[calc(100vh-150px)]">
          <CodeEditor
            readonly
            diff-editor
            full-screen
            :original="currentVersionData"
            :value="rollbackVersionData" />
        </div>
      </div>
    </template>
  </bcs-sideslider>
</template>
<script setup lang="ts">
import { onBeforeMount, ref, watch } from 'vue';

import useRecords, { IRevisionData } from './use-records';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import $i18n from '@/i18n/i18n-setup';

const {
  revisionDetail,
  revisionGameDetail,
  workloadHistory,
  gameWorkloadHistory,
  rollbackWorkload,
  rollbackGameWorkload,
} = useRecords();

const props = defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  revision: {
    type: [String, Number],
    default: '',
    required: true,
  },
  crd: {
    type: String,
    default: '',
  },
  namespace: {
    type: String,
    default: '',
    required: true,
  },
  name: {
    type: String,
    default: '',
    required: true,
  },
  category: {
    type: String,
    default: '',
    required: true,
  },
  clusterId: {
    type: String,
    default: '',
    required: true,
  },
  rollback: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['hidden', 'rollback-success']);

watch(() => props.value, () => {
  isShow.value = props.value;
});

watch(() => props.revision, () => {
  rollbackReversion.value = props.revision;
});

watch(() => props, () => {
  getRollbackVersionData();
  handleGetHistory();
}, { deep: true });

const isShow = ref(false);

// 版本详情
const detailLoading = ref(false);
const getRevisionDetail = async ($revision: string|number) => {
  if (!$revision || !isShow.value) return '';
  detailLoading.value = true;
  let data = { rollout_revision: '' };
  if (props.category === 'custom_objects') {
    data = await revisionGameDetail({
      $crd: props.crd,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      $revision,
      namespace: props.namespace,
    });
  } else {
    data = await revisionDetail({
      $namespaceId: props.namespace,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      $revision,
    });
  }

  detailLoading.value = false;
  return data?.rollout_revision;
};

// 当前版本
const currentVersion = ref('');
const currentVersionData = ref('');
const getCurrentVersionData = async () => {
  currentVersionData.value = await getRevisionDetail(currentVersion.value);
};
// 对比的版本（回滚）
const rollbackReversion = ref(props.revision);
const rollbackVersionData = ref('');
const getRollbackVersionData = async () => {
  rollbackVersionData.value = await getRevisionDetail(rollbackReversion.value);
};

// 版本列表
const onlineVersion = ref('');
const tableData = ref<IRevisionData[]>([]);
const versionListLoading = ref(false);
const handleGetHistory = async () => {
  if (!props.name || !isShow.value) return;
  versionListLoading.value = true;
  if (props.category === 'custom_objects') {
    tableData.value = await gameWorkloadHistory({
      $crd: props.crd,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
      namespace: props.namespace,
    });
  } else {
    tableData.value = await workloadHistory({
      $namespaceId: props.namespace,
      $clusterId: props.clusterId,
      $name: props.name,
      $category: props.category,
    });
  }

  onlineVersion.value = tableData.value[0]?.revision;
  // 初始化左侧diff数据
  currentVersion.value = onlineVersion.value;
  getCurrentVersionData();
  versionListLoading.value = false;
};

// 回滚
const handleRollback = () => {
  $bkInfo({
    type: 'warning',
    clsName: 'custom-info-confirm',
    title: $i18n.t('updateRecord.confirmRollout.title'),
    subTitle: $i18n.t('updateRecord.confirmRollout.subTitle', [rollbackReversion.value]),
    defaultInfo: true,
    confirmFn: async () => {
      let result = false;
      if (props.category === 'custom_objects') {
        result = await rollbackGameWorkload({
          $crd: props.crd,
          $clusterId: props.clusterId,
          $name: props.name,
          $category: props.category,
          $revision: rollbackReversion.value,
          namespace: props.namespace,
        });
      } else {
        result = await rollbackWorkload({
          $namespaceId: props.namespace,
          $clusterId: props.clusterId,
          $name: props.name,
          $category: props.category,
          $revision: rollbackReversion.value,
        });
      }

      if (result) {
        $bkMessage({
          theme: 'success',
          message: $i18n.t('generic.msg.success.ok'),
        });
        emits('rollback-success');
      }
    },
  });
};

// 隐藏
const hideSideslider = () => {
  rollbackReversion.value = props.revision;
  emits('hidden');
};

onBeforeMount(() => {
  getRollbackVersionData();
  handleGetHistory();
});

</script>

<style lang="postcss" scoped>

>>> .version-select {
  border: 1px solid #63656E;
  .bk-select-name {
    color: #B1B1B1;
  }
}
</style>
