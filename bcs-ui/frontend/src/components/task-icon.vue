<template>
  <span
    v-bk-tooltips="{
      content: desc,
      disabled: !desc,
      placement,
    }"
    :class="[
      'flex items-center justify-center rounded-sm',
      sizeMap[size],
      fontSizeMap[fontSize] || `text-[${fontSize}]`,
      clickable ? 'cursor-pointer hover:bg-[#4D4D4D]' : '',
    ]"
  >
    <i
      :style="{ color: StatusColor[status] || '#C4C6CC' }"
      :class="[icon, Icon.loading === icon || spin ? 'animate-spin-slow' : '']"
    >
    </i>
    <slot></slot>
  </span>
</template>
<script lang="ts" setup>
import { computed, PropType } from 'vue';

enum Icon {
  failed = 'bcs-icon bcs-icon-circle-wrong-filled',
  halfsuccess = 'bcs-icon bcs-icon-half-success-line',
  loading = 'bcs-icon bcs-icon-jiazai',
  success = 'bcs-icon bcs-icon-circle-correct-filled',
  terminate = 'bcs-icon bcs-icon-circle-terminate-filled',
  waiting = 'bcs-icon bcs-icon-waiting',
}
enum StatusColor {
  FAILED = '#DA4444',
  HALFSUCCESS = '#1CAB88',
  LOADING = '#3A84FF',
  SUCCESS = '#1CAB88',
  TERMINATE = '#DA4444',
  WAITING = '#979BA5',
}

const sizeMap = {
  default: 'w-[24px] h-[24px]',
  large: 'w-[32px] h-[32px]',
  none: '',
};

const fontSizeMap = {
  small: 'text-[10px]',
  default: 'text-[12px]',
  large: 'text-[14px]',
};

const props = defineProps({
  status: {
    type: String,
    default: '',
  },
  desc: {
    type: String,
    default: '',
  },
  clickable: {
    type: Boolean,
    default: false,
  },
  size: {
    type: String as PropType<'default' | 'large' | 'none'>,
    default: 'default',
  },
  color: {
    type: String,
    default: '',
  },
  fontSize: {
    type: String as PropType<keyof typeof fontSizeMap | string>,
    default: 'large',
  },
  placement: {
    type: String,
    default: 'bottom',
  },
  spin: {
    type: Boolean,
    default: false,
  },
});

const icon = computed(() => Icon[props.status?.toLocaleLowerCase() as keyof typeof Icon]);
</script>
