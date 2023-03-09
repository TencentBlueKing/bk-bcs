<script setup lang="ts">
  import { computed } from 'vue'
  import Diff from '../../../../../components/diff/index.vue'
  import { IConfigItem } from '../../../../../../types/config'

  const emits = defineEmits(['update:show'])

  const props = defineProps<{
    show: boolean,
    versionName: string,
    config: IConfigItem
  }>()

  const title = computed(() => {
    return `配置项对比 - ${props.config?.spec?.name}`
  })

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    :title="title"
    ext-cls="version-compare-dialog"
    :width="1200"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @closed="handleClose">
      <div class="diff-content-wrapper">
        <diff :panelName="props.versionName" type="file"></diff>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <bk-button @click="handleClose">关闭</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .diff-content-wrapper {
    padding-bottom: 20px;
    height: 580px;
  }
</style>