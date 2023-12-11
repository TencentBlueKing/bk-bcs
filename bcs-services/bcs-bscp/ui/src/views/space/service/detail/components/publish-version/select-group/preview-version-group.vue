<template>
  <div class="version-callapse-item" :key="previewGroup.id">
    <div class="version-header" @click="fold = !fold">
      <span class="arrow-icon">
        <AngleRight v-if="fold" />
        <AngleDown v-else />
      </span>
      <div :class="['version-type-marking', previewGroup.type]">
        【{{ TYPE_MAP[previewGroup.type as keyof typeof TYPE_MAP] }}】
      </div>
      <span v-if="previewGroup.type === 'modify'" class="name"> - {{ previewGroup.name }}</span>
      <span class="group-count-wrapper"
        >共 <span class="count">{{ previewGroup.children.length }}</span> 个分组</span>
      <bk-button
        v-if="previewGroup.id && previewGroup.id !== versionData.id"
        text
        class="diff-btn"
        theme="primary"
        @click.stop="emits('diff', previewGroup.id)"
      >
        版本对比
      </bk-button>
    </div>
    <div v-if="!fold" class="group-list">
      <div v-for="group in previewGroup.children" class="group-item" :key="group.id">
        <span class="node-name">{{ group.name }}</span>
        <span v-if="group.rules && group.rules.length > 0" class="split-line">|</span>
        <div class="rules">
          <div v-for="(rule, index) in group.rules" :key="index">
            <template v-if="index > 0"> & </template>
            <rule-tag class="tag-item" :rule="rule" />
          </div>
        </div>
        <span
          v-if="props.allowPreviewDelete && !props.disabled.includes(group.id)"
          class="del-icon"
          @click="emits('delete', group.id)"
        >
          <Del />
        </span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { Del, AngleDown, AngleRight } from 'bkui-vue/lib/icon';
import { IGroupPreviewItem } from '../../../../../../../../types/group';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../../store/config';
import RuleTag from '../../../../../groups/components/rule-tag.vue';
const versionStore = useConfigStore();
const { versionData } = storeToRefs(versionStore);
const props = defineProps<{
  allowPreviewDelete: boolean;
  previewGroup: IGroupPreviewItem;
  disabled: number[];
}>();

const emits = defineEmits(['diff', 'delete']);

const TYPE_MAP = {
  current: '当前版本',
  modify: '变更版本',
  plain: '首次上线',
};

const fold = ref(false);
</script>
<style lang="scss" scoped>
.version-callapse-item {
  margin-bottom: 16px;
  padding: 0 24px;
  .version-header {
    display: flex;
    align-items: center;
    margin-bottom: 4px;
    line-height: 24px;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    .arrow-icon {
      font-size: 18px;
      line-height: 1;
    }
    .version-type-marking {
      &.modify {
        color: #ff9c01;
      }
    }
  }
  .group-count-wrapper {
    margin-left: 16px;
    color: #979ba5;
    .count {
      color: #3a84ff;
      font-weight: 700;
    }
  }
  .diff-btn {
    margin-left: auto;
  }
  .group-item {
    display: flex;
    align-items: center;
    position: relative;
    margin-bottom: 2px;
    padding: 8px 30px 8px 16px;
    background: #ffffff;
    border-radius: 2px;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
    &:hover {
      background: #e1ecff;
      .del-icon {
        display: block;
      }
    }
    .node-name {
      font-size: 12px;
      line-height: 20px;
      color: #63656e;
    }
    .split-line {
      margin: 0 4px 0 16px;
      line-height: 16px;
      font-size: 12px;
      color: #979ba5;
    }
    .rules {
      display: flex;
      align-items: center;
      line-height: 16px;
      font-size: 12px;
      color: #979ba5;
      white-space: nowrap;
    }
    .del-icon {
      display: none;
      position: absolute;
      top: 12px;
      right: 9px;
      font-size: 14px;
      color: #939ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
}
</style>
