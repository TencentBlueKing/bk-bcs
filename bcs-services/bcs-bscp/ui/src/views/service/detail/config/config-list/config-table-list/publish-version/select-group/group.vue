<script setup lang="ts">
  import { ref } from 'vue'
  import { IGroupTreeItem, IGroupItemInService } from '../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../types/config'
  import GroupTree from './tree.vue'

  const props = withDefaults(defineProps<{
    groupListLoading: boolean;
    groupList: IGroupItemInService[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    groups: IGroupTreeItem[];
  }>(), {
    groupList: () => [],
    versionList: () => []
  })

  const emits = defineEmits(['change'])

  const type = ref('select')

  const handleTypeChange = (val: string) => {
    type.value = val
  }

  const handleSelectGroup = (val: IGroupTreeItem[]) => {
    emits('change', val)
  }

</script>
<template>
  <div class="group-select-wrapper">
    <h3 class="title">选择上线范围</h3>
    <div class="select-group-radius">
      <bk-radio-group :model-value="type" @change="handleTypeChange">
        <bk-radio label="all">
          全部分组上线
        </bk-radio>
        <bk-radio label="select">
          选择分组上线
          <GroupTree
            v-if="type === 'select'"
            :group-list="props.groupList"
            :group-list-loading="props.groupListLoading"
            :version-list="props.versionList"
            :version-list-loading="props.versionListLoading"
            :value="props.groups"
            @change="handleSelectGroup">
          </GroupTree>
        </bk-radio>
        <bk-radio label="exclude">
          排除分组上线
          <GroupTree
            v-if="type === 'exclude'"
            :group-list="groupList"
            :group-list-loading="groupListLoading"
            :version-list="versionList"
            :version-list-loading="versionListLoading"
            :value="props.groups"
            @change="handleSelectGroup">
          </GroupTree>
        </bk-radio>
      </bk-radio-group>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .group-select-wrapper {
    height: 100%
  }
  .title {
    margin: 0 0 22px;
    line-height: 19px;
    font-size: 14px;
    font-weight: 700;
    color: #63656e;
  }
  .bk-radio-group {
    display: block;
  }
  .bk-radio {
    display: block;
    margin: 0 0 12px;
    &:last-of-type {
      margin-bottom: 0;
    }
  }
</style>