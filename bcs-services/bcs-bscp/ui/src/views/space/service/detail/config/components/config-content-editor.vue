<script setup lang="ts">
  import { ref } from 'vue';
  import BkMessage from 'bkui-vue/lib/message';
  import { InfoLine, FilliscreenLine, UnfullScreen } from 'bkui-vue/lib/icon'
  import ReadFileContent from './read-file-content.vue';
  import CodeEditor from '../../../../../../components/code-editor/index.vue'

  const props = withDefaults(defineProps<{
    content: string;
    editable: boolean;
  }>(), {
    editable: true
  })

  const emits = defineEmits(['change'])

  const isOpenFullScreen = ref(false)

  const handleFileReadComplete = (content: string) => {
    emits('change', content)
  }

  // 打开全屏
  const handleOpenFullScreen = () => {
    isOpenFullScreen.value = true
    window.addEventListener('keydown', handleEscClose, { once: true })
    BkMessage({
      theme: 'primary',
      message: '按 Esc 即可退出全屏模式'
    })
  }

  const handleCloseFullScreen = () => {
    isOpenFullScreen.value = false
    window.removeEventListener('keydown', handleEscClose)
  }

  // Esc按键事件处理
  const handleEscClose = (event: KeyboardEvent) => {
    if (event.code === 'Escape') {
      isOpenFullScreen.value = false
    }
  }

</script>
<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div :class="['config-content-editor', { fullscreen: isOpenFullScreen }]">
      <div class="editor-title">
        <div class="tips">
          <InfoLine class="info-icon" />
          仅支持大小不超过 40M
        </div>
        <div v-if="editable" class="btns">
            <ReadFileContent @completed="handleFileReadComplete" />
            <FilliscreenLine
              v-if="!isOpenFullScreen"
              v-bk-tooltips="{
                content: '全屏',
                placement: 'top',
                distance: 20
              }"
              @click="handleOpenFullScreen" />
            <UnfullScreen
              v-else
              v-bk-tooltips="{
                content: '退出全屏',
                placement: 'bottom',
                distance: 20
              }"
              @click="handleCloseFullScreen" />
        </div>
      </div>
      <div class="editor-content">
        <CodeEditor :model-value="props.content" :editable="editable" @update:model-value="emits('change', $event)" />
      </div>
    </div>
  </Teleport>
</template>
<style lang="scss" scoped>
.config-content-editor {
  height: 440px;
  &.fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
  }
  .editor-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    height: 40px;
    color: #979ba5;
    background: #2e2e2e;
    border-radius: 2px 2px 0 0;
    .tips {
      display: flex;
      align-items: center;
      font-size: 12px;
      .info-icon {
        margin-right: 4px;
      }
    }
    .btns {
      display: flex;
      align-items: center;
      & > span{
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .editor-content {
    height: calc(100% - 40px);
  }
}
</style>