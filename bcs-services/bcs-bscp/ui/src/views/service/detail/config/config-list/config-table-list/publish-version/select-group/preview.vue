<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { IGroupTreeItem, IGroupItemInService } from '../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../types/config'
  import RuleTag from '../../../../../../../groups/components/rule-tag.vue';

  interface IGroupPreviewItem {
    id: number;
    name: string;
    type: String;
    children: IGroupTreeItem[]
  }

  // 将分组按照版本聚合
    const aggregateGroup = (groups: IGroupTreeItem[]) => {
      const list: IGroupPreviewItem[] = []
      const modifyVersions: IGroupPreviewItem[] = []
      const noVersions: IGroupPreviewItem[] = [{ id: 0, name: '无版本', type: 'plain', children: [] }]
      groups.forEach((group) => {
        const { release_id, release_name } = group
        if (release_id) {
          const version = modifyVersions.find(item => item.id === release_id)
          if (version) {
            if (!version.children.find(item => item.id === group.id)) {
              version.children.push(group)
            }
          } else {
            modifyVersions.push({ id: release_id, name: <string>release_name, type: 'modify', children: [group] })
          }
        } else {
          noVersions[0].children.push(group)
        }
      })
      list.push(...modifyVersions)
      if (noVersions[0].children.length > 0) {
        list.push(...noVersions)
      }
      return list
    }

  const props = defineProps<{
    groupListLoading: boolean;
    groupList: IGroupItemInService[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    value: IGroupTreeItem[];
  }>()

  const TYPE_MAP = {
    'current': '当前版本',
    'modify': '变更版本',
    'plain': '无版本'
  }

  const previewData = ref<IGroupPreviewItem[]>([])

  watch(() => props.value, (val) => {
    previewData.value = aggregateGroup(val)
  }, { immediate: true })


</script>
<template>
  <div class="roll-back-preview">
    <h3 class="title">
      上线预览
      <span class="tips">上线后，所选分组将从以下各版本更新至当前版本</span>
    </h3>
    <div class="version-list-wrapper">
      <bk-exception v-if="previewData.length === 0" scene="part" type="empty">
        <div class="empty-tips">
          暂无预览
          <p>请先从左侧选择待上线的分组范围</p>
        </div>
      </bk-exception>
      <div
        v-else
        v-for="version in previewData"
        class="version-callapse-item"
        :key="version.id">
        <div class="version-header">
          <div :class="['version-type-marking', version.type]">【{{ TYPE_MAP[version.type as keyof typeof TYPE_MAP] }}】</div>
          <span v-if="version.type === 'modify'" class="name"> - {{ version.name }}</span>
          <span class="group-count-wrapper">共 <span class="count">{{ version.children.length }}</span> 个分组</span>
        </div>
        <div class="group-list">
          <div v-for="group in version.children" class="group-item" :key="group.id">
            <span class="node-name">{{ group.name }}</span>
            <span class="split-line">|</span>
            <div class="rules">
              <template v-for="(rule, index) in group.rules" :key="index">
                <template v-if="index > 0"> ； </template>
                <rule-tag class="tag-item" :rule="rule"/>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .roll-back-preview {
    height: 100%;
  }
  .version-list-wrapper {
    height: calc(100% - 36px);
    overflow: auto;
  }
  .title {
    margin: 0 0 16px;
    line-height: 19px;
    font-size: 14px;
    font-weight: 700;
    color: #63656e;
    .tips {
      margin-left: 16px;
      line-height: 20px;
      color: #979ba5;
      font-size: 12px;
      font-weight: 400;
    }
  }
  .empty-tips {
    font-size: 14px;
    color: #63656e;
    & > p {
      margin: 8px 0 0;
      color: #979ba5;
      font-size: 12px;
    }
  }
  .version-callapse-item {
    margin-bottom: 16px;
    .version-header {
      display: flex;
      align-items: center;
      margin-bottom: 4px;
      line-height: 24px;
      font-size: 12px;
      color: #63656e;
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
    .group-item {
      display: flex;
      align-items: center;
      margin-bottom: 2px;
      padding: 8px 16px;
      background: #ffffff;
      border-radius: 2px;
      &:hover {
        background: #e1ecff;
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
      }
    }
  }
</style>