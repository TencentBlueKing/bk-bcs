<script setup lang="ts">
  import { computed, onMounted, ref, watch } from 'vue'
  import { Search, AngleDown, AngleUp } from 'bkui-vue/lib/icon'
  import { IGroupToPublish } from '../../../../../../../../../../types/group'
  import { IConfigVersion } from '../../../../../../../../../../types/config'
  import RuleTag from '../../../../../../../groups/components/rule-tag.vue'

  interface treeParentNodeData {
    node_id: string;
    name: string;
    count: number;
    children: IGroupNodeData[];
  }
  
  interface IGroupNodeData extends IGroupToPublish {
    node_id: string;
  }

  // 将全量分组数据按照规则key分组，并记录所有分组节点数据
  const categorizingData = (data: IGroupToPublish[]) => {
    const nodeItemList: IGroupNodeData[] = []
    const treeNodeData: treeParentNodeData[] = []
    data.forEach(group => {
      if (group.id === 0) { // id为0表示默认分组，在分组节点树中不可选
        return
      }
      group.rules.forEach(rule => {
        const nodeId = `${rule.key}_${group.id}` // 用在节点树上做唯一标识
        const data = treeNodeData.find(item => item.node_id === rule.key)
        const nodeData = { ...group, node_id: nodeId}
        if (data) {
          data.count++
          data.children.push(nodeData)
        } else {
          treeNodeData.push({
            node_id: rule.key,
            name: rule.key,
            count: 1,
            children: [nodeData]
          })
        }
        nodeItemList.push({ ...nodeData })
      })
    })
    treeData.value = treeNodeData
    allGroupNode.value = nodeItemList
  }

  const props = withDefaults(defineProps<{
    groupListLoading: boolean;
    groupList: IGroupToPublish[];
    versionListLoading: boolean;
    versionList: IConfigVersion[];
    value: IGroupToPublish[];
  }>(), {
    groupList: () => [],
    versionList: () => []
  })

  const emits = defineEmits(['change'])

  const treeData = ref<{ name: string; count: number; children: IGroupNodeData[] }[]>([])
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
  watch(() => props.value, (newVal, oldVal) => {
    console.log('newVal: ', newVal.map(item => item.id))
    if (oldVal) {
      console.log('oldVal: ',  oldVal.map(item => item.id))
    }
    
    // tree组件UI上选中节点
    newVal.forEach(checkedGroupItem => {
      const groupNodes = allGroupNode.value.filter(node => node.id === checkedGroupItem.id)
      groupNodes.forEach(node => {
        treeRef.value.setChecked(node.node_id, true)
      })
    })
    // tree组件UI上取消选中节点
    if (oldVal) {
      oldVal.forEach(group => {
        if (!newVal.find(item => item.id === group.id)) {
          const groupNodes = allGroupNode.value.filter(node => node.id === group.id)
          groupNodes.forEach(node => {
            treeRef.value.setChecked(node.node_id, false)
          })
        }
      })
    }
  })

  onMounted(() => {
    if (props.value.length > 0) {
      // tree组件UI上选中节点
      props.value.forEach(group => {
        const groupNodes = allGroupNode.value.filter(node => node.id === group.id)
        groupNodes.forEach(node => {
          treeRef.value.setChecked(node.node_id, true)
        })
      })
    }
  })

  // 全选
  const handleSelectAll = () => {
    const groupList: IGroupToPublish[] = []
    allGroupNode.value.forEach(node => {
      if (!groupList.find(group => group.id === node.id)) {
        groupList.push(node)
      }
    })
    emits('change', groupList)
  }

  // 全不选
  const handleClearAll = () => {
    emits('change', [])
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
    console.log(val, itemValue, item)
    searchStr.value = val
  }

  // 选中节点
  const handleNodeChecked = (selected: string[]) => {
    console.log(selected)
    const list: IGroupToPublish[] = []
    selected.forEach(id => {
      const group = allGroupNode.value.find(group => group.node_id === id)
      if (group && !list.find(item => item.id === group.id)) { // 相同分组的节点去重
        list.push(group)
      }
    })
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