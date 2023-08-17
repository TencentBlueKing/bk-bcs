<script lang="ts" setup>
  import { ref, computed } from 'vue'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';

  const props = defineProps<{
    show: boolean;
    value: ITemplateConfigItem[];
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
    ext-cls="move-out-configs-dialog"
    title="批量移出当前套餐"
    confirm-text="确定移出"
    :width="480"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <div class="selected-mark">已选 <span class="num">{{ props.value.length }}</span> 个配置项</div>
    <p class="tips">以下服务配置的未命名版本引用目标套餐的内容也将更新</p>
    <bk-table>
      <bk-table-column label="所在模板套餐"></bk-table-column>
      <bk-table-column label="使用此套餐的服务"></bk-table-column>
    </bk-table>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .selected-mark {
    display: inline-block;
    margin-bottom: 16px;
    padding: 0 12px;
    height: 32px;
    line-height: 32px;
    border-radius: 16px;
    font-size: 12px;
    color: #63656e;
    background: #f0f1f5;
    .num {
      color: #3a84ff;
    }
  }
</style>
