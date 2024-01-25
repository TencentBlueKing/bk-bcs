<template>
  <bk-sideslider :is-show="props.show" :title="t('版本对比')" :width="1200" @closed="handleClose">
    <div class="diff-content-area">
      <div class="header-wrapper">
        <div class="base-title">
          <bk-select
            class="version-select"
            :model-value="selectedVersion"
            :loading="versionListLoading"
            :clearable="false"
            :filterable="true"
            @change="handleSelectVersion"
            :show-on-init="true">
            <bk-option v-for="version in versionList" :key="version.id" :label="version.spec.name" :value="version.id">
            </bk-option>
          </bk-select>
        </div>
        <div class="current-title">{{ props.crtVersion.spec.name }}</div>
      </div>
      <div class="diff-code-content">
        <DiffText
          :current="crtVersion.spec.content"
          :current-language="props.type"
          :base="baseContent"
          :base-language="props.type"/>
      </div>
    </div>
    <div class="actions-btn">
      <bk-button @click="handleClose">{{ t('关闭') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { IScriptVersion, IScriptVersionListItem } from '../../../../../types/script';
import { getScriptVersionList } from '../../../../api/script';
import DiffText from '../../../../components/diff/text.vue';

const { t } = useI18n();

const props = defineProps<{
  spaceId: string;
  scriptId: number;
  type: string;
  show: boolean;
  crtVersion: IScriptVersion;
}>();

const emits = defineEmits(['update:show']);

const selectedVersion = ref();
const versionList = ref<IScriptVersion[]>([]);
const versionListLoading = ref(false);
const baseContent = ref('');

onMounted(() => {
  getVersionList();
});

// 获取版本列表
const getVersionList = async () => {
  versionListLoading.value = true;
  const params = {
    start: 0,
    all: true,
  };
  const res = await getScriptVersionList(props.spaceId, props.scriptId, params);
  versionList.value = (res.details as IScriptVersionListItem[])
    .filter(item => item.hook_revision.id !== props.crtVersion.id)
    .map(item => item.hook_revision);
  versionListLoading.value = false;
};

const handleSelectVersion = (id: number) => {
  const version = versionList.value.find(item => item.id === id);
  if (version) {
    baseContent.value = version.spec.content;
  }
};

const handleClose = () => {
  selectedVersion.value = undefined;
  versionList.value = [];
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.diff-content-area {
  height: calc(100vh - 100px);
}
.header-wrapper {
  display: flex;
  align-items: center;
  background: #313238;
  .base-title,
  .current-title {
    padding: 5px 24px;
    font-size: 12px;
    color: #b6b6b6;
  }
  .base-title {
    width: 586px;
    border-right: 1px solid #1d1d1d;
  }
  .current-title {
    flex: 1;
  }
  .version-select {
    width: 340px;
    :deep(.bk-input) {
      color: #b1b1b1;
      border-color: #63656e;
      background: transparent;
      input {
        color: #b1b1b1;
        background: transparent;
      }
    }
  }
}
.diff-code-content {
  height: calc(100% - 42px);
}
.actions-btn {
  padding: 8px 24px;
  background: #fafbfd;
  box-shadow: 0 -1px 0 0 #dcdee5;
  .bk-button {
    min-width: 88px;
  }
}
</style>
