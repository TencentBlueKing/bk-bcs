<template>
  <div v-bkloading="{ isLoading: loading }">
    <bcs-tab
      class="h-full"
      :active.sync="activeTabName"
      :before-toggle="beforeToggle"
      type="card-tab">
      <bcs-tab-panel
        name="basicInfo"
        :label="$t('serviceMesh.label.basicInfo')">
        <BasicInfo
          v-if="isEdit"
          :is-edit="isEdit"
          :data="data"
          :active="activeTabName"
          @save="handleUpdate"
          @change="handleChange"
          @cancel="handleCancel" />
        <Basic
          v-else
          class="px-[20px]"
          :data="data" />
      </bcs-tab-panel>
      <bcs-tab-panel
        name="network"
        :label="$t('serviceMesh.label.network')">
        <MeshConfig
          v-if="isEdit"
          :is-edit="isEdit"
          :data="data"
          :active="activeTabName"
          @save="handleUpdate"
          @change="handleChange"
          @cancel="handleCancel" />
        <FeatureConfig
          v-else
          :data="data"
          class="px-[20px]" />
      </bcs-tab-panel>
      <bcs-tab-panel
        name="master"
        :label="$t('serviceMesh.label.master')">
        <Master
          v-if="isEdit"
          :is-edit="isEdit"
          :data="data"
          :active="activeTabName"
          @save="handleUpdate"
          @change="handleChange"
          @cancel="handleCancel" />
        <HighAvailability
          v-else
          :data="data"
          class="px-[20px]" />
      </bcs-tab-panel>
    </bcs-tab>
    <div
      v-if="!isEdit"
      :class="[
        'absolute bottom-0 left-0 w-full bg-white py-[10px] px-[40px] z-[999] border-t border-t-[#e6e6ec]'
      ]">
      <bcs-button
        v-authority="{
          clickable: web_annotations.perms[meshId]?.['MeshManager.UpdateIstio'],
          disablePerms: true,
          originClick: true,
        }"
        theme="primary"
        @click.stop="handleEdit">
        {{ $t('generic.button.edit') }}</bcs-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { cloneDeep } from 'lodash';
import { onMounted, ref } from 'vue';

import BasicInfo from './basic-info.vue';
import Basic from './detail/basic.vue';
import FeatureConfig from './detail/feature-config.vue';
import HighAvailability from './detail/high-availability.vue';
import Master from './master.vue';
import MeshConfig from './mesh-config.vue';
import useMesh, { IMesh } from './use-mesh';

import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  meshId: {
    type: String,
    default: '',
  },
  projectCode: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['success']);

const isEdit = ref<boolean>(false);

const activeTabName = ref('basicInfo');
const loading = ref(false);

const unUseProperties = ['status', 'statusMessage', 'networkID', 'chartVersion', 'createTime', 'createBy', 'updateTime', 'updateBy'];
const formData = ref<IMesh>();
async function handleUpdate(val: IMesh) {
  formData.value = cloneDeep(val);
  unUseProperties.forEach((key) => {
    formData.value && delete formData.value[key];
  });

  const result = await handleUpdateMesh(formData.value);
  if (result) {
    $bkMessage({
      theme: 'success',
      message: $i18n.t('generic.msg.success.ok'),
    });
    isEdit.value = false;
    isChange.value = false;
    await getMeshData();
    emits('success', data.value);
  }
}

// 修改
const isChange = ref(false);
function handleChange() {
  isChange.value = true;
}

function handleCancel() {
  isEdit.value = false;
  // 重置
  formData.value = cloneDeep(data.value);
}

// tab 切换前的钩子
function beforeToggle(tab) {
  if (!isEdit.value || !isChange.value) return true;
  $bkInfo({
    title: $i18n.t('generic.msg.info.exitTips.text'),
    subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
    defaultInfo: true,
    okText: $i18n.t('generic.button.exit'),
    confirmFn() {
      activeTabName.value = tab;
      isChange.value = false;
      isEdit.value = false;
    },
  });
}

const data = ref<IMesh>();
const {
  web_annotations,
  handleGetMeshDetail,
  handleUpdateMesh,
  handleGetConfig,
} = useMesh();
async function getMeshData() {
  loading.value = true;
  data.value = await handleGetMeshDetail({
    meshID: props.meshId,
    projectCode: props.projectCode,
  });
  if (data.value) {
    formData.value = cloneDeep(data.value);
  }
  loading.value = false;
}

// 编辑
function handleEdit() {
  if (!web_annotations.value.perms[props.meshId]?.['MeshManager.UpdateIstio']) {
    // 没有权限时接口会触发权限弹窗，类似用户点击申请权限
    formData.value && handleUpdate(formData.value);
    return;
  }
  isEdit.value = true;
}

onMounted(async () => {
  await getMeshData();
  await handleGetConfig(); // 编辑态需要versions
});

defineExpose({
  isEdit,
  isChange,
});

</script>
