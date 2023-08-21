<script lang="ts" setup>
  import { ref, watch } from 'vue'
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

  watch(() => props.show, val => {
    if (val) {
      selectedPkgs.value = []
    }
  })

  const handleConfirm = async () => {
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
    header-align="center"
    footer-align="center"
    :title="`确认将配置项【${name}】移出套餐?`"
    :width="600"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
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
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" :disabled="selectedPkgs.length === 0" @click="handleConfirm">确认移出</bk-button>
        <bk-button @click="close">取消</bk-button>
      </div>
    </template>
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
  .actions-wrapper {
    padding-bottom: 20px;
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
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
