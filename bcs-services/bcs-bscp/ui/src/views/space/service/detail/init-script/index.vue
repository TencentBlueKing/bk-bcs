<template>
  <div class="init-script-page">
    <div class="script-select-area">
      <bk-form form-type="vertical">
        <bk-form-item :label="t('前置脚本')">
          <div class="select-wrapper">
            <ScriptSelector
              type="pre"
              :id="formData.pre.id"
              :disabled="viewMode"
              :loading="scriptsLoading"
              :list="scriptsData"
              @change="handleSelectScript"
              @refresh="getScripts" />
            <bk-button
              class="preview-button"
              text
              theme="primary"
              :disabled="typeof formData.pre.id !== 'number' || formData.pre.id === 0"
              @click="handleOpenPreview('pre')">
              {{ t('预览') }}
            </bk-button>
          </div>
        </bk-form-item>
        <bk-form-item :label="t('后置脚本')">
          <div class="select-wrapper">
            <ScriptSelector
              type="post"
              :id="formData.post.id"
              :disabled="viewMode"
              :loading="scriptsLoading"
              :list="scriptsData"
              @change="handleSelectScript"
              @refresh="getScripts" />
            <bk-button
              class="preview-button"
              text
              theme="primary"
              :disabled="typeof formData.post.id !== 'number' || formData.post.id === 0"
              @click="handleOpenPreview('post')">
              {{ t('预览') }}
            </bk-button>
          </div>
        </bk-form-item>
      </bk-form>
      <bk-button
        v-if="!viewMode"
        v-cursor="{ active: !hasEditServicePerm }"
        :class="['submit-button', { 'bk-button-with-no-perm': !hasEditServicePerm }]"
        theme="primary"
        :disabled="hasEditServicePerm && submitButtonDisabled"
        :loading="pending"
        @click="handleSubmit">
        {{ t('保存设置') }}
      </bk-button>
    </div>
    <bk-loading v-if="previewConfig.open" class="preview-area" :loading="contentLoading">
      <ScriptEditor
        :model-value="previewConfig.content"
        :editable="false"
        :upload-icon="false"
        :language="previewConfig.type"
        :is-preview="true">
        <template #header>
          <div class="script-preview-title">
            <div class="close-area" @click="previewConfig.open = false">
              <AngleRight class="arrow-icon" />
            </div>
            <div class="title">{{ `${t('脚本预览')} - ${previewConfig.name}` }}</div>
          </div>
        </template>
      </ScriptEditor>
    </bk-loading>
  </div>
