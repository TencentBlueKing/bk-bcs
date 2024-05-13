<template>
  <div class="text-diff-legend">
    <div class="color-wrapper">
      <div class="color-info">
        <div class="color-box red"></div>
        <div class="color-box gray"></div>
        <div class="color-text" style="color: #812c2eff">{{ t('删除') }}</div>
      </div>
      <div class="color-info">
        <div class="color-box red"></div>
        <div class="color-box green"></div>
        <div class="color-text" style="color: #aeaeaeff">{{ t('变化') }}</div>
      </div>
      <div class="color-info">
        <div class="color-box gray"></div>
        <div class="color-box green"></div>
        <div class="color-text" style="color: #6e963cff">{{ t('新增') }}</div>
      </div>
    </div>
    <div class="separator-wrapper">
      <div class="number">{{ diffNumber }}<span>/</span>{{ currentDiffNumber + permissionDiffNumber }}</div>
      <div class="line"></div>
      <div class="button">
        <angle-up fill="#C4C6CC" @click="previous" />
        <angle-down fill="#C4C6CC" @click="next" />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch, onBeforeUnmount } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { AngleDown, AngleUp } from 'bkui-vue/lib/icon';
  import * as monaco from 'monaco-editor';

  const { t } = useI18n();

  let contentNavigator: monaco.editor.IDiffNavigator;
  let permissionNavigator: monaco.editor.IDiffNavigator;
  const contentLineChange = ref<monaco.editor.ILineChange[]>([]);
  const contentDiffNumber = ref(0);
  const currentDiffNumber = ref(0);
  const diffNumber = computed(() => contentDiffNumber.value + props.permissionDiffNumber);

  const props = defineProps<{
    diffEditor: monaco.editor.IStandaloneDiffEditor;
    permissionEditor: monaco.editor.IStandaloneDiffEditor;
    permissionDiffNumber: number;
  }>();

  watch(
    () => props.diffEditor,
    () => {
      createNavigator();
      getContentDiffNumber();
    },
  );

  watch([() => contentDiffNumber.value, () => props.permissionDiffNumber], () => {
    getCurrentDiffIndex();
  });

  onBeforeUnmount(() => {
    if (contentNavigator) {
      contentNavigator.dispose();
    }
    if (permissionNavigator) {
      permissionNavigator.dispose();
    }
  });

  // 设置差异导航
  const createNavigator = () => {
    contentNavigator = monaco.editor.createDiffNavigator(props.diffEditor, {
      followsCaret: true,
      ignoreCharChanges: true,
    });
  };

  // 获取内容差异个数
  const getContentDiffNumber = () => {
    props.diffEditor.onDidUpdateDiff(() => {
      contentLineChange.value = props.diffEditor.getLineChanges() || [];
      contentDiffNumber.value = contentLineChange.value.length;
    });
  };

  // 获取当前差异行
  const getCurrentDiffIndex = () => {
    const position = props.diffEditor.getPosition() as monaco.Position;
    if (contentLineChange.value?.length === 0) {
      currentDiffNumber.value = 0;
      return;
    }
    contentLineChange.value.forEach((item: any, index: number) => {
      if (item.modifiedStartLineNumber <= position.lineNumber && item.modifiedEndLineNumber >= position.lineNumber) {
        currentDiffNumber.value = index + 1;
      }
    });
  };

  const previous = () => {
    contentNavigator.previous();
    getCurrentDiffIndex();
  };

  const next = () => {
    contentNavigator.next();
    getCurrentDiffIndex();
  };
</script>

<style scoped lang="scss">
  .text-diff-legend {
    display: flex;
    align-items: center;
    width: 1197.72px;
    height: 47.5px;
    background: #1d1d1d;
    box-shadow: 0 -1px 0 0 #313238;
    .color-wrapper {
      display: flex;
      justify-content: space-between;
      width: 205px;
      margin: 0 15px;
      .color-info {
        display: flex;
        align-items: center;
        .color-box {
          width: 12px;
          height: 12px;
          margin-right: 1px;
        }
        .color-text {
          margin-left: 6px;
        }
        .red {
          background: #702622;
        }
        .gray {
          background: #666666;
        }
        .green {
          background: #3d4d1f;
        }
      }
    }
    .separator-wrapper {
      display: flex;
      justify-content: space-evenly;
      align-items: center;
      width: 121px;
      height: 32px;
      background: #63656e;
      border-radius: 2px;
      .number {
        font-size: 12px;
        color: #c4c6cc;
      }
      .line {
        width: 1px;
        height: 16px;
        background: #c4c6cc;
      }
      .button {
        display: flex;
        align-items: center;
        font-size: 24px;
        span {
          border-radius: 50%;
          &:hover {
            background: rgba($color: #928e8e, $alpha: 0.3);
          }
        }
      }
    }
  }
</style>
