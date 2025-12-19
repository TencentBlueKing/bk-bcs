<template>
  <bcs-form
    ref="formRef"
    :model="baseInfo"
    :rules="rules"
    :label-width="200"
    form-type="vertical"
    class="px-[20px]">
    <div class="font-bold mb-[10px] text-[14px]">{{ $t('serviceMesh.label.info') }}</div>
    <div class="flex mb-[14px]">
      <bk-form-item
        property="name"
        class="mr-[24px]"
        :label="$t('serviceMesh.label.name')"
        error-display-type="normal"
        required>
        <bk-input clearable v-model.trim="baseInfo.name"></bk-input>
      </bk-form-item>
      <bk-form-item
        property="version"
        class="!mt-0"
        :label="$t('serviceMesh.label.version')"
        error-display-type="normal"
        required>
        <bk-radio-group v-model="baseInfo.version" property="version" @change="handelVersionChange">
          <bk-radio-button
            v-for="(item, index) in (configData.istioVersions || [])"
            :key="`${item.version}-${index}`"
            :disabled="isEdit && item.version !== baseInfo.version"
            :value="item.version">
            {{ item.name }}
          </bk-radio-button>
        </bk-radio-group>
      </bk-form-item>
    </div>
    <bk-form-item
      property="revision"
      label="revision"
      :desc="$t('serviceMesh.tips.revisionDesc')">
      <bk-input :disabled="isEdit" clearable v-model.trim="baseInfo.revision"></bk-input>
    </bk-form-item>
    <bk-form-item class="!mt-[14px]" :label="$t('serviceMesh.label.description')">
      <bk-input type="textarea" v-model="baseInfo.description" :maxlength="100"></bk-input>
    </bk-form-item>
    <bcs-divider class="!mt-[24px] !mb-[8px]"></bcs-divider>
    <div class="font-bold mb-[10px] text-[14px]">{{ $t('serviceMesh.label.clusterConfig') }}</div>
    <bk-form-item
      property="primaryClusters"
      :label="$t('serviceMesh.label.primaryClusters')"
      required
      error-display-type="normal"
      :desc="$t('serviceMesh.tips.primaryClusterDesc')">
      <ClusterSelector
        :disabled="isEdit"
        :allowed-version-range="allowedVersionRange"
        :disabled-fun="handleDisabledCluster"
        v-model="baseInfo.primaryClusters" />
    </bk-form-item>
    <ContentSwitcher
      class="mt-[24px] text-[12px]"
      :label="$t('serviceMesh.label.multiCluster')"
      :before-close="handleBeforeClose"
      :disabled="!!subClusters?.length"
      :disabled-desc="$t('serviceMesh.tips.clearClustersTip')"
      v-model="baseInfo.multiClusterEnabled">
      <div class="grid grid-cols-2">
        <bk-form-item
          :label="$t('serviceMesh.label.differentNetwork')"
          :desc="{
            content: $t('serviceMesh.tips.differentNetworkDesc'),
            placement: 'top-start',
          }">
          <bk-radio-group v-model="baseInfo.differentNetwork">
            <bk-radio
              class="text-[12px]"
              :value="true">
              {{ $t('serviceMesh.label.connected') }}
            </bk-radio>
            <bk-radio
              class="text-[12px]"
              :value="false"
              :disabled="true">
              {{ $t('serviceMesh.label.notConnected') }}
            </bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item
          property="clbID"
          error-display-type="normal"
          required
          class="!mt-0"
          :label="$t('serviceMesh.label.clbID')"
          :desc="{
            content: $t('serviceMesh.tips.clbIDDesc'),
            placement: 'top-start',
          }">
          <span
            v-bk-tooltips="{
              content: $t('serviceMesh.tips.clearClustersTip'),
              disabled: !subClusters?.length,
              interactive: false,
            }">
            <bk-input :disabled="!!subClusters?.length" clearable v-model="baseInfo.clbID"></bk-input>
          </span>
        </bk-form-item>
      </div>
      <div class="flex items-center mt-[8px]">
        <span
          class="underline decoration-dashed underline-offset-4 decoration-[#979ba5] cursor-pointer"
          v-bk-tooltips="{
            content: $t('serviceMesh.tips.subClustersDesc'),
            placement: 'top-start',
          }">{{ $t('serviceMesh.label.subClusters') }}</span>
        <ClusterSelector
          ext-cls="!border-0 !rounded-none !shadow-none"
          multiple
          trigger
          :disabled-fun="handleDisabledSubCluster"
          :allowed-version-range="allowedVersionRange"
          :enable-values="Object.keys(originSubClustersMap)"
          v-model="tempClusters"
          @change="handleClusterChange">
          <template #trigger>
            <bk-button
              theme="primary"
              class="ml-[16px] text-[12px]"
              text>
              <i class="bcs-icon bcs-icon-plus-circle"></i>
              {{ $t('generic.button.add') }}
            </bk-button>
          </template>
        </ClusterSelector>
      </div>
      <bcs-table
        class="mt-[8px]"
        custom-header-color="#F5F7FA"
        empty-block-class-name="border-b border-[#DCDEE5]"
        cell-class-name="!h-[36px]"
        header-cell-class-name="!h-[36px]"
        :outer-border="false"
        :data="subClusters">
        <bcs-table-column width="240" :label="$t('generic.label.cluster')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ `${row.clusterName || '--'}(${row.clusterID || '--'})` }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.clusterRegion')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.region || '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column :label="$t('serviceMesh.label.status')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ originSubClustersMap[row.clusterID]
              ? originSubClustersMap[row.clusterID]?.status : '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column width="160" :label="$t('serviceMesh.label.joinTime')" show-overflow-tooltip>
          <template #default="{ row }">
            {{ originSubClustersMap[row.clusterID]
              ? timeFormat(Number(originSubClustersMap[row.clusterID]?.joinTime)) : '--' }}
          </template>
        </bcs-table-column>
        <bcs-table-column :resizable="false" width="50" fixed="right">
          <template #default="{ row }">
            <i
              class="bcs-icon bcs-icon-minus-circle text-[16px] text-[#979ba5] cursor-pointer"
              @click="handleDel(row)">
            </i>
          </template>
        </bcs-table-column>
        <template #empty>
          {{ $t('serviceMesh.label.tableEmpty') }}
        </template>
      </bcs-table>
    </ContentSwitcher>
    <div
      :class="[
        'absolute bottom-0 left-0 w-full bg-white py-[10px] px-[40px] z-[999] border-t border-t-[#e6e6ec]'
      ]">
      <bk-button
        v-if="!isEdit"
        theme="primary"
        @click="nextData">{{ $t('generic.button.next') }}</bk-button>
      <bk-button
        v-else
        theme="primary"
        @click="handleSave">{{ $t('generic.button.save') }}</bk-button>
      <bk-button theme="default" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </bcs-form>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { computed, onBeforeMount, ref, watch } from 'vue';

