<script setup lang="ts">
  import { ref, computed } from 'vue'
  import { IGroupToPublish } from '../../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../../types/config'
  import GroupTree from './tree.vue'

  const props = withDefaults(defineProps<{
    groupListLoading: boolean;
    groupList: IGroupToPublish[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    disabled?: number[];
    value: IGroupToPublish[];
  }>(), {
    groupList: () => [],
    versionList: () => [],
    disabled: () => [],
    value: () => []
  })

  const emits = defineEmits(['togglePreviewDelete', 'change'])

  const type = ref('select')

  // 选择上线的分组，排除分组需要做取反操作
  const selectedGroup = computed(() => {
    if (type.value === 'exclude') {
      return props.groupList.filter(group => props.value.findIndex(item => item.id === group.id) === -1)
    }
    return props.value
  })

  // 切换选择分组类型
  const handleTypeChange = (val: string) => {
    type.value = val
    if (val === 'all') {
      handleSelectGroup(props.groupList)
    } else if (val === 'select') {
      const list = props.groupList.filter(group => props.disabled.includes(group.id))
      handleSelectGroup(list)
    } else {
      let list: IGroupToPublish[] = []
      if (props.disabled.length > 0) {
        debugger
        list = props.groupList.filter(group => !props.disabled.includes(group.id))
      }
      handleSelectGroup(list)
    }
    emits('togglePreviewDelete', val === 'select')
  }

  const handleSelectGroup = (val: IGroupToPublish[]) => {
    if (type.value === 'exclude') {
      const list: IGroupToPublish[] = props.groupList.filter(group => val.findIndex(item => item.id === group.id) === -1)
      emits('change', list)
    } else {
      emits('change', val)
    }
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
            :disabled="props.disabled"
            :value="selectedGroup"
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
            :disabled="props.disabled"
            :value="selectedGroup"
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