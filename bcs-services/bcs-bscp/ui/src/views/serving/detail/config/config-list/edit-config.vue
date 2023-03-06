<script setup lang="ts">
  import { defineProps, defineEmits, ref, watch } from 'vue'
  import { cloneDeep } from 'lodash'
  import ConfigForm from './config-form.vue'
  import { getConfigItemDetail, getConfigContent, updateServingConfigItem } from '../../../../../api/config'
  import { IServingEditParams } from '../../../../../types'

  const getDefaultConfig = () => {
    return {
      biz_id: props.bkBizId,
      app_id: props.appId,
      name: '',
      path: '',
      file_type: 'text',
      file_mode: 'unix',
      user: '',
      user_group: 'root',
      privilege: '',
    }
  }

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    releaseId: number|null
    configId: number,
    show: Boolean
  }>()

  const emit = defineEmits(['update:show', 'confirm'])

  const configDetailLoading = ref(true)
  const config = ref<IServingEditParams>(getDefaultConfig())
  const content = ref('')

  watch(
    () => props.show,
    (val) => {
      if (val) {
        getConfigDetail()
      }
    }
  )

  const getConfigDetail = async() => {
    try {
      configDetailLoading.value = true
      const detail = await getConfigItemDetail(props.configId, props.bkBizId, props.appId)
      // @ts-ignore
      const { name, path, file_type, permission } = detail.config_item.spec
      config.value = { id: props.configId, biz_id: props.bkBizId, app_id: props.appId, name, file_type, path, ...permission }

      // @ts-ignore
      const signature = detail.content.spec.signature
      const configContent = <object | string> await getConfigContent(props.bkBizId, props.appId, signature)
      content.value = configContent as string
    } catch (e) {
      console.error(e)
    } finally {
      configDetailLoading.value = false
    }
    // const { id, spec } = config
    // const { privilege, user } = permission
    // activeConfig.value = { 
    //   id,
    //   biz_id: props.bkBizId,
    //   app_id: props.appId,
    //   name,
    //   file_type,
    //   file_mode,
    //   path,
    //   user,
    //   user_group,
    //   privilege
    // }
  }

  const submitConfig = (data: IServingEditParams) => {
    return updateServingConfigItem(data)
  }

  const close = () => {
    emit('update:show', false)
  }
</script>
<template>
    <bk-sideslider
      width="640"
      title="编辑配置项"
      :is-show="props.show"
      :before-close="close">
        <bk-loading :loading="configDetailLoading" style="height: 100%;">
          <ConfigForm
            v-if="!configDetailLoading"
            :config="config"
            :content="content"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            :submit-fn="submitConfig"
            @confirm="$emit('confirm')"
            @cancel="close" />
        </bk-loading>
    </bk-sideslider>
</template>
<style lang="scss" scoped>
  :deep(.bk-modal-content) {
    height: 100%;
  }
</style>