import ClusterSelector from './cluster-selector.vue';
import ContentSwitcher from './content-switcher.vue';
import useMesh, { IMesh, MeshCluster } from './use-mesh';

import { timeFormat } from '@/common/util';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';

type IParams = Pick<IMesh, 'name' | 'description' | 'version' | 'primaryClusters' | 'remoteClusters' | 'multiClusterEnabled' | 'clbID' | 'differentNetwork' | 'revision'>;

const props = defineProps({
  isEdit: {
    type: Boolean,
    default: false,
  },
  data: {
    type: Object,
    default: () => ({}),
  },
  active: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['next', 'cancel', 'save', 'change']);

const { configData, getClusterList, clusterList } = useMesh();

const baseInfo = ref<IParams>({
  name: '',
  description: '',
  version: '',
  primaryClusters: [],
  remoteClusters: [],
  multiClusterEnabled: false,
  clbID: '',
  differentNetwork: true,
  revision: '',
});
const rules = ref({
  name: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  version: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  primaryClusters: [
    {
      required: true,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
    {
      validator: (value: string[]) => value.length > 0,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
  clbID: [
    {
      validator: (value: string) => !baseInfo.value.multiClusterEnabled || value,
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
    },
  ],
});

// 允许的集群版本
const allowedVersionRange = computed(() => configData.value.istioVersions
  ?.find?.(item => item.version === baseInfo.value.version)?.kubeVersion);

// 切换版本，清空集群
function handelVersionChange() {
  baseInfo.value.primaryClusters = [];
  baseInfo.value.remoteClusters = [];
  tempClusters.value = [];
  subClusters.value = [];
}

// 从集群
const subClusters = ref<MeshCluster[]>([]); // 从集群列表
const tempClusters = ref<string[]>([]); // 下拉框绑定值
const originSubClustersMap = ref<Record<string, MeshCluster>>({}); // 详情从集群
function handleClusterChange(value: string[]) {
  subClusters.value = clusterList.value.filter(item => value.includes(item.clusterID));
}
// 选择从集群时，判断是否已被当作主集群
function handleDisabledSubCluster(val: MeshCluster) {
  return baseInfo.value.primaryClusters.includes(val.clusterID);
}
// 选择主集群时，判断是否已被当从集群
function handleDisabledCluster(val: MeshCluster) {
  return tempClusters.value.includes(val.clusterID);
}
// 点击表格删除按钮
function handleDel(row: MeshCluster) {
  tempClusters.value = tempClusters.value.filter(item => item !== row.clusterID);
  subClusters.value = subClusters.value.filter(item => item.clusterID !== row.clusterID);
}

const { focusOnErrorField } = useFocusOnErrorField();
const formRef = ref();
async function nextData() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    focusOnErrorField();
    return;
  };

  baseInfo.value.remoteClusters = subClusters.value.map(item => ({
    clusterID: item.clusterID,
    clusterName: item.clusterName,
    region: item.region,
  }));
  emits('next', baseInfo.value);
};
function handleCancel() {
  emits('cancel');
}

async function handleSave() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) {
    focusOnErrorField();
    return;
  };
  baseInfo.value.remoteClusters = subClusters.value.map((item) => {
    const cluster = originSubClustersMap.value[item.clusterID];
    if (cluster) {
      return { ...cluster };
    }
    return {
      clusterID: item.clusterID,
      clusterName: item.clusterName,
      region: item.region,
    };
  });
  emits('save', baseInfo.value);
}

