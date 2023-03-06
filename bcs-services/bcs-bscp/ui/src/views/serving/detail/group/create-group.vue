<script setup lang="ts">
  import { ref } from 'vue'

  const props = defineProps<{
    show: boolean
  }>()

  const emits = defineEmits(['update:show'])

  const pending = ref(false)


  const handleConfirm = () => {
    handleClose()
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    title="创建分组"
    :is-show="props.show"
    :width="960"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
      <bk-alert theme="info" title="同分类下的不同分组之间需要保证没有交集，否则客户端将产生配置错误的风险！" />
      <div class="config-wrapper">
        <bk-form class="form-content-area">
          <bk-form-item label="分组名称" required>
            <bk-input></bk-input>
          </bk-form-item>
          <bk-form-item label="分组标签" required>
            <bk-select></bk-select>
          </bk-form-item>
          <bk-form-item label="调试用分组" required>
            <bk-switcher></bk-switcher>
            <p>启用调试用分组后，仅可使用 UID 作为分组规则，且配置版本将不跟随主线</p>
          </bk-form-item>
          <bk-form-item label="分组规则" required>
            <div class="rule-config">
              <bk-input style="width: 80px;"></bk-input>
              <bk-select style="width: 72px;"></bk-select>
              <bk-input style="width: 120px;"></bk-input>
            </div>
          </bk-form-item>
        </bk-form>
        <div class="group-intersection-detect">
          <h4 class="title">分组交集检测</h4>
          <bk-table class="rule-table" :border="['outer']">
            <bk-table-column label="分组名称"></bk-table-column>
            <bk-table-column label="分组规则"></bk-table-column>
          </bk-table>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">提交</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .config-wrapper {
    display: flex;
    height: 530px;
  }
  .form-content-area {
    padding: 24px 24px 24px 0;
    width: 50%;
    height: 100%;
    overflow: auto;
  }
  .rule-config {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
  .group-intersection-detect {
    padding: 16px;
    width: 50%;
    background: #f5f7fa;
    .title {
      margin: 0;
    }
    .rule-table {
      margin-top: 16px;
    }
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .bk-modal-wrapper.bk-dialog-wrapper .bk-dialog-header {
    padding-bottom: 20px;
  }
</style>
