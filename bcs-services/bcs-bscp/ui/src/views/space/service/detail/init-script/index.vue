<script setup lang="ts">
  import { ref, computed, onMounted, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import BkMessage from 'bkui-vue/lib/message';
  import { IScriptItem } from '../../../../../../types/script'
  import { useGlobalStore } from '../../../../../store/global'
  import { useConfigStore } from '../../../../../store/config'
  import { getScriptList, getScriptVersionList, getScriptVersionDetail } from '../../../../../api/script'
  import { getConfigInitScript, updateConfigInitScript } from '../../../../../api/config'
  import ScriptEditor from '../../../scripts/components/script-editor.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    appId: number;
  }>()

  const scriptsLoading = ref(false)
  const scriptsData = ref<{ id: number; name: string; type: string; }[]>([])
  const previewConfig = ref({
    open: false,
    type: '',
    name: '',
    content: ''
  })
  const contentLoading = ref(false)
  const scriptCiteData = ref({
    post_hook_id: 0,
    post_hook_release_id: 0,
    pre_hook_id: 0,
    pre_hook_release_id: 0
  })
  const scriptCiteDataLoading = ref(false)
  const formData = ref({
    pre: 0,
    post: 0
  })
  const pending = ref(false)

  // 查看模式
  const viewMode = computed(() => {
    return typeof versionData.value.id === 'number' && versionData.value.id !== 0
  })

  // 配置数据是否修改
  const dataChanged = computed(() => {
    const { pre_hook_id, post_hook_id } = scriptCiteData.value
    const { pre, post } = formData.value
    return pre_hook_id !== pre || post_hook_id !== post
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
    const list = res.details.map((item: IScriptItem) => {
      return { id: item.id, name: item.spec.name, type: item.spec.type }
    })
    scriptsData.value = [{ id: 0, name: '<不使用脚本>' }, ...list]
    scriptsLoading.value = false
  }

  // 获取初始化脚本配置
  const getScriptSetting = async () => {
    if (versionData.value.id) {
      scriptCiteData.value = { ...versionData.value.spec.hook }
    } else {
      scriptCiteDataLoading.value = true
      const res = await getConfigInitScript(spaceId.value, props.appId)
      scriptCiteData.value = res.data.config_hook.spec
      scriptCiteDataLoading.value = false
    }
    formData.value = {
      pre: scriptCiteData.value.pre_hook_id,
      post: scriptCiteData.value.post_hook_id
    }
  }

  const getPreviewContent = async (scriptId: number, versionId: number) => {
    contentLoading.value = true
    if (viewMode.value) { // 查看模式，直接通过脚本版本id查询参数版本详情
      const res = await getScriptVersionDetail(spaceId.value, scriptId, versionId)
      previewConfig.value.content = res.spec.content
    } else { // 编辑模式，通过已上线状态去筛选脚本版本列表数据
      const params = {
        start: 0,
        all: true,
        state: 'deployed'
      }
      const res = await getScriptVersionList(spaceId.value, scriptId, params)
      if (res.details[0]) {
        previewConfig.value.content = res.details[0].spec.content
      }
    }
    contentLoading.value = false
  }

  const handleSelectScript = (type: string) => {
    if (previewConfig.value.open) {
      handleOpenPreview(type)
    }
  }

  // 点击预览
  const handleOpenPreview = (type: string) => {
    const id = type === 'pre' ? formData.value.pre : formData.value.post
    const versionId = type === 'pre' ? scriptCiteData.value.pre_hook_release_id : scriptCiteData.value.post_hook_release_id
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
        pre_hook_id: pre,
        post_hook_id: post
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
            <bk-select
              v-model="formData.pre"
              :clearable="false"
              :disabled="viewMode"
              :loading="scriptsLoading"
              @change="handleSelectScript('pre')">
              <bk-option v-for="script in scriptsData" :key="script.id" :value="script.id" :label="script.name"></bk-option>
            </bk-select>
            <bk-button
              class="preview-button"
              text
              theme="primary"
              :disabled="formData.pre === 0"
              @click="handleOpenPreview('pre')">
              预览
            </bk-button>
          </div>
        </bk-form-item>
        <bk-form-item label="后置脚本">
          <div class="select-wrapper">
            <bk-select
              v-model="formData.post"
              :clearable="false"
              :disabled="viewMode"
              :loading="scriptsLoading"
              @change="handleSelectScript('post')">
              <bk-option v-for="script in scriptsData" :key="script.id" :value="script.id" :label="script.name"></bk-option>
            </bk-select>
            <bk-button
              class="preview-button"
              text
              theme="primary"
              :disabled="formData.post === 0"
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
          <div class="script-preview-title">{{ `脚本预览 - ${previewConfig.name}` }}</div>
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
    padding: 0 24px;
    line-height: 40px;
    color: #c4c6cc;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  :deep(.script-editor) {
    height: 100%;
    .content-wrapper {
      height: calc(100% - 40px);
    }
  }
</style>
