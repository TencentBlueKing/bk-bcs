<template>
  <bcs-form
    ref="formRef"
    :model="baseInfo"
    :rules="rules"
    :label-width="200"
    form-type="vertical"
    class="px-[20px]">
    <div class="font-bold mb-[10px] text-[14px]">{{ $t('serviceMesh.label.info') }}</div>
    <div class="flex mb-[18px]">
      <bk-form-item
        property="name"
        class="mr-[24px]"
        :label="$t('serviceMesh.label.name')"
        error-display-type="normal"
        required>
        <bk-input v-model.trim="baseInfo.name"></bk-input>
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
            :value="item.version">
            {{ item.name }}
          </bk-radio-button>
        </bk-radio-group>
      </bk-form-item>
    </div>
    <bk-form-item :label="$t('serviceMesh.label.description')">
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
        :allowed-version-range="allowedVersionRange"
        v-model="baseInfo.primaryClusters" />
    </bk-form-item>
    <!-- <ContentSwitcher
      class="mt-[26px] text-[12px]"
      :label="'多集群'"
      v-model="clustersOpen">
      <div class="flex justify-between">
        <bk-form-item label="集群网络连通性" desc="通过云联网打通容器网络，或者自研云内要求在同一个 vpc">
          <bk-radio-group v-model="baseInfo.name">
            <bk-radio :value="'value1'">已打通</bk-radio>
            <bk-radio :value="'value2'" :disabled="true">未打通（暂不支持）</bk-radio>
          </bk-radio-group>
        </bk-form-item>
        <bk-form-item label="东西向网关 CLB" desc="用于从集群连接主集群控制面，填写一个 clb id">
          <bk-input v-model="baseInfo.name"></bk-input>
        </bk-form-item>
      </div>
      <div class="flex items-center">
        <span
          class="underline decoration-dashed underline-offset-4 decoration-[#979ba5] cursor-pointer"
          v-bk-tooltips="{
            content: '构建多集群，用于加入主机群控制面',
            placement: 'top',
          }">从集群</span>
        <ClusterSelector
          ext-cls="!border-0 !rounded-none !shadow-none"
          multiple
          trigger
          v-model="baseInfo.remoteClusters">
          <template #trigger>
            <bk-button
              theme="primary"
              class="ml-[16px] text-[12px]"
              text>
              <i class="bcs-icon bcs-icon-plus-circle"></i>
              添加
            </bk-button>
          </template>
        </ClusterSelector>
      </div>
      <bcs-table
        class="mt-[8px]"
        empty-text="暂无数据，请先添加集群"
        :outer-border="false"
        :data="baseInfo.remoteClusters">
        <bcs-table-column label="集群" prop="clusterName"></bcs-table-column>
        <bcs-table-column label="集群地域" prop="region"></bcs-table-column>
        <bcs-table-column label="加入时间" prop="createTime"></bcs-table-column>
      </bcs-table>
    </ContentSwitcher> -->
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
import { computed, ref, watch } from 'vue';

import ClusterSelector from './cluster-selector.vue';
import useMesh, { IMesh } from './use-mesh';

// import ContentSwitcher from './content-switcher.vue';
import $i18n from '@/i18n/i18n-setup';

type IParams = Pick<IMesh, 'name' | 'description' | 'version' | 'primaryClusters'>;

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

const { configData } = useMesh();

const baseInfo = ref<IParams>({
  name: '',
  description: '',
  version: '',
  primaryClusters: [],
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
});

const allowedVersionRange = computed(() => configData.value.istioVersions
  ?.find(item => item.version === baseInfo.value.version)?.kubeVersion);

function handelVersionChange() {
  baseInfo.value.primaryClusters = [];
}

const formRef = ref();
async function nextData() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) return;

  emits('next', baseInfo.value);
};
function handleCancel() {
  emits('cancel');
}

async function handleSave() {
  const result = await formRef.value.validate().catch(() => false);
  if (!result) return;
  emits('save', baseInfo.value);
}

watch(() => props.active, () => {
  if (props.isEdit) {
    baseInfo.value = cloneDeep(props.data) as IParams;
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

</script>
