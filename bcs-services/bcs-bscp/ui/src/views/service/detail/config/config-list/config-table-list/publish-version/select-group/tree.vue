<script setup lang="ts">
  import { computed, ref, watch, withDefaults } from 'vue'
  import { Search, AngleDown, AngleUp } from 'bkui-vue/lib/icon'
  import { IGroupTreeItem, IGroupItemInService } from '../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../types/config'
  import RuleTag from '../../../../../../../groups/components/rule-tag.vue'

  // 将全量分组数据按照规则key分组
    const categorizingData = (data: IGroupItemInService[]) => {
    const treeData: { name: string; count: number; children: IGroupTreeItem[] }[] = []
    data.forEach(group => {
      const selector = group.new_selector // @todo 待确认上线版本时，分组规则取哪个
      const rules = selector.labels_and || selector.labels_or
      const { group_id, group_name, release_id, release_name } = group
      rules?.forEach(rule => {
        const data = treeData.find(item => item.name === rule.key)
        if (data) {
          data.count++
          data.children.push({ id: group_id, name: group_name, release_id, release_name, rules: rules })
        } else {
          treeData.push({
            name: rule.key,
            count: 1,
            children: [{ id: group_id, name: group_name, release_id, release_name, rules: rules }]
          })
        }
      })
    })
    return treeData
  }

  const props = withDefaults(defineProps<{
    groupListLoading: boolean;
    groupList: IGroupItemInService[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    value: IGroupTreeItem[];
  }>(), {
    groupList: () => [],
    versionList: () => []
  })

  const emits = defineEmits(['change'])

  const categorizedData = ref<{ name: string; count: number; children: IGroupTreeItem[] }[]>([])
  const treeData = ref<{ name: string; count: number; children: IGroupTreeItem[] }[]>([])
  const groups = ref<IGroupTreeItem[]>([]) // 选中的分组
  const versionSelectorOpen = ref(false)
  const searchStr = ref('')
  const treeRef = ref()

  const searchOption = computed(() => {
    return {
      value: searchStr.value,
      match: handleSearch
    }
  })

  watch(() => props.groupList, (val) => {
    treeData.value = categorizingData(val)
    categorizedData.value = categorizingData(val)
  }, { immediate: true })

  watch(() => props.value, (val) => {
    console.log('value: ', val)
  })

  // 全选
  const handleSelectAll = () => {
    const allGroupNodes = getAllGroupNodes()
    groups.value = allGroupNodes
    treeRef.value.setChecked(allGroupNodes, true)
    emits('change', groups.value)
  }

  // 全不选
  const handleClearAll = () => {
    treeRef.value.setChecked(getAllGroupNodes(), false)
    groups.value = []
    emits('change', groups.value)
  }

  // 获取所有分组节点
  const getAllGroupNodes = () => {
    const allGroupNode: IGroupTreeItem[] = []
    treeData.value.forEach(item => allGroupNode.push(...item.children))
    return allGroupNode
  }

  // 按版本选择
  const handleSelectVersion = (versions: number[]) => {
    console.log(versions)
    const groupNodes: IGroupTreeItem[] = []
    const selectedVersion: IConfigVersion[] = []
    if (versions.includes(0)) { // 全选
      selectedVersion.push(...props.versionList)
    } else { // 选择部分
      versions.forEach(id => {
        const version = props.versionList.find(item => item.id === id)
        if (version) {
            selectedVersion.push(version)
        }
      })
    }
    const selectedGroups: number[] = []
    selectedVersion.forEach(version => {
      version.status.released_groups.forEach(group => {
        selectedGroups.push(group.id)
      })
    })
    const allGroupNode = getAllGroupNodes()
    selectedGroups.forEach(id => {
      const nodes = allGroupNode.filter(item => item.id === id)
      groupNodes.push(...nodes)
    })
    groups.value = groupNodes
    treeRef.value.setChecked(allGroupNode, false)
    treeRef.value.setChecked(groupNodes, true)
    emits('change', groupNodes)
  }

  // const handleSearch = () => {
  //   if (searchStr.value) {

  //   } else {
      
  //   }
  // }
  const handleSearch = (val: string, itemValue: string, item: any) => {
    console.log(val, itemValue, item)
    searchStr.value = val
  }

  const handleNodeChecked = (selected: string[]) => {
    const allGroupList: IGroupTreeItem[] = []
    const list: IGroupTreeItem[] = []
    treeData.value.forEach(item => allGroupList.push(...item.children))
    selected.forEach(item => {
      const group = allGroupList.find(group => group.__uuid === item)
      if (group) {
        list.push(group)
      }
    })
    groups.value = list
    emits('change', groups.value)
  }

</script>
<template>
  <div class="group-tree-container">
    <div class="operate-actions">
      <div class="btns">
        <bk-button text theme="primary" @click="handleSelectAll">全选</bk-button>
        <bk-button text theme="primary" @click="handleClearAll">全不选</bk-button>
        <bk-select
          class="version-select"
          empty-text="暂无可上线分组"
          :multiple="true"
          :filterable="true"
          :input-search="false"
          :popover-min-width="240"
          @change="handleSelectVersion"
          @toggle="versionSelectorOpen = $event">
          <template #trigger>
            <bk-button text theme="primary">
              按版本选择
              <AngleUp class="arrow-icon" v-if="versionSelectorOpen" />
              <AngleDown class="arrow-icon" v-else />
            </bk-button>
          </template>
          <bk-option :value="0">全选</bk-option>
          <bk-option v-for="version in props.versionList" :key="version.id" :label="version.spec.name" :value="version.id"></bk-option>
        </bk-select>
      </div>
      <bk-input
        v-model="searchStr"
        class="group-search-input"
        placeholder="搜索分组名称"
        :clearable="true">
        <template #suffix>
          <Search class="search-input-icon" />
        </template>
      </bk-input>
    </div>
    <div class="group-select-tree">
      <bk-tree
        ref="treeRef"
        label="name"
        :data="treeData"
        :show-checkbox="true"
        :expand-all="false"
        :show-node-type-icon="false"
        :search="searchOption"
        @node-checked="handleNodeChecked">
        <template #node="node">
          <div class="node-label">
            <div class="label">{{ node.name }}</div>
            <span v-if="node.count" class="count">({{ node.count }})</span>
            <template v-if="node.rules">
              <span class="split-line"> | </span>
              <div class="rules">
                <bk-overflow-title type="tips">
                  <span v-for="(rule, index) in node.rules" :key="index" class="rule">
                    <span v-if="index > 0 "> & </span>
                    <rule-tag class="tag-item" :rule="rule"/>
                  </span>
                </bk-overflow-title>
              </div>
            </template>
          </div>
        </template>
      </bk-tree>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .group-tree-container {
    margin: 8px 0 12px;
    padding-bottom: 10px;
    border: 1px solid #dcdee5;
    border-radius: 3px;
  }
  .operate-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 7px 15px 7px 24px;
    background: #f5f7fa;
    .btns {
      display: flex;
      align-items: center;
      > .bk-button {
        margin-right: 16px;
      }
    }
  }
  .arrow-icon {
    font-size: 16px;
  }
  .group-search-input {
    width: 240px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .group-select-tree {
    margin-top: 8px;
    max-height: 403px;
    overflow: auto;
    .node-label {
      display: flex;
      align-items: center;
      padding: 0 8px;
      color: #63656e;
      font-size: 12px;
    }
    .count {
      margin-left: 4px;
    }
    .split-line {
      margin: 0 4px;
      color: #979ba5;
    }
    .rules {
      min-width: 0;
      color: #979ba5;
    }
  }
</style>