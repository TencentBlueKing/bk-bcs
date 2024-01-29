<template>
  <bk-sideslider :is-show="props.show" :title="t('版本对比')" :width="1200" @closed="handleClose">
    <div class="diff-content-area">
      <diff :diff="configDiffData" :is-tpl="true" :id="props.templateSpaceId" :loading="false">
        <template #leftHead>
          <slot name="baseHead">
            <div class="diff-panel-head">
              <div class="version-tag base-version">{{t('对比版本')}}</div>
              <bk-select
                :model-value="selectedVersion"
                style="width: 320px"
                :loading="versionListLoading"
                :clearable="false"
                @change="handleSelectVersion"
              >
                <bk-option
                  v-for="version in versionList"
                  :key="version.id"
                  :label="version.spec.revision_name"
                  :value="version.id"
                >
                </bk-option>
              </bk-select>
            </div>
          </slot>
        </template>
        <template #rightHead>
          <slot name="currentHead">
            <div class="diff-panel-head">
              <div class="version-tag">{{t('当前版本')}}</div>
              <div class="version-name">{{ props.crtVersion.name }}</div>
            </div>
          </slot>
        </template>
      </diff>
    </div>
    <div class="actions-btn">
      <bk-button @click="handleClose">{{t('关闭')}}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { ITemplateVersionItem, DiffSliderDataType } from '../../../../../types/template';
import { IDiffDetail } from '../../../../../types/service';
import {
  getTemplateVersionsDetailByIds,
  getTemplateVersionList,
  downloadTemplateContent,
} from '../../../../api/template';

import { byteUnitConverse } from '../../../../utils';
import Diff from '../../../../components/diff/index.vue';

const { t } = useI18n();
const props = defineProps<{
  show: boolean;
  spaceId: string;
  templateSpaceId: number;
  crtVersion: DiffSliderDataType;
}>();

const emits = defineEmits(['update:show']);

const selectedVersion = ref();
const versionList = ref<ITemplateVersionItem[]>([]);
const versionListLoading = ref(false);
const configDiffData = ref<IDiffDetail>({
  contentType: 'text',
  current: {
    content: '',
  },
  base: {
    content: '',
  },
});

watch(
  () => props.show,
  async (val) => {
    if (val) {
      getVersionList();
      const detail: ITemplateVersionItem = await getTemplateVersionDetail(props.crtVersion.versionId);
      configDiffData.value.contentType = detail.spec.file_type === 'binary' ? 'file' : 'text';
      configDiffData.value.current.content = await getConfigContent(detail);
      configDiffData.value.current.permission = props.crtVersion.permission;
    }
  },
);

// 获取版本列表
const getVersionList = async () => {
  versionListLoading.value = true;
  const params = {
    start: 0,
    all: true,
  };
  const res = await getTemplateVersionList(props.spaceId, props.templateSpaceId, props.crtVersion.id, params);
  versionList.value = res.details.filter((item: ITemplateVersionItem) => item.id !== props.crtVersion.versionId);
  versionListLoading.value = false;
};

const getTemplateVersionDetail = async (versionId: number) => {
  const res = getTemplateVersionsDetailByIds(props.spaceId, [versionId]).then(res => res.details[0]);
  return res;
};

const getConfigContent = async (config: ITemplateVersionItem) => {
  const { id, spec, revision } = config;
  const { name, content_spec } = spec;
  const { signature, byte_size } = content_spec;
  if (config.spec.file_type === 'binary') {
    return { id, name, signature, update_at: revision.create_at, size: byteUnitConverse(Number(byte_size)) };
  }

  const configContent = await downloadTemplateContent(props.spaceId, props.templateSpaceId, signature);

  return String(configContent);
};

const handleSelectVersion = async (id: number) => {
  const version = versionList.value.find(item => item.id === id);
  if (version) {
    configDiffData.value.base.content = await getConfigContent(version);
    configDiffData.value.base.permission = version.spec.permission;
  }
};

const handleClose = () => {
  selectedVersion.value = undefined;
  versionList.value = [];
  configDiffData.value = {
    contentType: 'text',
    current: {
      content: '',
    },
    base: {
      content: '',
    },
  };
  emits('update:show', false);
};
</script>
<style lang="scss" scoped>
.diff-content-area {
  height: calc(100vh - 100px);
}
.diff-panel-head {
  display: flex;
  align-items: center;
  padding: 0 16px;
  width: 100%;
  height: 100%;
  font-size: 12px;
  color: #b6b6b6;
  background: #313238;
  .version-tag {
    margin-right: 8px;
    padding: 0 10px;
    height: 22px;
    line-height: 22px;
    font-size: 12px;
    color: #14a568;
    background: #e4faf0;
    border-radius: 2px;
    &.base-version {
      color: #3a84ff;
      background: #edf4ff;
    }
  }
  :deep(.bk-select) {
    .bk-input {
      border-color: #63656e;
    }
    .bk-input--text {
      color: #b6b6b6;
      background: #313238;
    }
  }
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