function handleBeforeClose() {
  subClusters.value = [];
  tempClusters.value = [];
  baseInfo.value.remoteClusters = [];
  baseInfo.value.clbID = '';
}

watch(() => props.active, () => {
  if (props.isEdit) {
    baseInfo.value = cloneDeep(props.data) as IParams;
    // 初始化从集群
    subClusters.value = props.data.remoteClusters || [];
    tempClusters.value = props.data.remoteClusters?.map?.(item => item.clusterID) || [];
    originSubClustersMap.value = cloneDeep(subClusters.value).reduce((acc, item) => {
      acc[item.clusterID] = item;
      return acc;
    }, {});
  }
}, { immediate: true });

watch(baseInfo, () => {
  if (!props.isEdit) return;
  emits('change');
}, { deep: true });

watch(configData, () => {
  // 编辑态不需要默认值
  if (props.isEdit) return;
  baseInfo.value.version = configData.value.istioVersions?.[0]?.version || '';
});

onBeforeMount(() => {
  getClusterList();
});

</script>
<style lang="postcss" scoped>
:deep(.bk-form-radio-button .bk-radio-button-input:disabled+.bk-radio-button-text) {
  border-left: 1px solid #dcdee5;
}
:deep(.bk-form-radio-button .bk-radio-button-text) {
  font-size: 12px;
}
:deep(.bk-table-empty-text) {
  padding: 0px;
}
:deep(.bk-table th>.cell) {
  height: 36px;
}
</style>
