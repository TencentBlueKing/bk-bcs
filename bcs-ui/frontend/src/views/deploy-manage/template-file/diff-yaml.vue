<template>
  <div class="flex flex-col" ref="contentRef">
    <!-- 操作栏 -->
    <div
      :class="[
        'flex items-center px-[16px] h-[40px]',
        'border-b-[1px] border-solid border-[#000]',
        'text-[#B6B6B6] text-[12px] bg-[#2E2E2E]'
      ]">
      <span class="flex-1">
        <span class="inline-flex items-center leading-[22px] px-[8px] rounded-sm bg-[#293f6f] text-[#6ca4fa]">
          {{ $t('generic.title.diffVersion') }}
        </span>
        <span class="ml-[12px]">{{ originalVersion }}</span>
      </span>
      <span class="flex items-center justify-between flex-1 pl-[16px]">
        <div class="flex items-center gap-[5px]">
          <span class="inline-flex items-center leading-[22px] px-[8px] rounded-sm bg-[#204f33] text-[#44bb64]">
            {{ $t('generic.title.curVersion') }}
          </span>
          <span
            v-if="isDraft"
            class="inline-flex items-center leading-[22px] px-[8px] rounded-sm bg-[#204f33] text-[#44bb64]">
            {{ $t('templateFile.tag.draft') }}
          </span>
        </div>
        <i
          :class="[
            'hover:text-[#699df4] cursor-pointer',
            isFullscreen ? 'bcs-icon bcs-icon-zoom-out' : 'bcs-icon bcs-icon-enlarge'
          ]"
          @click="switchFullScreen">
        </i>
      </span>
    </div>
    <CodeEditor
      :value="value"
      :original="original"
      :options="{
        roundedSelection: false,
        scrollBeyondLastLine: false,
        renderLineHighlight: 'none'
      }"
      diff-editor
      readonly
      class="flex-1 !h-0"
      ref="diffEditorRef"
      @diff-stat="handleDiffStatChange">
    </CodeEditor>
    <!-- 状态栏 -->
    <div
      :class="[
        'flex items-center gap-[20px] px-[16px] h-[40px] bg-[#1D1D1D] shadow-[0_-1px_0_0_rgba(49,50,56,1)]',
        'border-t-[1px] border-solid border-[#000] justify-between',
      ]">
      <div class="flex items-center gap-[20px]">
        <div
          v-for="item in diffColorList"
          :key="item.type"
          class="flex items-center text-[12px]">
          <span class="flex w-[12px] h-[12px] mr-[1px]" :style="{ background: item.current }"></span>
          <span class="flex w-[12px] h-[12px] mr-[6px]" :style="{ background: item.origin }"></span>
          <span :style="{ color: item.text }">{{ item.name }}</span>
        </div>
      </div>
      <div class="flex items-center pl-[16px] pr-[8px] h-[32px] text-[#838385] text-[12px] rounded-sm">
        <span class="flex items-center">
          {{ current }}/{{ counts }}
        </span>
        <span class="flex items-center ml-2">
          <bcs-icon
            class="!text-[20px] cursor-pointer hover:text-[#3a84ff] bg-[#575757] ml-2 rounded-sm text-[#dddddf] p-[2px]"
            type="angle-up"
            @click="previousDiffChange" />
          <bcs-icon
            class="!text-[20px] cursor-pointer hover:text-[#3a84ff] bg-[#575757] ml-2 rounded-sm text-[#dddddf] p-[2px]"
            type="angle-down"
            @click="nextDiffChange" />
        </span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, onMounted, ref, watch } from 'vue';

import CodeEditor from '@/components/monaco-editor/new-editor.vue';
import $i18n from '@/i18n/i18n-setup';

interface Props {
  value?: string
  original?: string
  originalVersion?: string
  isDraft?: boolean
}

defineProps<Props>();

// 全屏
const contentRef = ref();
const isFullscreen = ref(false);
function handleFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement;
};
function switchFullScreen() {
  if (!document.fullscreenElement) {
    contentRef.value?.requestFullscreen();
  } else {
    document.exitFullscreen();
  }
};

// diff说明
const diffColorList = [
  {
    type: 'delete',
    name: $i18n.t('diff.delete'),
    current: '#d80e0b',
    origin: '#666666',
    text: '#d80e0b',
  },
  {
    type: 'change',
    name: $i18n.t('diff.change'),
    current: '#d80e0b',
    origin: '#576d29',
    text: '#B6B6B6',
  },
  {
    type: 'add',
    name: $i18n.t('diff.add'),
    current: '#666666',
    origin: '#576d29',
    text: '#576d29',
  },
];

// diff 统计
const current = ref(1);
const counts = ref(0);
const diffEditorRef = ref();
function handleDiffStatChange(stat) {
  counts.value = stat.changesCount;
};
function nextDiffChange() {
  if (!counts.value) return;
  diffEditorRef.value?.nextDiffChange();
  if ((current.value + 1) > counts.value) {
    current.value = 1;
  } else {
    current.value += 1;
  }
};
function previousDiffChange() {
  if (!counts.value) return;
  diffEditorRef.value?.previousDiffChange();
  if ((current.value - 1) < 1) {
    current.value = counts.value;
  } else {
    current.value -= 1;
  }
};

// 重置count
watch(counts, () => {
  if (counts.value) {
    current.value = 1;
  } else {
    current.value = 0;
  }
}, { immediate: true });

onMounted(() => {
  document.addEventListener('fullscreenchange', handleFullscreenChange);
});

onBeforeMount(() => {
  document.removeEventListener('fullscreenchange', handleFullscreenChange);
});
</script>
