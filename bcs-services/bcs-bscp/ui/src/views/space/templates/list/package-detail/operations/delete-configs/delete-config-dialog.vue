<script lang="ts" setup>
  import { ref } from 'vue'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';

  const props = defineProps<{
    show: boolean;
    configs: ITemplateConfigItem[]
  }>()

  const emits = defineEmits(['update:show'])

  const pending = ref(false)

  const handleConfirm = () => {}

  const close = () => {
    emits('update:show', false)
  }
</script>
<template>
  <bk-dialog
    ext-cls="delete-configs-dialog"
    confirm-text="确认删除"
    header-align="center"
    footer-align="center"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">

  </bk-dialog>
</template>
<style lang="scss" scoped>
</style>
