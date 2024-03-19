<template>
  <div v-bkloading="{ loading: props.loading, opacity: 1 }" class="select-group-wrapper">
    <template v-if="!props.loading">
      <div class="group-tree-area">
        <Group
          :group-list="props.groupList"
          :version-list="props.versionList"
          :released-groups="props.releasedGroups"
          :release-type="props.releaseType"
          :disable-select="props.disableSelect"
          :value="props.groups"
          @release-type-change="emits('releaseTypeChange', $event)"
          @change="emits('change', $event)" />
      </div>
      <div class="preview-area">
        <Preview
          :group-list="props.groupList"
          :release-type="props.releaseType"
          :released-groups="props.releasedGroups"
          :value="props.groups"
          @diff="emits('openPreviewVersionDiff', $event)"
          @change="emits('change', $event)" />
      </div>
    </template>
  </div>
</template>
<script setup lang="ts">
  import { IConfigVersion } from '../../../../../../../../types/config';
  import { IGroupToPublish } from '../../../../../../../../types/group';
  import Group from './group.vue';
  import Preview from './preview.vue';

  const props = withDefaults(
    defineProps<{
      loading: boolean;
      versionList: IConfigVersion[];
      groupList: IGroupToPublish[];
      releaseType?: string;
      releasedGroups?: number[];
      groups: IGroupToPublish[];
      disableSelect?: boolean; // 是否隐藏【选择分组实例上线】方式
    }>(),
    {
      loading: true,
      releaseType: 'select',
      disableSelect: false,
    },
  );
  const emits = defineEmits(['openPreviewVersionDiff', 'releaseTypeChange', 'change']);
</script>
<style lang="scss" scoped>
  .select-group-wrapper {
    display: flex;
    align-items: center;
    min-width: 1366px;
    height: 100%;
    background: #ffffff;
  }
  .group-tree-area {
    padding: 24px;
    width: 566px;
    height: 100%;
    border-right: 1px solid #dcdee5;
  }
  .preview-area {
    flex: 1;
    padding: 24px 0;
    height: 100%;
    background: #f5f7fa;
    overflow: hidden;
  }
</style>
