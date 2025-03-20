<template>
  <span v-if="flagsMap.BKAI">
    <!-- AI小鲸对话框 -->
    <Assistant
      :is-show.sync="isShowAssistant"
      :loading="loading"
      :messages="messages"
      :position-limit="positionLimit"
      :prompts="prompts"
      :start-position="startPosition"
      :size-limit="sizeLimit"
      enable-popup
      @clear="handleClear"
      @close="handleClose"
      @send="handleSend"
      @stop="handleStop" />
  </span>
</template>
<script setup lang="ts">
import { debounce } from 'lodash';
import { ref } from 'vue';

import Assistant, { ChatHelper, IMessage, ISendData, IStartPosition, MessageStatus, RoleType } from '@blueking/ai-blueking/vue2';

import useAssistantStore from './use-assistant-store';

import '@blueking/ai-blueking/dist/vue2/style.css';
import { BCS_UI_PREFIX } from '@/common/constant';
import { Preset } from '@/components/assistant/use-assistant-store';
import { useAppData } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';


interface ISendConfig extends ISendData {
  preset: Preset
}

const {
  isShowAssistant,
  toggleAssistant,
} = useAssistantStore();

// 特性开关
const { flagsMap } = useAppData();

// AI小鲸
const streamID = ref(1);
const loading = ref(false);
const messages = ref<IMessage[]>([]);
const prompts = ref([]);
const positionLimit = ref({
  top: 0,
  bottom: 0,
  left: 0,
  right: 0,
});
const sizeLimit = ref({
  height: 460,
  width: 720,
});
const startPosition = ref<IStartPosition>({
  right: 24,
  top: window.innerHeight - sizeLimit.value.height - 10,
  bottom: 10,
  left: window.innerWidth - sizeLimit.value.width - 24,
});
// 清空消息
const handleClear = () => {
  messages.value.splice(0);
};

// 聊天开始
const handleStart = () => {
  loading.value = true;
  messages.value.push({
    role: RoleType.Assistant,
    content: $i18n.t('blueking.aiScriptsAssistant.loading'),
    status: MessageStatus.Loading,
  });
};

// 接收消息
const handleReceiveMessage = (msg: string, id: number | string, cover?: boolean) => {
  const currentMessage = messages.value.at(-1);
  if (!currentMessage) return;

  if (currentMessage.status === 'loading') {
    // 如果是loading状态，直接覆盖
    currentMessage.content = msg;
    currentMessage.status = MessageStatus.Success;
  } else if (currentMessage.status === 'success') {
    // 如果是后续消息，就追加消息
    currentMessage.content = cover ? msg : currentMessage.content + msg;
  }
};

// 聊天结束
const handleEnd = () => {
  loading.value = false;
  const currentMessage = messages.value.at(-1);
  if (!currentMessage) return;
  // loading 情况下终止
  if (currentMessage.status === 'loading') {
    currentMessage.content = $i18n.t('blueking.aiScriptsAssistant.breakLoading');
    currentMessage.status = MessageStatus.Error;
  }
};

// 终止聊天
const handleStop = async () => {
  await chatHelper.stop(streamID.value);
};

// 错误处理
const handleError = (msg: string) => {
  const currentMessage = messages.value.at(-1);
  if (!currentMessage) return;

  currentMessage.status = MessageStatus.Error;
  currentMessage.content = msg;
  loading.value = false;
};
const chatHelper = new ChatHelper(`${BCS_UI_PREFIX}/assistant`, handleStart, handleReceiveMessage, handleEnd, handleError, messages.value);
// 发送消息
const handleSend = async (args: ISendConfig) => {
  if (!flagsMap.value.BKAI) return;

  // 记录当前消息记录
  const chatHistory = [...messages.value];

  // 添加用户消息
  messages.value.push({
    role: RoleType.User,
    content: args.content,
    cite: args.cite,
  });
  // 根据参数构造输入内容
  // eslint-disable-next-line no-nested-ternary
  const input = args.prompt
    ? args.prompt                           // 如果有 prompt，直接使用
    : args.cite
      ? `${args.content}: ${args.cite}`     // 如果有 cite，拼接 content 和 cite
      : args.content;                       // 否则只使用 content
  // ai 消息，id是唯一标识当前流，调用 chatHelper.stop 的时候需要传入
  chatHelper.stream({
    inputs: {
      input,
      chat_history: chatHistory,
      preset: args.preset || 'QA',
    },
  }, streamID.value);
};
// 发送消息防抖(外部调用)
const handleSendMsg = debounce((msg: string, pre: Preset = 'QA') => {
  handleSend({ content: msg, preset: pre });
}, 1000);

// 关闭对话框
const handleClose = () => {
  toggleAssistant(false);
};
// 快捷prompt(暂时不启用改功能)
// const handleChoosePrompt = (prompt) => {
//   console.log(prompt);
// };

defineExpose({
  handleSendMsg,
});
</script>
<style lang="postcss">
.tippy-tooltip.ai-assistant-theme {
  padding: 0!important;
  box-shadow: unset !important;
  background: transparent;
}

.ai-modal, .ai-blueking-render-popup {
  z-index: 9999 !important;
}
</style>
