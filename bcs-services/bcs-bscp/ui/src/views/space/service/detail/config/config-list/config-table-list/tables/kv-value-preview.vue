<template>
  <div class="kv-value-preview">
    <div ref="valueRef" :class="['value-wrapper', { expanded: isExpanded }]">{{ props.value }}</div>
    <div class="operate-btns">
      <template v-if="showExpandOrFoldBtn">
        <bk-button v-if="isExpanded" text theme="primary" @click="isExpanded = false">{{ $t('收起') }}</bk-button>
        <bk-button v-else text theme="primary" @click="isExpanded = true">{{ $t('展开') }}</bk-button>
      </template>
      <bk-button v-if="showViewAllBtn" text theme="primary" @click="emits('viewAll')">
        {{ $t('查看完整配置') }}
      </bk-button>
    </div>
    <Copy class="copy-icon" @click="handleCopyText" />
  </div>
</template>
<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import Message from 'bkui-vue/lib/message';
  import { Copy } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';
  import { copyToClipBoard } from '../../../../../../../../utils/index';

  const { t } = useI18n();

  const emits = defineEmits(['viewAll']);

  const props = defineProps<{
    value: string;
  }>();

  const isExpanded = ref(false); // 是否已展开
  const scrollHeight = ref(0); // value高度
  const valueRef = ref();

  const showExpandOrFoldBtn = computed(() => scrollHeight.value > 100); // value内容高度超出5行时（5 * 20），显示展开/收起按钮
  const showViewAllBtn = computed(() => scrollHeight.value > 200 && isExpanded.value); // 内容如果超出10行，展开后，显示查看全部按钮

  onMounted(() => {
    if (valueRef.value) {
      scrollHeight.value = valueRef.value.scrollHeight;
    }
  });

  const handleCopyText = () => {
    copyToClipBoard(props.value);
    Message({
      theme: 'success',
      message: t('配置项值已复制'),
    });
  };
</script>
<style lang="scss" scoped>
  .kv-value-preview {
    position: relative;
    padding: 10px 20px 10px 0;
  }
  .value-wrapper {
    line-height: 20px;
    max-height: 100px; // 默认最多显示5行
    overflow: hidden;
    word-break: break-word;
    white-space: pre-wrap;
    &.expanded {
      max-height: 200px; // 展开后，最多显示10行
    }
  }
  .operate-btns {
    line-height: 20px;
    .bk-button:not(:first-child) {
      margin-left: 8px;
    }
  }
  .copy-icon {
    position: absolute;
    top: 50%;
    right: 0;
    transform: translateY(-50%);
    color: #979ba5;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
    }
  }
</style>
