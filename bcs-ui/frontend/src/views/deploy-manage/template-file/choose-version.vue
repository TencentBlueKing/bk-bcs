<template>
  <bcs-dialog
    :value="value"
    :title="$t('templateFile.title.editNewVersion')"
    width="480"
    header-position="left"
    @cancel="cancel">
    <bk-form form-type="vertical" :label-width="300">
      <bk-form-item required :label="$t('templateFile.label.rebaseVersion')">
        <bcs-select :loading="loading" :clearable="false" searchable v-model="version">
          <bcs-option v-for="item in versionList" :key="item.version" :id="item.version" :name="item.version" />
        </bcs-select>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div>
        <bcs-button theme="primary" :disabled="!version" @click="confirm">
          {{ $t('generic.button.create') }}
        </bcs-button>
        <bcs-button @click="cancel">{{ $t('generic.button.cancel') }}</bcs-button>
      </div>
    </template>
  </bcs-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import { ITemplateVersionItem } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';
import $router from '@/router';

const props = defineProps({
  // 弹窗是否显示
  value: {
    type: Boolean,
    default: false,
  },
  // 模板文件ID
  id: {
    type: String,
    default: '',
  },
  // 默认版本数据
  defaultVersion: String,
  mode: String, // 编辑时默认模式
});

const emits = defineEmits(['cancel']);

const loading = ref(false);

// 版本列表
const version = ref(props.defaultVersion);
const versionList = ref<ITemplateVersionItem[]>([]);
const listTemplateVersion = async () => {
  if (!props.id || !props.value) return;
  loading.value = true;
  versionList.value = await TemplateSetService.ListTemplateVersion({
    $templateID: props.id,
  }).catch(() => []);
  loading.value = false;
};


// 创建新版本
const confirm = () => {
  $router.push({
    name: 'addTemplateFileVersion',
    params: {
      id: props.id,
    },
    query: {
      mode: props.mode || $router.currentRoute?.query?.mode,
      versionID: versionList.value.find(item => item.version === version.value)?.id as string,
    },
  });
};

// 取消
const cancel = () => {
  emits('cancel');
};

watch(() => props.value, async () => {
  if (props.value) {
    await listTemplateVersion();
  }
});

watch(() => props.defaultVersion, () => {
  version.value = props.defaultVersion;
});
</script>
