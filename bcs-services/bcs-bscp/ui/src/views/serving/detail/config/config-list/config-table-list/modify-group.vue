<script setup lang="ts">
  import { ref } from 'vue'
  import CategoryGroupSelect from '../../../group/components/category-group-select.vue';

  const props = defineProps<{
    bkBizId: string,
    appId: number,
    releaseId: number
  }>()

  const showDialog = ref(false)
  const pending = ref(false)
  const groups = ref([])


  const handleConfirm = () => {
    showDialog.value = false
  }

  const handleClose = () => {
    showDialog.value = false
  }

</script>
<template>
  <section class="modify-group">
    <bk-button theme="primary" @click="showDialog = true">调整分组上线</bk-button>
    <bk-dialog
      title="调整分组上线"
      :width="480"
      ext-cls="modify-group-dialog"
      :is-show="showDialog"
      :esc-close="false"
      :quick-close="false"
      :is-loading="pending"
      @closed="handleClose"
      @confirm="handleConfirm">
      <bk-form class="group-edit-form" form-type="vertical">
        <bk-form-item label="当前分类">
          <span class="category-name">地区</span>
        </bk-form-item>
        <bk-form-item label="上线分组" required>
          <CategoryGroupSelect size="small" :app-id="props.appId" :multiple="true" :value="groups" />
        </bk-form-item>
      </bk-form>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">确定上线</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
    </bk-dialog>
  </section>
</template>
<style lang="scss" scoped>
  .group-edit-form {
    padding-bottom: 40px;
  }
  :deep(.bk-form-label) {
    font-size: 12px;
  }
  .category-name {
    font-size: 12px;
    color: #000000;
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .modify-group-dialog.bk-dialog-wrapper .bk-dialog-header {
    padding-bottom: 20px;
  }
</style>