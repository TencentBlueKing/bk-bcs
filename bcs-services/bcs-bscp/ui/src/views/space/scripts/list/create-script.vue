<template>
  <DetailLayout :name="t('新建脚本')" @close="handleClose">
    <template #content>
      <div class="create-script-forms">
        <bk-form ref="formRef" form-type="vertical" :model="formData" :rules="rules">
          <bk-form-item class="fixed-width-form" :label="t('脚本名称')" property="name" required>
            <bk-input v-model="formData.name" :placeholder="t('请输入')"/>
          </bk-form-item>
          <bk-form-item class="fixed-width-form" property="tag" :label="t('分类标签')">
            <!-- <bk-input v-model="formData.tag" /> -->
            <bk-tag-input
              v-model="selectTags"
              :placeholder="t('请选择标签或输入新标签按Enter结束')"
              display-key="tag"
              save-key="tag"
              search-key="tag"
              :max-data="1"
              :list="tagsData"
              :allow-create="true"
              trigger="focus"/>
          </bk-form-item>
          <bk-form-item class="fixed-width-form" property="memo" :label="t('脚本描述')">
            <bk-input
              v-model="formData.memo"
              type="textarea"
              :placeholder="t('请输入')"
              :rows="3"
              :maxlength="200"
              :resize="true" />
          </bk-form-item>
          <bk-form-item :label="t('脚本内容')" property="content" required>
            <div class="script-content-wrapper">
              <ScriptEditor v-model="showContent" :language="formData.type">
                <template #header>
                  <div class="language-tabs">
                    <div
                      v-for="item in SCRIPT_TYPE"
                      :key="item.id"
                      :class="['tab', { actived: formData.type === item.id }]"
                      @click="formData.type = item.id">
                      {{ item.name }}
                    </div>
                  </div>
                </template>
              </ScriptEditor>
            </div>
          </bk-form-item>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="handleCreate">{{ t('创建') }}</bk-button>
        <bk-button @click="handleClose">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </DetailLayout>
</template>
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import BkMessage from 'bkui-vue/lib/message';
import useGlobalStore from '../../../../store/global';
import { EScriptType, IScriptEditingForm, IScriptTagItem } from '../../../../../types/script';
import { createScript, getScriptTagList } from '../../../../api/script';
import DetailLayout from '../components/detail-layout.vue';
import ScriptEditor from '../components/script-editor.vue';

const { spaceId } = storeToRefs(useGlobalStore());
const { t } = useI18n();

const emits = defineEmits(['update:show', 'created']);

const SCRIPT_TYPE = [
  { id: EScriptType.Shell, name: 'Shell' },
  { id: EScriptType.Python, name: 'Python' },
];

const formRef = ref();
const pending = ref(false);
const tagsLoading = ref(false);
const tagsData = ref<IScriptTagItem[]>([]);
const selectTags = ref<string[]>([]);
const formData = ref<IScriptEditingForm>({
  name: '',
  tag: '',
  memo: '',
  type: EScriptType.Shell,
  content: '',
});
const formDataContent = ref({
  shell: '#!/bin/bash',
  python: '#!/usr/bin/env python\n# -*- coding: utf8 -*-',
});
const showContent = computed({
  get: () => (formData.value.type === 'shell' ? formDataContent.value.shell : formDataContent.value.python),
  set: (val) => {
    if (!val.endsWith('\n')) {
      val += '\n';
    }
    formData.value.type === 'shell' ? formDataContent.value.shell = val : formDataContent.value.python = val;
  },
});

const rules = {
  name: [
    {
      validator: (value: string) => value.length <= 64,
      message: t('不能超过64个字符'),
      trigger: 'change',
    },
  ],
  memo: [
    {
      validator: (value: string) => value.length <= 200,
      message: t('最大长度200个字符'),
    },
  ],
};

onMounted(() => {
  getTags();
});

// 获取标签列表
const getTags = async () => {
  tagsLoading.value = true;
  const res = await getScriptTagList(spaceId.value);
  tagsData.value = res.details;
  tagsLoading.value = false;
};

const handleCreate = async () => {
  formData.value.content = formData.value.type === 'shell' ? formDataContent.value.shell : formDataContent.value.python;
  await formRef.value.validate();
  try {
    pending.value = true;
    formData.value.tag = selectTags.value[0];
    await createScript(spaceId.value, formData.value);
    BkMessage({
      theme: 'success',
      message: t('脚本创建成功'),
    });
    handleClose();
    emits('created');
  } catch (e) {
    console.error(e);
  } finally {
    pending.value = false;
  }
};

const handleClose = () => {
  emits('update:show', false);
};
</script>
<style scoped lang="scss">
.create-script-forms {
  padding: 24px 48px;
  height: 100%;
  background: #f5f7fa;
  overflow: auto;
}
.fixed-width-form {
  width: 520px;
}
.script-content-wrapper {
  min-width: 520px;
}

.language-tabs {
  display: flex;
  align-items: center;
  background: #2e2e2e;
  .tab {
    padding: 10px 24px;
    line-height: 20px;
    font-size: 14px;
    color: #8a8f99;
    border-top: 3px solid #2e2e2e;
    cursor: pointer;
    &.actived {
      color: #c4c6cc;
      font-weight: 700;
      background: #1a1a1a;
      border-color: #3a84ff;
    }
  }
}
.actions-wrapper {
  .bk-button {
    margin-right: 8px;
    min-width: 88px;
  }
}
</style>
