<template>
  <div class="roll-back-preview">
    <h3 class="title">
      上线预览
      <span class="tips">上线后，以下分组将从以下各版本更新至当前版本</span>
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
          :group-list="props.groupList"
          :preview-group="previewGroup"
          :allow-preview-delete="props.releaseType === 'select'"
          :disabled="props.disabled"
          @diff="emits('diff', $event)"
          @delete="handleDelete"
        >
        </preview-version-group>
      </template>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';
import { IGroupToPublish, IGroupPreviewItem } from '../../../../../../../../types/group';
import { IConfigVersion } from '../../../../../../../../types/config';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../../store/config';
import PreviewVersionGroup from './preview-version-group.vue';

const versionStore = useConfigStore();
const { versionData } = storeToRefs(versionStore);

// 将分组按照版本聚合
const aggregateGroup = (groups: IGroupToPublish[]) => {
  const list: IGroupPreviewItem[] = [];
  // 变更版本
  const modifiedVersionGroups: IGroupPreviewItem[] = [];
  // 首次上线
  const initialReleaseGroup: IGroupPreviewItem = { id: 0, name: '首次上线', type: 'plain', children: [] };
  // 需要被展示的分组
  const groupsToBePreviewed: IGroupToPublish[] = groups.filter(group => {
    const { id, release_id } = group;
    // 过滤掉当前版本已上线分组
    if (release_id === versionData.value.id) {
      return false;
    }

    return id === 0 || release_id > 0 || props.releaseType === 'select';
  })

  // 全部实例上线
  // 只展示已上线的分组，如果默认分组已上线，放到【变更版本】分组中，否则放到【首次上线】分组中
  // 选择分组实例上线
  // 新添加的分组状态取决于默认分组是否已上线
  // 排除分组实例上线
  // 默认不勾选分组，至少勾选一个分组才能提交

  // 默认分组是否已上线，则将分组放到【变更版本】中，否则放到【首次上线】中
  groupsToBePreviewed.forEach((group) => {
    const { release_id, release_name } = group;
    if (props.isDefaultGroupReleased) {
      const version = modifiedVersionGroups.find(item => item.id === release_id);
      if (version) {
        version.children.push(group);
      } else {
        const defaultGroup = props.groupList.find(group => group.id === 0);
        const name = release_id === 0 ? (defaultGroup as IGroupToPublish).release_name : release_name
        modifiedVersionGroups.push({ id: release_id, name, type: 'modify', children: [group] });
      }
    } else {
      initialReleaseGroup.children.push(group);
    }
  });

  list.push(...modifiedVersionGroups);
  if (initialReleaseGroup.children.length > 0) {
    list.push(initialReleaseGroup);
  }
  return list;
};

const props = withDefaults(
  defineProps<{
    groupListLoading: boolean;
    groupList: IGroupToPublish[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    releaseType: string;
    isDefaultGroupReleased: boolean;
    disabled?: number[];
    value: IGroupToPublish[];
  }>(),
  {
    disabled: () => [],
  },
);

const emits = defineEmits(['diff', 'change']);

const previewData = ref<IGroupPreviewItem[]>([]);

watch(
  () => props.value,
  (val) => {
    previewData.value = aggregateGroup(val);
  },
  { immediate: true },
);

const handleDelete = (id: number) => {
  emits(
    'change',
    props.value.filter(group => group.id !== id),
  );
};
</script>
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
