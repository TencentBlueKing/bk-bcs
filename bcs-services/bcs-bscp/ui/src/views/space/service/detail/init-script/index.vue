<script setup lang="ts">
  import { ref, computed, onMounted, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import BkMessage from 'bkui-vue/lib/message';
  import { AngleRight } from 'bkui-vue/lib/icon';
  import { IScriptItem } from '../../../../../../types/script'
  import { useGlobalStore } from '../../../../../store/global'
  import { useConfigStore } from '../../../../../store/config'
  import { getScriptList, getScriptVersionDetail } from '../../../../../api/script'
  import { getConfigScript, getDefaultConfigScriptData, updateConfigInitScript } from '../../../../../api/config'
  import ScriptEditor from '../../../scripts/components/script-editor.vue'
  import ScriptSelector from './script-selector.vue';

  const { spaceId } = storeToRefs(useGlobalStore())
  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    appId: number;
  }>()

  const scriptsLoading = ref(false)
  const scriptsData = ref<{ id: number; versionId: number; name: string; type: string; }[]>([])
  const previewConfig = ref({
    open: false,
    type: '',
    name: '',
    content: ''
  })
  const contentLoading = ref(false)
  const scriptCiteData = ref({
    pre_hook: getDefaultConfigScriptData(),
    post_hook: getDefaultConfigScriptData()
  })
  const scriptCiteDataLoading = ref(false)
  const formData = ref<{ pre: { id: number; versionId: number }; post: { id: number; versionId: number } }>({
    pre: {
      id: 0,
      versionId: 0
    },
    post: {
      id: 0,
      versionId: 0
    }
  })
  const pending = ref(false)

  // 查看模式
  const viewMode = computed(() => {
    return typeof versionData.value.id === 'number' && versionData.value.id !== 0
  })

  // 配置数据是否修改
  const dataChanged = computed(() => {
    const { pre_hook, post_hook } = scriptCiteData.value
    const { pre, post } = formData.value
    return pre_hook.hook_id !== pre.id || post_hook.hook_id !== post.id
  })

  watch(() => versionData.value.id, () => {
    getScriptSetting()
    previewConfig.value.open = false
  })

  onMounted(() => {
    getScripts()
    getScriptSetting()
  })
  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params = {
      start: 0,
      all: true
    }
    const res = await getScriptList(spaceId.value, params)
    const list = (<IScriptItem[]>res.details).map(item => {
      return { id: item.hook.id, versionId: item.published_revision_id, name: item.hook.spec.name, type: item.hook.spec.type }
    })
    scriptsData.value = [{ id: 0, versionId: 0, name: '<不使用脚本>', type: '' }, ...list]
    scriptsLoading.value = false
  }

  // 获取初始化脚本配置
  const getScriptSetting = async () => {
    scriptCiteDataLoading.value = true
    scriptCiteData.value = await getConfigScript(spaceId.value, props.appId, versionData.value.id)
    scriptCiteDataLoading.value = false
    formData.value = {
      pre: {
        id: scriptCiteData.value.pre_hook.hook_id,
        versionId: scriptCiteData.value.pre_hook.hook_revision_id
      },
      post: {
        id: scriptCiteData.value.post_hook.hook_id,
        versionId: scriptCiteData.value.post_hook.hook_revision_id
      }
    }
  }

  // 获取脚本预览内容
  const getPreviewContent = async (scriptId: number, versionId: number) => {
    contentLoading.value = true
    const res = await getScriptVersionDetail(spaceId.value, scriptId, versionId)
    previewConfig.value.content = res.spec.content
    contentLoading.value = false
  }

  // 选择脚本
  const handleSelectScript = (id: number, type: string) => {
    const script = scriptsData.value.find(item => item.id === id)
    if (script) {
      if (type === 'pre') {
        formData.value.pre.versionId = script.versionId
        formData.value.pre.id = id
      } else {
        formData.value.post.versionId = script.versionId
        formData.value.post.id = id
      }
    }
    if (id === 0) {
      previewConfig.value.open = false
    }
    if (previewConfig.value.open) {
      handleOpenPreview(type)
    }
  }

  // 点击预览
  const handleOpenPreview = (type: string) => {
    const id = type === 'pre' ? formData.value.pre.id : formData.value.post.id
    const versionId = type === 'pre' ? formData.value.pre.versionId : formData.value.post.versionId
    const script = scriptsData.value.find(item => item.id === id)
    if (script) {
      previewConfig.value = {
        open: true,
        name: script.name,
        type: script.type,
        content: ''
      }
      getPreviewContent(script.id, versionId)
    }
  }

  // 保存配置
  const handleSubmit = async() => {
    try {
      pending.value = true
      const { pre, post } = formData.value
      const params = {
        pre_hook_id: pre.id,
        post_hook_id: post.id
      }
      await updateConfigInitScript(spaceId.value, props.appId, params)
      BkMessage({
        theme: 'success',
        message: '初始化脚本设置成功'
      })
    } catch (e) {
      console.error(e)
    } finally {
      pending.value = false
    }
  }

</script>
<template>
  <div class="init-script-page">
    <div class="script-select-area">
      <bk-form form-type="vertical">
        <bk-form-item label="前置脚本">
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
              预览
            </bk-button>
          </div>
        </bk-form-item>
        <bk-form-item label="后置脚本">
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
              预览
            </bk-button>
          </div>
        </bk-form-item>
      </bk-form>
      <bk-button
        v-if="!viewMode"
        class="submit-button"
        theme="primary"
        :disabled="!dataChanged"
        :loading="pending"
        @click="handleSubmit">
        保存设置
      </bk-button>
    </div>
    <bk-loading v-if="previewConfig.open" class="preview-area" :loading="contentLoading">
      <ScriptEditor :model-value="previewConfig.content" :editable="false" :upload-icon="false" :language="previewConfig.type">
        <template #header>
          <div class="script-preview-title">
            <div class="close-area" @click="previewConfig.open = false">
              <AngleRight class="arrow-icon" />
            </div>
            <div class="title">{{ `脚本预览 - ${previewConfig.name}` }}</div>
          </div>
        </template>
      </ScriptEditor>
    </bk-loading>
  </div>
</template>
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
