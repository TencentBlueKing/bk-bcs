<script setup lang="ts">
  import { ref } from 'vue'
  import groupEditForm from './components/group-edit-form.vue';

  const props = defineProps<{
    show: boolean
  }>()

  const pending = ref(false)

  const handleClose = () => {}

  const handleConfirm = () => {}
</script>
<template>
  <bk-dialog
  title="调整分组上线"
  :width="640"
  ext-cls="modify-group-dialog"
  :is-show="props.show"
  :esc-close="false"
  :quick-close="false"
  :is-loading="pending"
  @closed="handleClose"
  @confirm="handleConfirm">
    <!-- <group-edit-form></group-edit-form> -->
  </bk-dialog>
</template>