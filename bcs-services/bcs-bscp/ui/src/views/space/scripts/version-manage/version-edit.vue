<template>
  <section :class="['script-content', { 'view-mode': !props.editable }]">
    <ScriptEditor
      v-model="localVal.content"
      :language="props.type"
      :editable="props.editable"
      :upload-icon="props.editable">
      <template #header>
        <div class="title">{{ title }}</div>
      </template>
      <template v-if="props.editable" #preContent="{ fullscreen }">
        <div v-show="!fullscreen" class="version-config-form">
          <bk-form ref="formRef" :rules="rules" form-type="vertical" :model="localVal">
            <bk-form-item :label="t('版本号')" property="name">
              <bk-input v-model="localVal.name" :placeholder="t('请输入')" />
            </bk-form-item>
            <bk-form-item :label="t('版本说明')" propperty="memo">
              <bk-input v-model="localVal.memo" type="textarea" :placeholder="t('请输入')" :rows="8" :resize="true" />
            </bk-form-item>
          </bk-form>
        </div>
      </template>
    </ScriptEditor>
    <div v-if="props.editable" class="action-btns">
      <div v-if="!isEditVersion">
        <bk-button class="submit-btn" theme="primary" :loading="pending" @click="handleSubmit">
          {{ t('保存') }}
        </bk-button>
        <bk-button class="cancel-btn" @click="emits('close')">{{ t('取消') }}</bk-button>
      </div>
      <div v-else>
        <bk-button class="submit-btn" theme="primary" :loading="pending" @click="handleSubmit">
          {{ t('上线') }}
        </bk-button>
        <bk-button class="edit-btn" @click="emits('close')">{{ t('编辑') }}</bk-button>
        <bk-button class="cancel-btn" @click="emits('close')">{{ t('删除') }}</bk-button>
      </div>
    </div>
  </section>
</template>
<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import BkMessage from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../store/global';
  import { IScriptVersionForm } from '../../../../../types/script';
  import { createScriptVersion, updateScriptVersion } from '../../../../api/script';
  import ScriptEditor from '../components/script-editor.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      type: string;
      editable?: boolean;
      scriptId: number;
      versionData: IScriptVersionForm;
    }>(),
    {
      editable: true,
    },
  );

  const emits = defineEmits(['close', 'submitted']);

  const rules = {
    name: [
      {
        validator: (value: string) =>
          /^[\u4e00-\u9fa5a-zA-Z0-9][\u4e00-\u9fa5a-zA-Z0-9_\-()\s]*[\u4e00-\u9fa5a-zA-Z0-9]$/.test(value),
        message: t(
          '无效名称，只允许包含中文、英文、数字、下划线()、连字符(-)、空格，且必须以中文、英文、数字开头和结尾',
        ),
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
  const localVal = ref<IScriptVersionForm>({
    id: 0,
    name: '',
    memo: '',
    content: '',
  });
  const formRef = ref();
  const pending = ref(false);

  const title = computed(() => {
    if (!props.editable) {
      return props.versionData.name;
    }
    return isEditVersion.value ? t('编辑版本') : t('新建版本');
  });

  const isEditVersion = computed(() => !!props.versionData.id);

  watch(
    () => props.versionData,
    (val) => {
      localVal.value = { ...val };
    },
    { immediate: true },
  );

  const handleSubmit = async () => {
    await formRef.value.validate();
    if (!localVal.value.content) {
      BkMessage({
        theme: 'error',
        message: t('脚本内容不能为空'),
      });
      return;
    }
    try {
      pending.value = true;
      if (!localVal.value.content.endsWith('\n')) {
        localVal.value.content += '\n';
      }
      const { name, memo, content } = localVal.value;
      const params = { name, memo, content };
      if (localVal.value.id) {
        await updateScriptVersion(spaceId.value, props.scriptId, localVal.value.id, params);
        emits('submitted', { ...localVal.value }, 'update');
        BkMessage({
          theme: 'success',
          message: t('编辑版本成功'),
        });
      } else {
        const res = await createScriptVersion(spaceId.value, props.scriptId, params);
        emits('submitted', { ...localVal.value, id: res.id }, 'create');
        BkMessage({
          theme: 'success',
          message: t('新建版本成功'),
        });
      }
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
    }
  };
</script>
<style lang="scss" scoped>
  .script-content {
    height: 100%;
    background: #2a2a2a;
    :deep(.script-editor) {
      height: calc(100% - 46px);
    }
    &.view-mode {
      :deep(.script-editor) {
        height: 100%;
        .code-editor-wrapper {
          width: 100%;
        }
      }
    }
  }
  .title {
    padding: 10px 24px;
    line-height: 20px;
    font-size: 14px;
    color: #8a8f99;
  }
  .version-config-form {
    padding: 24px 20px 24px;
    width: 260px;
    :deep(.bk-form) {
      .bk-form-label {
        font-size: 12px;
        color: #979ba5;
      }
      .bk-form-item {
        margin-bottom: 40px !important;
      }
      .bk-input {
        border: 1px solid #63656e;
      }
      .bk-input--text {
        background: transparent;
        color: #c4c6cc;
        &::placeholder {
          color: #63656e;
        }
      }
      .bk-textarea {
        background: transparent;
        border: 1px solid #63656e;
        textarea {
          color: #c4c6cc;
          background: transparent;
          &::placeholder {
            color: #63656e;
          }
        }
      }
    }
  }
  :deep(.script-editor) {
    .content-wrapper {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      height: calc(100% - 40px);
    }
    .code-editor-wrapper {
      width: calc(100% - 260px);
    }
  }
  .action-btns {
    padding: 7px 24px;
    background: #2a2a2a;
    box-shadow: 0 -1px 0 0 #141414;
    .submit-btn {
      margin-right: 8px;
      min-width: 120px;
    }
    .cancel-btn {
      min-width: 88px;
      background: transparent;
      border-color: #979ba5;
      color: #979ba5;
    }
    .edit-btn {
      @extend .cancel-btn;
      margin-right: 8px;
    }
  }
</style>
