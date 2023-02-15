<script setup lang="ts">
  import { defineProps, defineEmits ,ref } from 'vue'
  import ConfigItemEdit from './config-item-edit.vue'

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
    bkBizId: number,
    appId: number
  }>()

  defineEmits(['update'])

  const slideShow = ref(false)
  const setting = ref(getDefaultConfig())

  const handleCreateConfig = () => {
    slideShow.value = true
    setting.value = getDefaultConfig()
  }

</script>
<template>
  <section class="create-config-btn">
    <bk-button style="margin: 16px 0;" outline theme="primary" @click="handleCreateConfig">新增配置项</bk-button>
    <ConfigItemEdit
      v-model:show="slideShow"
      :config="setting"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      @confirm="$emit('update')" />
  </section>
</template>