<script setup lang="ts">
  import { ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import ConfigForm from './config-form.vue'
  import { getConfigItemDetail, getReleasedConfigItemDetail, getConfigContent } from '../../../../../../../api/config'
  import { getTemplateVersionsDetailByIds, downloadTemplateContent } from '../../../../../../../api/template'
  import { getConfigEditParams } from '../../../../../../../utils/config'
  import { IVariableEditParams } from '../../../../../../../../types/variable';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config'
  import { getReleasedAppVariables } from '../../../../../../../api/variable'
  import { useConfigStore } from '../../../../../../../store/config'

  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    id: number;
    versionId: number;
    type: string; // 取值为config/template，分别表示非模板套餐下配置项和模板套餐下配置项
    show: Boolean;
  }>()

  const emits = defineEmits(['update:show'])

  const detailLoading = ref(true)
  const configForm = ref<IConfigEditParams>(getConfigEditParams())
  const content = ref<string|IFileConfigContentSummary>('')
  const variables = ref<IVariableEditParams[]>([])
  const variablesLoading = ref(false)

  watch(
    () => props.show,
    (val) => {
      if (val) {
        getDetailData()
      }
    }
  )

  const getDetailData = async() => {
    detailLoading.value = true
    if (props.type === 'config') {
      getConfigDetail()
    } else if (props.type === 'template') {
      getTemplateDetail()
    }
    // 未命名版本id为0，不需要展示变量替换
    if (props.versionId) {
      getVariableList()
    }
  }

  // 获取非模板套餐下配置项详情配置，非文件类型配置项内容下载内容，文件类型手动点击时再下载
  const getConfigDetail = async() => {
    try {
      let detail
      let signature
      let byte_size
      if (versionData.value.id) {
        detail = await getReleasedConfigItemDetail(props.bkBizId, props.appId, versionData.value.id, props.id)
        const { origin_byte_size, origin_signature } = detail.config_item.commit_spec.content
        byte_size = origin_byte_size
        signature = origin_signature
      } else {
        detail = await getConfigItemDetail(props.bkBizId, props.id, props.appId)
        byte_size = detail.content.byte_size
        signature = detail.content.signature
      }
      const { name, memo, path, file_type, permission } = detail.config_item.spec
      configForm.value = { id: props.id, name, memo, file_type, path, ...permission }
      if (file_type === 'binary') {
        content.value = { name, signature, size: byte_size }
      } else {
        const configContent = await getConfigContent(props.bkBizId, props.appId, signature)
        content.value = String(configContent)
      }
    } catch (e) {
      console.error(e)
    } finally {
      detailLoading.value = false
    }
  }

  // 获取模板配置详情，非文件类型配置项内容下载内容，文件类型手动点击时再下载
  const getTemplateDetail = async() => {
    try {
      detailLoading.value = true
      const res = await getTemplateVersionsDetailByIds(props.bkBizId, [props.id])
      if (res.details.length === 1) {
        const { attachment, spec } = res.details[0]
        const { name, path, revision_memo, file_type, permission } = spec
        const { signature, byte_size } = spec.content_spec
        configForm.value = { id: props.id, name: name, memo: revision_memo, file_type, path, ...permission }
        if (file_type === 'binary') {
          content.value = { name: name, signature, size: String(byte_size) }
        } else {
          const configContent = await downloadTemplateContent(props.bkBizId, attachment.template_space_id, signature)
          content.value = String(configContent)
        }
      }
    } catch (e) {
      console.error(e)
    } finally {
      detailLoading.value = false
    }
  }

  const getVariableList = async () => {
    variablesLoading.value = true
    const res = await getReleasedAppVariables(props.bkBizId, props.appId, props.versionId)
    variables.value = res.details
    variablesLoading.value = false
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
    <bk-sideslider
      width="640"
      :title="'查看配置项'"
      :is-show="props.show"
      @closed="close">
        <bk-loading :loading="detailLoading" class="config-loading-container">
          <ConfigForm
            v-if="props.show && !detailLoading"
            class="config-form-wrapper"
            :editable="false"
            :config="configForm"
            :content="content"
            :variables="variables"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"/>
        </bk-loading>
        <section class="action-btns">
          <bk-button @click="close">关闭</bk-button>
      </section>
    </bk-sideslider>
</template>
<style lang="scss" scoped>
  .config-loading-container {
    height: calc(100vh - 101px);
    overflow: auto;
    .config-form-wrapper {
      padding: 20px 40px;
      height: 100%;
    }
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
