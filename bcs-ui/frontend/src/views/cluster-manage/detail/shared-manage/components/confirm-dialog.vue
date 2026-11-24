<template>
  <!-- 共享管理通用二次确认弹窗：图标 + 描述 + 可选命名空间明细列表 + 操作按钮 -->
  <bcs-dialog
    class="shared-confirm-dialog"
    :value="value"
    :width="width"
    :close-icon="false"
    :draggable="false"
    header-position="left"
    @input="handleInput">
    <template #header>
      <!-- 对齐原型 dialog header：橙色感叹号图标 + 标题 -->
      <div class="flex items-center gap-[10px] !text-[18px] leading-[26px] text-[#313238]">
        <i class="w-[24px] h-[24px] rounded-full bg-[#FF9C01] text-white not-italic flex items-center justify-center text-[14px] font-semibold shrink-0">!</i>
        <span>{{ title }}</span>
      </div>
    </template>
    <p class="text-[12px] text-[#63656E] leading-[20px]">
      <slot>{{ desc }}</slot>
    </p>
    <!-- 命名空间明细（可选，对齐原型 .dlist：带边框圆角容器 + 行分隔线） -->
    <div
      v-if="namespaces.length"
      class="mt-[8px] max-h-[160px] overflow-y-auto border border-[#F0F1F5] rounded-[2px]">
      <div
        v-for="(ns, index) in namespaces"
        :key="`${ns.name}-${index}`"
        class="ns-item text-[12px] flex items-center justify-between gap-[12px] py-[6px] px-[10px] leading-[20px] border-b border-[#F0F1F5] last:border-b-0">
        <span class="truncate text-[#313238]">{{ ns.name }}</span>
        <span class="flex items-center gap-[6px] shrink-0">
          <bcs-tag v-if="ns.tag" theme="warning" class="!h-[22px] !leading-[22px]">{{ ns.tag }}</bcs-tag>
          <span class="text-[#979BA5]">{{ ns.project }}</span>
        </span>
      </div>
    </div>
    <!-- 底部附加内容（可选，如回收时的 ResourceQuota 勾选） -->
    <slot name="append"></slot>
    <template #footer>
      <bcs-button theme="danger" @click="handleConfirm">
        {{ confirmText }}
      </bcs-button>
      <bcs-button v-if="secondaryText" @click="$emit('secondary')">
        {{ secondaryText }}
      </bcs-button>
      <bcs-button @click="handleCancel">
        {{ $t('generic.button.cancel') }}
      </bcs-button>
    </template>
  </bcs-dialog>
</template>

<script lang="ts" setup>
// 共享管理二次确认弹窗（关闭共享 / 回收 / 移出可见范围共用）
interface INsItem {
  name: string
  project: string
  // 可选状态标签（如 创建审批中）
  tag?: string
}

defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '',
  },
  desc: {
    type: String,
    default: '',
  },
  width: {
    type: Number,
    default: 560,
  },
  confirmText: {
    type: String,
    default: '',
  },
  secondaryText: {
    type: String,
    default: '',
  },
  namespaces: {
    type: Array as () => INsItem[],
    default: () => [],
  },
});

const emits = defineEmits(['input', 'confirm', 'secondary', 'cancel']);

const handleInput = (value: boolean) => {
  emits('input', value);
};
const handleConfirm = () => {
  emits('confirm');
};
const handleCancel = () => {
  emits('input', false);
  emits('cancel');
};
</script>

<style lang="postcss">
/* 弹窗垂直居中：bk-dialog 默认 top: 100px 固定定位，改为 flex 容器居中
   （弹窗经 transfer-dom 挂载到 body，样式需全局作用域，通过 .shared-confirm-dialog 限定范围） */
.shared-confirm-dialog .bk-dialog-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.shared-confirm-dialog .bk-dialog {
  position: relative !important;
  top: 0 !important;

  .bk-dialog-tool {
    display: none;
  }
}

/* header 自定义内容（图标 + 标题）下移左右内边距，对齐原型 20px 24px 12px */
.shared-confirm-dialog .bk-dialog-header {
  padding: 20px 24px 12px;

  .bk-dialog-close {
    top: 18px;
  }
}
</style>
