<template>
  <bcs-dialog
    :value="value"
    :width="width"
    :show-footer="false"
    @value-change="handleValueChange">
    <template #header>
      <div class="flex flex-col items-center justify-center">
        <span
          class="flex items-center justify-center text-[#ff9c01]
          bg-[#ffe8c3] text-[24px] rounded-full w-[42px] h-[42px]">
          <i class="bk-icon icon-exclamation"></i>
        </span>
        <span class="text-[#313238] text-[20px] leading-[32px] mt-[20px]">
          {{ title }}
        </span>
      </div>
    </template>
    <div class="bg-[#F5F6FA] rounded-sm py-[12px] px-[16px]">
      <div>{{ subTitle }}</div>
      <div
        class="text-[14px] leading-[22px] flex items-start"
        v-for="item, index in tipsArr"
        :key="index">
        <span
          class="flex items-center justify-center w-[14px] h-[14px]
        text-[#fff] bg-[#ff9c01] text-[12px] rounded-full mt-[4px]">
          <i class="bk-icon icon-exclamation"></i>
        </span>
        <span class="ml-[5px] flex-1">{{ item }}</span>
      </div>
    </div>
    <div class="flex items-center justify-center mt-[16px]">
      <bcs-button
        :theme="theme"
        :loading="loading"
        class="mr10 min-w-[88px]"
        @click="handleConfirm">
        {{ okText || $t('确定') }}
      </bcs-button>
      <bk-button :disabled="loading" class="min-w-[88px]" @click="handleCancel">{{ cancelText || $t('取消') }}</bk-button>
    </div>
  </bcs-dialog>
</template>
<script lang="ts">
import { computed, defineComponent, toRefs, ref, PropType } from 'vue';

export default defineComponent({
  name: 'ConfirmDialog',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '',
    },
    subTitle: {
      type: String,
      default: '',
    },
    tips: {
      type: [Array, String],
      default: '',
    },
    okText: {
      type: String,
      default: '',
    },
    cancelText: {
      type: String,
      default: '',
    },
    width: {
      type: Number,
      default: 580,
    },
    confirm: {
      type: Function as PropType<any>,
      default: () => (() => {}),
    },
    theme: {
      type: String,
      default: 'danger',
    },
  },
  setup(props, ctx) {
    const { tips, confirm } = toRefs(props);
    const tipsArr = computed(() => (Array.isArray(tips.value) ? tips.value : [tips.value]));

    const loading = ref(false);
    const handleValueChange = (v) => {
      ctx.emit('change', v);
    };
    const handleCancel = () => {
      ctx.emit('cancel');
      ctx.emit('change', false);
    };
    const handleConfirm = async () => {
      ctx.emit('confirm');
      if (typeof confirm.value !== 'function') return;
      loading.value = true;
      try {
        await confirm.value();
      } catch (error) {
        loading.value = false;
      }
      loading.value  = false;
      ctx.emit('change', false);
    };

    return {
      loading,
      tipsArr,
      handleValueChange,
      handleCancel,
      handleConfirm,
    };
  },
});
</script>