</template>
<script setup lang="ts">
  import { ref, computed, onMounted, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import BkMessage from 'bkui-vue/lib/message';
  import { AngleRight } from 'bkui-vue/lib/icon';
  import { IScriptItem } from '../../../../../../types/script';
  import { ICommonQuery } from '../../../../../../types/index';
  import { IBoundTemplateGroup } from '../../../../../../types/config';
  import useGlobalStore from '../../../../../store/global';
  import useServiceStore from '../../../../../store/service';
  import useConfigStore from '../../../../../store/config';
  import { getScriptList, getScriptVersionDetail } from '../../../../../api/script';
  import {
    getConfigList,
    getBoundTemplates,
    getConfigScript,
    getDefaultConfigScriptData,
    updateConfigInitScript,
  } from '../../../../../api/config';
  import ScriptEditor from '../../../scripts/components/script-editor.vue';
  import ScriptSelector from './script-selector.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const configStore = useConfigStore();
  const serviceStore = useServiceStore();
  const { versionData } = storeToRefs(configStore);
  const { checkPermBeforeOperate } = serviceStore;
  const { permCheckLoading, hasEditServicePerm } = storeToRefs(serviceStore);
  const { t } = useI18n();

  const props = defineProps<{
    appId: number;
  }>();

  const scriptsLoading = ref(false);
  const scriptsData = ref<{ id: number; versionId: number; name: string; type: string }[]>([]);
  const previewConfig = ref({
    open: false,
    type: '',
    name: '',
    content: '',
  });
  const contentLoading = ref(false);
  const scriptCiteData = ref({
    pre_hook: getDefaultConfigScriptData(),
    post_hook: getDefaultConfigScriptData(),
  });
  const scriptCiteDataLoading = ref(false);
  const submitButtonDisabled = ref(true);
  const formData = ref<{ pre: { id: number; versionId: number }; post: { id: number; versionId: number } }>({
    pre: {
      id: 0,
      versionId: 0,
    },
    post: {
      id: 0,
      versionId: 0,
    },
  });
  const pending = ref(false);

  // 查看模式
  const viewMode = computed(() => typeof versionData.value.id === 'number' && versionData.value.id !== 0);

  watch(
    () => versionData.value.id,
    (id) => {
      getScriptSetting();
      previewConfig.value.open = false;
      if (id === 0) {
        getAllConfigList();
      }
    },
  );

  onMounted(() => {
    getScripts();
    getScriptSetting();
    if (versionData.value.id === 0) {
      getAllConfigList();
    }
  });
  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true;
    const params = {
      start: 0,
      all: true,
    };
    const res = await getScriptList(spaceId.value, params);
    const list = (res.details as IScriptItem[])
      .map((item) => ({
        id: item.hook.id,
        versionId: item.published_revision_id,
        name: item.hook.spec.name,
        type: item.hook.spec.type,
      }))
      .sort((a, b) => a.name.localeCompare(b.name, 'zh-Hans-CN'));
    scriptsData.value = [{ id: 0, versionId: 0, name: t('<不使用脚本>'), type: '' }, ...list];
    scriptsLoading.value = false;
  };

  // 获取初始化脚本配置
  const getScriptSetting = async () => {
    scriptCiteDataLoading.value = true;
    scriptCiteData.value = await getConfigScript(spaceId.value, props.appId, versionData.value.id);
    scriptCiteDataLoading.value = false;
    formData.value = {
      pre: {
        id: scriptCiteData.value.pre_hook.hook_id,
        versionId: scriptCiteData.value.pre_hook.hook_revision_id,
      },
      post: {
        id: scriptCiteData.value.post_hook.hook_id,
        versionId: scriptCiteData.value.post_hook.hook_revision_id,
      },
    };
  };

  // 获取脚本预览内容
  const getPreviewContent = async (scriptId: number, versionId: number) => {
    contentLoading.value = true;
    const res = await getScriptVersionDetail(spaceId.value, scriptId, versionId);
    previewConfig.value.content = res.spec.content;
    contentLoading.value = false;
  };

  const getAllConfigList = async () => {
    configStore.$patch((state) => {
      state.createVersionBtnLoading = true;
    });
    const [existConfigCount, tplConfigCount] = await Promise.all([getCommonConfigList(), getBoundTemplateList()]);
    configStore.$patch((state) => {
      state.createVersionBtnLoading = false;
      state.allExistConfigCount = existConfigCount + tplConfigCount;
    });
  };

  // 获取非模板配置文件列表
  const getCommonConfigList = async () => {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    const res = await getConfigList(spaceId.value, props.appId, params);
    return res.details.filter((item: any) => item.file_state !== 'DELETE').length;
  };

  // 获取模板配置文件列表
  const getBoundTemplateList = async () => {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    const res = await getBoundTemplates(spaceId.value, props.appId, params);
    return res.details.reduce((acc: number, crt: IBoundTemplateGroup) => acc + crt.template_revisions.length, 0);
  };

  // 选择脚本
  const handleSelectScript = (id: number, type: string) => {
    const script = scriptsData.value.find((item) => item.id === id);
    if (script) {
      if (type === 'pre') {
        formData.value.pre.versionId = script.versionId;
        formData.value.pre.id = id;
      } else {
        formData.value.post.versionId = script.versionId;
        formData.value.post.id = id;
      }
    }
    submitButtonDisabled.value = false;
    handleOpenPreview(type);
  };

  // 点击预览
  const handleOpenPreview = (type: string) => {
    const id = type === 'pre' ? formData.value.pre.id : formData.value.post.id;
    const versionId = type === 'pre' ? formData.value.pre.versionId : formData.value.post.versionId;
    const script = scriptsData.value.find((item) => item.id === id);
    if (script && script.id > 0) {
      previewConfig.value = {
        open: true,
        name: script.name,
        type: script.type,
        content: '',
      };
      getPreviewContent(script.id, versionId);
    } else {
      previewConfig.value.open = false;
    }
  };

  // 保存配置
  const handleSubmit = async () => {
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    try {
      pending.value = true;
      const { pre, post } = formData.value;
      const params = {
        pre_hook_id: pre.id,
        post_hook_id: post.id,
      };
      await updateConfigInitScript(spaceId.value, props.appId, params);
      BkMessage({
        theme: 'success',
        message: t('初始化脚本设置成功'),
      });
    } catch (e) {
      console.error(e);
    } finally {
      pending.value = false;
      submitButtonDisabled.value = true;
    }
  };
</script>
<style lang="scss" scoped>
  .init-script-page {
    display: flex;
    align-items: top;
    height: 100%;
  }
  .script-select-area {
    padding: 24px 32px 24px 24px;
    width: 528px;
    height: 100%;
    .select-wrapper {
      display: flex;
      align-items: center;
      justify-content: space-between;
      .bk-select {
        width: 426px;
      }
    }
  }
  .preview-area {
    width: calc(100% - 528px);
    height: 100%;
  }
  .script-preview-title {
    display: flex;
    align-items: center;
    padding-right: 24px;
    width: 100%;
    height: 40px;
    .close-area {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 100%;
      background: #63656e;
      color: #ffffff;
      font-size: 20px;
      cursor: pointer;
    }
    .title {
      padding: 0 5px;
      line-height: 40px;
      color: #c4c6cc;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
  :deep(.script-editor) {
    height: 100%;
    .content-wrapper {
      height: calc(100% - 40px);
    }
  }
</style>
