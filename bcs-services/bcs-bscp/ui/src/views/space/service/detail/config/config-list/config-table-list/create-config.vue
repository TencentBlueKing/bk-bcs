<script setup lang="ts">
  import { ref } from 'vue'
  import ConfigForm from './config-form.vue'
  import { createServiceConfigItem } from '../../../../../../../api/config'
  import { IAppEditParams } from '../../../../../../../../types/app'


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
    appId: number
  }>()

  defineEmits(['confirm'])

  const slideShow = ref(false)
  const setting = ref(getDefaultConfig())
  const content = ref('')

  const handleCreateConfig = () => {
    slideShow.value = true
    setting.value = getDefaultConfig()
  }

  const submitConfig = (params: IAppEditParams) => {
    return createServiceConfigItem(params)
  }

  const close = () => {
    slideShow.value = false
  }

</script>
<template>
  <section class="create-config-btn">
    <bk-button outline theme="primary" @click="handleCreateConfig">新增配置文件</bk-button>
    <bk-sideslider
      width="640"
      title="新增配置文件"
      :is-show="slideShow"
      :before-close="close">
      <ConfigForm
        :config="setting"
        :content="content"
        :editable="true"
        :bk-biz-id="props.bkBizId"
        :app-id="props.appId"
        :submit-fn="submitConfig"
        @confirm="$emit('confirm')"
        @cancel="close" />
    </bk-sideslider>
  </section>
</template>
<style lang="scss" scoped>
  :deep(.bk-modal-content) {
    height: 100%;
  }
</style>