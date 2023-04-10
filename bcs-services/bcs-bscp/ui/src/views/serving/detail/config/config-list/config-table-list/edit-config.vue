<script setup lang="ts">
  import { ref, watch, computed } from 'vue'
  import { storeToRefs } from 'pinia'
  import ConfigForm from './config-form.vue'
  import { getConfigItemDetail, getConfigContent, updateServingConfigItem } from '../../../../../../api/config'
  import { IFileConfigContentSummary } from '../../../../../../../types/config'
  import { IAppEditParams } from '../../../../../../../types/app'
  
  import { useConfigStore } from '../../../../../../store/config'

  const { versionData } = storeToRefs(useConfigStore())

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
    configId: number,
    show: Boolean
  }>()

  const emit = defineEmits(['update:show', 'confirm'])

  const configDetailLoading = ref(true)
  const config = ref<IAppEditParams>(getDefaultConfig())
  const content = ref<string|IFileConfigContentSummary>('')

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
      const detail = await getConfigItemDetail(props.configId, props.appId, params)
      const { name, path, file_type, permission } = detail.config_item.spec
      config.value = { id: props.configId, biz_id: props.bkBizId, app_id: props.appId, name, file_type, path, ...permission }
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

  const submitConfig = (data: IAppEditParams) => {
    return updateServingConfigItem(data)
  }

  const close = () => {
    emit('update:show', false)
  }
</script>
<template>
    <bk-sideslider
      width="640"
      :title="`${editable ? '编辑' : '查看'}配置项`"
      :is-show="props.show"
      :before-close="close">
        <bk-loading :loading="configDetailLoading" style="height: 100%;">
          <ConfigForm
            v-if="!configDetailLoading"
            :config="config"
            :content="content"
            :editable="editable"
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
