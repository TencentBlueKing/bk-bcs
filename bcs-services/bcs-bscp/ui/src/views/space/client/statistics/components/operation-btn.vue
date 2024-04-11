<template>
  <div class="btn-wrap">
    <left-turn-line class="action-icon" @click="emits('refresh')" />
    <FilliscreenLine v-if="!isOpenFullScreen" class="action-icon" @click="handleOpenFullScreen" />
    <UnfullScreen v-else class="action-icon" @click="handleCloseFullScreen" />
  </div>
</template>

<script lang="ts" setup>
  import { LeftTurnLine, FilliscreenLine, UnfullScreen } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  defineProps<{
    isOpenFullScreen: boolean;
  }>();
  const emits = defineEmits(['refresh', 'toggleFullScreen']);

  // 打开全屏
  const handleOpenFullScreen = () => {
    emits('toggleFullScreen');
    window.addEventListener('keydown', handleEscClose, { once: true });
    BkMessage({
      theme: 'primary',
      message: t('按 Esc 即可退出全屏模式'),
    });
  };

  const handleCloseFullScreen = () => {
    emits('toggleFullScreen');
    window.removeEventListener('keydown', handleEscClose);
  };

  // Esc按键事件处理
  const handleEscClose = (event: KeyboardEvent) => {
    if (event.code === 'Escape') {
      emits('toggleFullScreen');
    }
  };
</script>

<style scoped lang="scss">
  .btn-wrap {
    display: flex;
    .action-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-left: 4px;
      width: 28px;
      height: 28px;
      background: #fafbfd;
      border: 1px solid #dcdee5;
      border-radius: 2px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
</style>
