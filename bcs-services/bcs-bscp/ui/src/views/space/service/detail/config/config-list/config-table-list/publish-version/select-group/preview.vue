<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { IGroupToPublish, IGroupPreviewItem } from '../../../../../../../../../../types/group'
  import PreviewVersionGroup from './preview-version-group.vue';

  // 将分组按照版本聚合
  const aggregateGroup = (groups: IGroupToPublish[]) => {
    const list: IGroupPreviewItem[] = []
    const modifyVersions: IGroupPreviewItem[] = []
    const noVersions: IGroupPreviewItem[] = [{ id: 0, name: '无版本', type: 'plain', children: [] }]
    groups.forEach((group) => {
      const { release_id, release_name } = group
      if (release_id) {
        const version = modifyVersions.find(item => item.id === release_id)
        if (version) {
          version.children.push(group)
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

  const props = withDefaults(defineProps<{
    groupListLoading: boolean;
    groupList: IGroupToPublish[];
    allowPreviewDelete: boolean;
    disabled?: number[];
    value: IGroupToPublish[];
  }>(), {
    disabled: () => []
  })

  const emits = defineEmits(['diff', 'change'])

  const previewData = ref<IGroupPreviewItem[]>([])

  watch(() => props.value, (val) => {
    previewData.value = aggregateGroup(val)
  }, { immediate: true })

  const handleDelete = (id: number) => {
    emits('change', props.value.filter(group => group.id !== id))
  }

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
      <template v-else>
        <preview-version-group
          v-for="previewGroup in previewData"
          :key="previewGroup.id"
          :preview-group="previewGroup"
          :allow-preview-delete="allowPreviewDelete"
          :disabled="props.disabled"
          @diff="emits('diff', $event)"
          @delete="handleDelete">
        </preview-version-group>
      </template>
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
    padding: 0 24px;
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
</style>