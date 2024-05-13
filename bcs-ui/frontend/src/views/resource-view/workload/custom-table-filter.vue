<template>
  <bcs-popover
    theme="light navigation-message"
    trigger="click"
    placement="bottom"
    :on-hide="onHide"
    :on-show="onShow"
    ref="popoverRef">
    <span class="flex items-center">
      {{ label }}
      <i
        :class="[
          'bk-icon icon-funnel text-[14px]',
          !isShow ? 'text-[#c4c6cc]': '',
          searchValue ? '!text-[#3a84ff]' : ''
        ]">
      </i>
    </span>
    <template #content>
      <div class="p-[10px]">
        <bcs-input :value="searchValue" type="textarea" @change="handleInputChange"></bcs-input>
      </div>
      <div class="flex items-center h-[32px] bcs-border-top">
        <bcs-button
          class="flex-1"
          text
          size="small"
          @click="handleConfirm">{{ $t('generic.button.confirm') }}</bcs-button>
        <bcs-button
          class="flex-1"
          text
          size="small"
          @click="handleReset">{{ $t('dashboard.workload.editor.reset') }}</bcs-button>
      </div>
    </template>
  </bcs-popover>
</template>
<script setup lang="ts">
import { ref } from 'vue';

defineProps({
  label: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['confirm']);

const popoverRef = ref();
const isShow = ref(false);
const searchValue = ref('');
const tmpValue = ref('');

const handleInputChange = (value: string) => {
  tmpValue.value = value;
};

// const show = () => {
//   popoverRef.value.showHandler();
// };
const hide = () => {
  popoverRef.value.hideHandler();
};

const onHide = () => {
  isShow.value = false;
};

const onShow = () => {
  isShow.value = true;
};

const handleReset = () => {
  searchValue.value = '';
  tmpValue.value = '';
  hide();
  emits('confirm', searchValue.value);
};

const handleConfirm = () => {
  searchValue.value = tmpValue.value;
  hide();
  emits('confirm', searchValue.value);
};
</script>
