<script lang="ts" setup>
  import { ref, computed } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Warn } from 'bkui-vue/lib/icon';
  import { Message } from 'bkui-vue';
  import { useGlobalStore } from '../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../store/template'
  import { moveOutTemplateFromPackage } from '../../../../../../../api/template'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
    id: number;
    name: string;
  }>()

  const emits = defineEmits(['update:show', 'movedOut'])

  const selectedPkgs = ref<number[]>([])
  const pending = ref(false)

  const handleConfirm = async () => {
    if (selectedPkgs.value.length === 0) {
      Message({
        theme: 'warn',
        message: '请选择套餐'
      })
      return
    }
    try {
      pending.value = true
      await moveOutTemplateFromPackage(spaceId.value, currentTemplateSpace.value, [props.id], selectedPkgs.value)
      emits('movedOut')
      close()
      Message({
        theme: 'success',
        message: '添加配置项成功'
      })
    } catch (e) {
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
