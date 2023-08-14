<script setup lang="ts">
  import { ref } from 'vue'
  import { Message } from 'bkui-vue';
  import ConfigForm from './config-form.vue'
  import { createServiceConfigItem, updateConfigContent } from '../../../../../../../api/config'
  import { getConfigEditParams } from '../../../../../../../utils/config'
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config'

  const props = defineProps<{
    bkBizId: string,
    appId: number
  }>()

  const emits = defineEmits(['confirm'])

  const slideShow = ref(false)
  const configForm = ref<IConfigEditParams>(getConfigEditParams())
  const fileUploading = ref(false)
  const pending = ref(false)
  const content = ref<IFileConfigContentSummary|string>('')
  const formRef = ref()

  const handleOpenCreateConfig = () => {
    slideShow.value = true
    configForm.value = getConfigEditParams()
  }

  const handleFormChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary|string) => {
    configForm.value = data
    content.value = configContent
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
      await createServiceConfigItem(props.appId, props.bkBizId, params)
      emits('confirm')
      Message({
        theme: 'success',
        message: '新建配置项成功'
      })
    }catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const close = () => {
    slideShow.value = false
  }

</script>
<template>
  <section class="create-config-btn">
    <bk-button outline theme="primary" @click="handleOpenCreateConfig">新增配置文件</bk-button>
    <bk-sideslider
      width="640"
      title="新增配置文件"
      :is-show="slideShow"
      :before-close="close">
      <ConfigForm
        ref="formRef"
        class="config-form-wrapper"
        v-model:fileUploading="fileUploading"
        :config="configForm"
        :content="content"
        :editable="true"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        @change="handleFormChange"/>
      <section class="action-btns">
        <bk-button theme="primary" :loading="pending" :disabled="fileUploading" @click="handleSubmit">保存</bk-button>
        <bk-button @click="close">取消</bk-button>
      </section>
    </bk-sideslider>
  </section>
</template>
<style lang="scss" scoped>
  .config-form-wrapper {
    padding: 20px 40px;
    height: calc(100vh - 101px);
    overflow: auto;
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
