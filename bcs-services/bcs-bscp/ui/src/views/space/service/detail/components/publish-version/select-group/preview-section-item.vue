<template>
  <div class="version-callapse-item" :key="previewGroup.id">
    <div class="version-header" @click="fold = !fold">
      <span class="angle-icon">
        <AngleRight v-if="fold" />
        <AngleDown v-else />
      </span>
      <div :class="['version-type-marking', previewGroup.type]">
        【{{ TYPE_MAP[previewGroup.type as keyof typeof TYPE_MAP] }}】
      </div>
      <span v-if="previewGroup.type === 'modify'" class="release-name">
        {{ previewGroup.name }} <ArrowsRight class="arrow-icon" /> {{ versionData.spec.name }}
      </span>
      <span v-else-if="previewGroup.type === 'retain'" class="release-name">
        {{ previewGroup.name }}
      </span>
      <span v-if="!hasDefaultGroup" class="group-count-wrapper">
        {{ t('共') }}
        <span class="count">{{ previewGroup.children.length }}</span>
        {{ t('个分组') }}
      </span>
      <div class="action-btns">
        <template v-if="previewGroup.type === 'modify'">
          <bk-button
            v-if="props.releaseType === 'all' && !hasDefaultGroup"
            text
            theme="primary"
            @click.stop="handleDelete(previewGroup.children)">
            {{ t('排除此版本下所有分组') }}
          </bk-button>
          <bk-button text theme="primary" @click.stop="emits('diff', previewGroup.id)">
            {{ t('版本对比') }}
          </bk-button>
        </template>
        <bk-button
          v-if="previewGroup.type === 'retain'"
          text
          theme="primary"
          @click.stop="handleAdd(previewGroup.children)">
          {{ t('取消排除以下所有分组') }}
        </bk-button>
      </div>
    </div>
    <div v-if="!fold" class="group-list">
      <div v-for="group in previewGroup.children" class="group-item" :key="group.id">
        <span class="node-name">
          {{ group.name }}
        </span>
        <span v-if="group.rules && group.rules.length > 0" class="split-line">|</span>
        <div class="rules">
          <div v-for="(rule, index) in group.rules" :key="index">
            <template v-if="index > 0"> & </template>
            <rule-tag class="tag-item" :rule="rule" />
          </div>
        </div>
        <div class="group-operations">
          <span
            v-if="props.releaseType === 'select' && !props.releasedGroups.includes(group.id)"
            class="del-icon"
            @click="handleDelete([group])">
            <Del />
          </span>
          <template v-if="props.releaseType === 'all' && !hasDefaultGroup">
            <bk-button
              v-if="previewGroup.type === 'modify'"
              text
              size="small"
              theme="primary"
              @click.stop="handleDelete([group])">
              {{ t('排除此分组') }}
            </bk-button>
            <bk-button
              v-if="previewGroup.type === 'retain'"
              text
              size="small"
              theme="primary"
              @click.stop="handleAdd([group])">
              {{ t('取消排除') }}
            </bk-button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Del, AngleDown, AngleRight, ArrowsRight } from 'bkui-vue/lib/icon';
  import { IGroupPreviewItem, IGroupToPublish } from '../../../../../../../../types/group';
  import { storeToRefs } from 'pinia';
  import useConfigStore from '../../../../../../../store/config';
  import RuleTag from '../../../../../groups/components/rule-tag.vue';

  const versionStore = useConfigStore();
  const { versionData } = storeToRefs(versionStore);
  const { t } = useI18n();

  const props = defineProps<{
    releaseType: string;
    sectionType: 'diff' | 'exclude'; // 预览分组差异或排除分组
    previewGroup: IGroupPreviewItem;
    releasedGroups: number[];
    value: IGroupToPublish[];
  }>();

  const emits = defineEmits(['diff', 'change']);

  const TYPE_MAP = {
    plain: t('首次上线'),
    modify: t('变更版本'),
    retain: t('保留版本'),
  };

  const fold = ref(false);

  const hasDefaultGroup = computed(() => {
    return props.previewGroup.children.some((item) => item.id === 0);
  });

  // 排除
  const handleDelete = (groups: IGroupToPublish[]) => {
    const list = props.value.slice();
    groups.forEach((group) => {
      const index = list.findIndex((item) => item.id === group.id);
      if (index > -1) {
        list.splice(index, 1);
      }
    });
    emits('change', list);
  };

  // 取消排除
  const handleAdd = (groups: IGroupToPublish[]) => {
    const list = props.value.slice();
    groups.forEach((group) => {
      const index = list.findIndex((item) => item.id === group.id);
      if (index === -1) {
        list.push(group);
      }
    });
    emits('change', list);
  };
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
      .angle-icon {
        font-size: 18px;
        line-height: 1;
      }
      .version-type-marking {
        &.modify {
          color: #ff9c01;
        }
      }
      .release-name {
        display: inline-flex;
        align-items: center;
        .arrow-icon {
          font-size: 20px;
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
    .action-btns {
      display: flex;
      align-items: center;
      margin-left: auto;
      font-size: 12px;
      .bk-button {
        margin-right: 16px;
      }
    }
    .group-item {
      display: flex;
      align-items: center;
      position: relative;
      margin-bottom: 2px;
      padding: 8px 16px 8px 16px;
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
        overflow: hidden;
      }
      .group-operations {
        display: inline-flex;
        align-items: center;
        margin-left: auto;
        font-size: 12px;
      }
      .del-icon {
        display: none;
        // position: absolute;
        // top: 12px;
        // right: 9px;
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
