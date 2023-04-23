<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Search } from 'bkui-vue/lib/icon'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { IGroupTreeItem, IGroupItem } from '../../../../../../../../../types/group'
  import { getSpaceGroupList } from '../../../../../../../../api/group'
  import RuleTag from '../../../../../../../groups/components/rule-tag.vue'

  const { spaceId } = storeToRefs(useGlobalStore())

  const emits = defineEmits(['change'])

  const listLoading = ref(true)
  const groupList = ref<{ label: string; count: number; children: IGroupTreeItem[] }[]>([])
  const groups = ref<IGroupTreeItem[]>([]) // 选中的分组

  // const categoryListLoading = ref(false)
  // const categoryList = ref<ICategoryTreeItem[]>([])

  onMounted(() => {
    getAllGroupData()
  })

  // 获取所有分组，并转化为tree组件需要的结构
  const getAllGroupData = async() => {
    listLoading.value = true
    const res = await getSpaceGroupList(spaceId.value)
    groupList.value = categorizingData(res.details)
    listLoading.value = false
  }

    // 将全量分组数据按照分类分组
  const categorizingData = (data: IGroupItem[]) => {
    const treeData: { label: string; count: number; children: IGroupTreeItem[] }[] = []
    data.forEach(group => {
      const selector = group.selector.labels_and || group.selector.labels_or
      const { id, name } = group
      selector?.forEach(rule => {
        const data = treeData.find(item => item.label === rule.key)
        if (data) {
          data.count++
          data.children.push({ id, label: name, rules: selector })
        } else {
          treeData.push({
            label: rule.key,
            count: 1,
            children: [{ id, label: name, rules: selector }]
          })
        }
      })
    })
    return treeData
  }

  const handleNodeChecked = (selected: string[]) => {
    console.log(selected)
    const allGroupList: IGroupTreeItem[] = []
    const list: IGroupTreeItem[] = []
    groupList.value.forEach(item => allGroupList.push(...item.children))
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
  <div class="group-tree-select">
    <h3 class="title">选择上线分组</h3>
    <bk-input class="group-search-input" placeholder="请输入">
      <template #suffix>
        <Search class="search-input-icon" />
      </template>
    </bk-input>
    <div class="tree-wrapper">
      <bk-tree
        :data="groupList"
        :show-checkbox="true"
        :expand-all="true"
        :show-node-type-icon="false"
        @node-checked="handleNodeChecked">
        <template #node="node">
          <div class="node-label">
            <div class="label">{{ node.label }}</div>
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
  .group-tree-select {
    height: 100%
  }
  .title {
    margin: 0 0 16px;
    line-height: 19px;
    font-size: 14px;
    font-weight: 700;
    color: #63656e;
  }
  .group-search-input {
    width: 100%;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
  }
  .tree-wrapper {
    margin-top: 8px;
    height: calc(100% - 100px);
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