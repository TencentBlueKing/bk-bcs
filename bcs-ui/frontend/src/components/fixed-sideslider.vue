<template>
  <bcs-sideslider
    :is-show.sync="localValue"
    :title="title"
    :quick-close="quickClose"
    :width="width"
    :before-close="handleBefore"
    @hidden="handleCancel">
    <template #content>
      <div class="content" ref="contentDom">
        <slot></slot>
      </div>
      <slot name="footer">
        <div
          :class="[
            'h-[48px] flex items-center px-[24px] bg-[#fff] z-[100] w-full sticky bottom-0',
            { 'border-t-1 shadow-[0_-2px_4px_0_rgb(0_0_0_/_5%)]': isOverflow },
          ]">
          <bcs-button :loading="btnLoading" theme="primary" @click="handleConfirm">
            {{ okText }}
          </bcs-button>
          <bcs-button :loading="btnLoading" @click="handleCancel">{{ cancelText }}</bcs-button>
        </div>
      </slot>
    </template>
  </bcs-sideslider>
</template>
<script lang="ts" setup>
import { ref, watch } from 'vue';

import i18n from '@/i18n/i18n-setup';

interface IProps {
  isShow: boolean
  title: string;
  quickClose: boolean;
  width: number
  okText?: string
  cancelText?: string
  btnLoading?: boolean
}

const props = withDefaults(defineProps<IProps>(), {
  isShow: false,
  title: '',
  quickClose: true,
  width: 400,
  okText: i18n.t('generic.button.confirm'),
  cancelText: i18n.t('generic.button.cancel'),
  btnLoading: false,
});
const emit = defineEmits(['confirm', 'cancel']);

const localValue = ref(props.isShow);

const contentDom = ref<HTMLElement|null>(null);
const isOverflow = ref(false);
let observer;

function handleConfirm() {
  emit('confirm');
}
function handleCancel() {
  localValue.value = false;
  isOverflow.value = false;
  handleBefore();
  emit('cancel');
}
// 关闭侧滑 内容会被销毁 取消监听
function handleBefore() {
  if (observer && contentDom.value) {
    // 取消监听
    observer.unobserve(contentDom.value);
  }
  return Promise.resolve(true);
}

watch(() => props.isShow, (val) => {
  localValue.value = val;
});

watch(contentDom, () => {
  if (!contentDom.value) return;
  // 实时监听 contentDom 的高度是否超过 侧滑内容高度 控制 footer 上边框显示/隐藏
  observer = new ResizeObserver((entries) => {
    const entry = entries[0];
    const contentHeight = entry.contentRect.height;
    const viewportHeight = window.innerHeight;

    // 检查内容高度是否超过
    const isOver = contentHeight > (viewportHeight - 52 - 48);

    // 可以根据需要添加样式或逻辑处理
    if (isOver) {
      // 内容超过，可能需要添加滚动条或其他处理
      isOverflow.value = true;
    } else {
      isOverflow.value = false;
    }
  });
  // 开始监听
  observer.observe(contentDom.value);
});
</script>
<style lang="postcss" scoped>
:deep(.bk-sideslider-content) {
  height: calc(100vh - 52px);
  overflow: auto;
}
</style>
