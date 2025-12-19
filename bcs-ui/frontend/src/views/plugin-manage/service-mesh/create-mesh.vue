<template>
  <bcs-tab
    class="h-full"
    :active.sync="activeTabName"
    type="card-tab"
    v-bkloading="{ isLoading: loading }">
    <bcs-tab-panel
      :name="steps[0].name"
      :disabled="steps[0].disabled">
      <template #label>
        <div class="flex items-center px-[6px]">
          <span class="mr-[4px]">①</span>
          <span>{{ $t('serviceMesh.label.basicInfo') }}</span>
        </div>
      </template>
      <BasicInfo
        @next="nextStep"
        @cancel="handleCancel" />
    </bcs-tab-panel>
    <bcs-tab-panel
      :name="steps[1].name"
      :disabled="steps[1].disabled">
      <template #label>
        <div class="flex items-center px-[6px]">
          <span class="mr-[4px]">②</span>
          <span>{{ $t('serviceMesh.label.network') }}</span>
        </div>
      </template>
      <MeshConfig
        @pre="preStep"
        @next="nextStep"
        @cancel="handleCancel" />
    </bcs-tab-panel>
    <bcs-tab-panel
      :name="steps[2].name"
      :disabled="steps[2].disabled">
      <template #label>
        <div class="flex items-center px-[6px]">
          <span class="mr-[4px]">③</span>
          <span>{{ $t('serviceMesh.label.master') }}</span>
        </div>
      </template>
      <Master
        @pre="preStep"
        @submit="handleCreate"
        @cancel="handleCancel" />
    </bcs-tab-panel>
  </bcs-tab>
</template>
<script setup lang="ts">
import { mergeWith } from 'lodash';
import { onMounted, ref } from 'vue';

import BasicInfo from './basic-info.vue';
import Master from './master.vue';
import MeshConfig from './mesh-config.vue';
import useMesh, { IMesh } from './use-mesh';

import { meshCreate  } from '@/api/modules/mesh-manager';
import $bkMessage from '@/common/bkmagic';
import { useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';

const emits = defineEmits(['cancel']);

const activeTabName = ref('basicInfo');
const steps = ref([
  { name: 'basicInfo', disabled: false },
  { name: 'network',  disabled: true },
  { name: 'master',  disabled: true },
]);

const { handleGetConfig, loading } = useMesh();

const { projectID, projectCode } = useProject();
const formData = ref<Partial<IMesh>>({
  projectID: projectID.value,
  projectCode: projectCode.value,
});
function nextStep(data = {}) {
  formData.value = mergeWith({}, formData.value, data, (objValue, srcValue) => {
    if (Array.isArray(objValue)) {
      return srcValue;
    }
  });
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index + 1 < steps.value.length) {
    steps.value[index + 1].disabled = false;
    activeTabName.value = steps.value[index + 1]?.name;
  }
}
function preStep() {
  const index = steps.value.findIndex(step => activeTabName.value === step.name);
  if (index > -1 && index - 1 >= 0) {
    activeTabName.value = steps.value[index - 1]?.name;
  }
};

async function handleCreate(data = {}) {
  nextStep(data);

  const result = await meshCreate(formData.value)
    .then(() => true)
    .catch(() => false);
  if (!result) {
    return;
  }

  handleCancel();
  $bkMessage({
    theme: 'success',
    message: $i18n.t('generic.msg.success.create'),
  });
}

function handleCancel() {
  emits('cancel');
}

onMounted(async () => {
  await handleGetConfig();
});

</script>
