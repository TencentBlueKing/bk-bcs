<script setup lang="ts">
  import { ref, watch, computed } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue';
  import ConfigForm from './config-form.vue'
  import { getConfigItemDetail, updateConfigContent, getConfigContent, updateServiceConfigItem } from '../../../../../../../api/config'
  import { getConfigEditParams } from '../../../../../../../utils/config'
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config'
  import { useConfigStore } from '../../../../../../../store/config'
  import useModalCloseConfirmation from '../../../../../../../utils/hooks/use-modal-close-confirmation'

  const { versionData } = storeToRefs(useConfigStore())

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    configId: number,
    show: Boolean
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const configDetailLoading = ref(true)
  const configForm = ref<IConfigEditParams>(getConfigEditParams())
  const content = ref<string|IFileConfigContentSummary>('')
  const formRef = ref()
  const fileUploading = ref(false)
  const pending = ref(false)
  const isFormChange = ref(false)

  const editable = computed(() => {
    return versionData.value.id === 0
  })

  watch(
    () => props.show,
    (val) => {
      if (val) {
        getConfigDetail()
      }
    }
  )

  // 获取配置项详情配置及配置内容
  const getConfigDetail = async() => {
    try {
      configDetailLoading.value = true
      const params: { release_id?: number } = {}
      if (versionData.value.id) {
        params.release_id = versionData.value.id
      }
      const detail = await getConfigItemDetail(props.bkBizId, props.configId, props.appId, params)
      const { name, memo, path, file_type, permission } = detail.config_item.spec
      configForm.value = { id: props.configId, name, memo, file_type, path, ...permission }
      const signature = detail.content.signature
      if (file_type === 'binary') {
        content.value = { name, signature, size: detail.content.byte_size }
      } else {
        const configContent = await getConfigContent(props.bkBizId, props.appId, signature)
        content.value = String(configContent)
      }
    } catch (e) {
      console.error(e)
    } finally {
      configDetailLoading.value = false
    }
  }

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation()
      return result
    }
    return true
  }

  const handleChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary|string) => {
    configForm.value = data
    content.value = configContent
    isFormChange.value = true
  }

  const handleSubmit = async() => {
    const isValid = await formRef.value.validate()
    if (!isValid) return

    try {
      pending.value = true
      let sign = await formRef.value.getSignature()
      let size = 0
      if (configForm.value.file_type === 'binary') {
        size = Number((<IFileConfigContentSummary>content.value).size)
      } else {
        const stringContent = <string>content.value
        size = new Blob([stringContent]).size
        await updateConfigContent(props.bkBizId, props.appId, stringContent, sign)
      }
      const params = { ...configForm.value, ...{ sign, byte_size: size } }
      await updateServiceConfigItem(props.configId, props.appId, props.bkBizId, params)
      emits('confirm')
      Message({
        theme: 'success',
        message: '编辑配置项成功'
      })
    }catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
    <bk-sideslider
      width="640"
      :title="`${editable ? '编辑' : '查看'}配置项`"
      :is-show="props.show"
      :before-close="handleBeforeClose"
      @closed="close">
        <bk-loading :loading="configDetailLoading" style="height: 100%;">
          <ConfigForm
            v-if="!configDetailLoading"
            ref="formRef"
            class="config-form-wrapper"
            v-model:fileUploading="fileUploading"
            :config="configForm"
            :content="content"
            :editable="editable"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            @change="handleChange" />
        </bk-loading>
        <section class="action-btns">
          <bk-button
            theme="primary"
            :loading="pending"
            :disabled="configDetailLoading || fileUploading"
            @click="handleSubmit">
            保存
          </bk-button>
          <bk-button @click="close">取消</bk-button>
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
