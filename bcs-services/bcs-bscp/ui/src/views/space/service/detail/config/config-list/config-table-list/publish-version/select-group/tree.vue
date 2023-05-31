<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { Search, AngleDown, AngleUp } from 'bkui-vue/lib/icon'
  import { IGroupToPublish } from '../../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../../types/config'
  import RuleTag from '../../../../../../../groups/components/rule-tag.vue'

  interface ITreeParentNodeData {
    parent: boolean;
    node_id: string;
    name: string;
    count: number;
    checked: boolean;
    indeterminate: boolean;
    children: IGroupNodeData[];
  }
  
  interface IGroupNodeData extends IGroupToPublish {
    node_id: string;
    checked: boolean;
    disabled: boolean;
  }

  // 将全量分组数据按照规则key分组，并记录所有分组节点数据
  const categorizingData = (groupList: IGroupToPublish[]) => {
    const nodeItemList: IGroupNodeData[] = []
    const treeNodeData: ITreeParentNodeData[] = []
    groupList.forEach(group => {
      if (group.id === 0) { // id为0表示默认分组，在分组节点树中不可选
        return
      }
      const checked = props.value.findIndex(item => item.id === group.id) > -1
      const disabled = props.disabled.includes(group.id)
      group.rules.forEach(rule => {
        const nodeId = `${rule.key}_${group.id}` // 用在节点树上做唯一标识
        const parentNode = treeNodeData.find(item => item.node_id === rule.key)
        const nodeData = { ...group, node_id: nodeId, checked, disabled }
        if (parentNode) {
          parentNode.count++
          parentNode.children.push(nodeData)
        } else {
          treeNodeData.push({
            parent: true,
            node_id: rule.key,
            name: rule.key,
            count: 1,
            checked: false,
            indeterminate: false,
            children: [nodeData]
          })
        }
        nodeItemList.push({ ...nodeData })
      })
    })
    treeNodeData.forEach(parentNode => {
      parentNode.checked = isParentNodeChecked(parentNode)
      parentNode.indeterminate = isParentNodeIndeterminate(parentNode)
    })
    treeData.value = treeNodeData
    allGroupNode.value = nodeItemList
  }

  // 父级分类节点是否选中
  const isParentNodeChecked = (node: ITreeParentNodeData) => {
    return node.children.length > 0 && node.children.every(group => group.checked)
  }

  // 父级分类节点是否半选
  const isParentNodeIndeterminate = (node: ITreeParentNodeData) => {
    return node.children.length > 0 && !node.children.every(group => group.checked) && node.children.some(group => group.checked)
  }

  const props = defineProps<{
    groupListLoading: boolean;
    groupList: IGroupToPublish[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    disabled: number[]; // 调整分组上线时，【选择分组上线】已选择分组不可取消，【排除分组上线】已选择分组不可勾选
    value: IGroupToPublish[];
  }>()

  const emits = defineEmits(['change'])

  const treeData = ref<ITreeParentNodeData[]>([])
  const allGroupNode = ref<IGroupNodeData[]>([]) // 树中所有的分组叶子节点
  const versionSelectorOpen = ref(false)
  const searchStr = ref('')
  const treeRef = ref()

  const searchOption = computed(() => {
    return {
      value: searchStr.value,
      match: handleSearch
    }
  })

  // 分组列表变更
  watch(() => props.groupList, (val) => {
    categorizingData(val)
  }, { immediate: true })

  // 选中小组变更
  watch(() => props.value, val => {
    const ids = val.map(item => item.id)
    treeData.value.forEach(parentNode => {
      parentNode.children.forEach(node => {
        node.checked = ids.includes(node.id)
      })
      parentNode.checked = isParentNodeChecked(parentNode)
      parentNode.indeterminate = isParentNodeIndeterminate(parentNode)
    })
  })

  // 全选
  const handleSelectAll = () => {
    const groupList: IGroupToPublish[] = []
    props.groupList.forEach(group => {
      const hasGroupChecked = props.value.findIndex(item => item.id === group.id) > -1 // 分组在编辑前是否选中
      const hasAdded = groupList.findIndex(item => item.id === group.id) > -1// 分组已添加
      const isDisabled = props.disabled.includes(group.id)
      if (group.id !== 0 && !hasAdded && (!isDisabled || hasGroupChecked)) {
        groupList.push(group)
      }
    })
    emits('change', groupList)
  }

  // 全不选
  const handleClearAll = () => {
    const hasCheckedGroups = props.groupList.filter(group => {
      return props.disabled.includes(group.id) && props.value.findIndex(item => item.id === group.id) > -1
    })
    emits('change', hasCheckedGroups)
  }

  // 按版本选择
  const handleSelectVersion = (versions: number[]) => {
    const selectedVersion: IConfigVersion[] = []
    const list: IGroupToPublish[] = []
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
    selectedVersion.forEach(version => {
      version.status.released_groups.forEach(releaseItem => {
        if (!list.find(item => releaseItem.id === item.id)) {
          const group = allGroupNode.value.find(groupItem => groupItem.id === releaseItem.id)
          if (group) {
            list.push(group)
          }
        }
      })
    })
    emits('change', list)
  }

  // 节点搜索
  const handleSearch = (val: string, itemValue: string, item: any) => {
    return itemValue.toLowerCase().includes(val.toLowerCase())
  }

  // 选中/取消选中节点
  const handleNodeCheckChange = (node: IGroupNodeData|ITreeParentNodeData, checked: boolean) => {
    const list = props.value.slice()
    if (node.hasOwnProperty('parent')) { // 分类节点
      const treeParentNode = treeData.value.find(parentNode => parentNode.node_id === node.node_id)
      if (treeParentNode) {
        if (checked) {
          treeParentNode.children.filter(group => !group.disabled).forEach(group => {
            if (!list.find(item => item.id === group.id)) {
              list.push(group)
            }
          })
        } else {
          treeParentNode.children.filter(group => !group.disabled).forEach(group => {
            const index = list.findIndex(item => item.id === group.id)
            if (index > -1) {
              list.splice(index, 1)
            }
          })
        }
      }
    } else { // 叶子节点
      const group = props.groupList.find(group => group.id === (<IGroupNodeData>node).id)
      if (group) {
        if (checked) {
          list.push(group)
        } else {
          const index = list.findIndex(item => item.id === group.id)
          list.splice(index, 1)
        }
      }
    }
    emits('change', list)
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
        node-key="node_id"
        :data="treeData"
        :expand-all="false"
        :show-node-type-icon="false"
        :search="searchOption">
        <template #node="node">
          <div class="node-item-wrapper">
            <bk-checkbox
              size="small"
              :model-value="node.checked"
              :disabled="node.disabled"
              :indeterminate="node.indeterminate"
              @change="handleNodeCheckChange(node, $event)">
            </bk-checkbox>
            <div class="node-label">
              <div class="label">{{ node.name }}</div>
              <span v-if="node.count" class="count">({{ node.count }})</span>
              <template v-if="node.rules">
                <span class="split-line"> | </span>
                <div class="rules">
                  <span v-for="(rule, index) in node.rules" :key="index" class="rule">
                    <span v-if="index > 0 "> & </span>
                    <rule-tag class="tag-item" :rule="rule"/>
                  </span>
                </div>
              </template>
            </div>
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
    .node-item-wrapper {
      display: flex;
      align-items: center;
      overflow: hidden;
    }
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