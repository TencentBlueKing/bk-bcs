<template>
  <bcs-dialog
    theme="primary"
    header-position="left"
    :title="title"
    :width="1200"
    v-model="show"
    @cancel="handleCancel">
    <slot></slot>
  </bcs-dialog>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  isShow: {
    type: Boolean,
    default: false,
  },
});

const $emits = defineEmits(['cancel']);

const show = ref(false);

function handleCancel() {
  $emits('cancel', false);
}

watch(() => props.isShow, () => {
  show.value = props.isShow;
});
</script>
<style scoped lang="postcss">
:deep(.bk-dialog-content) {
  max-height: 80vh;
  .bk-dialog-body {
    overflow: auto;
    max-height: calc(80vh - 116px);
  }
}
</style>
