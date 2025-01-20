<template>
  <div>
    <div class="flex flex-col items-center justify-center">
      <i
        :class="[
          'bk-icon icon-exclamation',
          'flex items-center justify-center rounded-full',
          ' text-[26px] bg-[#ffe8c3] text-[#ff9c01] leading-[42px] w-[42px] h-[42px]'
        ]">
      </i>
      <div class="text-[#313238] text-[20px] mt-[20px] leading-[32px]">
        {{ title }}
      </div>
    </div>
    <div class="flex items-center justify-center leading-[22px] mt-[8px]">
      {{ subTitle }}
    </div>
    <slot></slot>
    <div class="flex items-center justify-center mt-[24px]">
      <bk-button
        theme="danger"
        class="min-w-[88px]"
        :loading="isLoading"
        @click="handleConfirm">{{ $t('generic.button.confirm') }}</bk-button>
      <bk-button class="min-w-[88px]" @click="handleCancel">{{ $t('generic.button.cancel') }}</bk-button>
    </div>
  </div>
</template>
<script setup lang="ts">

defineProps({
  title: {
    type: String,
    default: '',
  },
  subTitle: {
    type: String,
    default: '',
  },
  isLoading: {
    type: Boolean,
    default: false,
  },
});
const emits = defineEmits(['cancel', 'confirm']);

const handleConfirm = async () => {
  emits('confirm');
};
const handleCancel = () => {
  emits('cancel');
};

</script>
