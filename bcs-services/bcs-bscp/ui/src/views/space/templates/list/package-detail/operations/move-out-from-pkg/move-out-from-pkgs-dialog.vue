<script lang="ts" setup>
  import { ref, computed } from 'vue'
  import { Warn } from 'bkui-vue/lib/icon';

  const props = defineProps<{
    show: boolean;
    id: number;
    name: string;
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
    ext-cls="move-out-from-pkgs-dialog"
    confirm-text="确定移出"
    header-align="center"
    footer-align="center"
    :title="`确认将配置项【${name}】移出套餐?`"
    :width="600"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <bk-table>
      <bk-table-column type="selection" min-width="30" width="40" />
      <bk-table-column label="所在模板套餐"></bk-table-column>
      <bk-table-column label="使用此套餐的服务"></bk-table-column>
    </bk-table>
    <p class="tips">
      <Warn class="warn-icon" />
      移出后配置项将不存在任一套餐。你仍可在「全部配置项」或「未指定套餐」分类下找回。
    </p>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .tips {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #63656e;
    .warn-icon {
      margin-right: 4px;
      font-size: 14px;
      color: #ff9c05;
    }
  }
</style>
<style lang="scss">
  .move-out-from-pkgs-dialog.bk-modal-wrapper.bk-dialog-wrapper {
    .bk-modal-footer {
      padding: 32px 0 48px;
      background: #ffffff;
      border-top: none;
      .bk-button {
        min-width: 88px;
      }
    }
  }
</style>
