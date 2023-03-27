<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Search } from 'bkui-vue/lib/icon'
  import { IGroupCategoriesQuery, ECategoryType, IAllCategoryGroupItem, ICategoryTreeItem, IGroupTreeItem, IGroupRuleItem } from '../../../../../../../../../types/group'
  import { getAllGroupList } from '../../../../../../../../api/group';
  import { useServingStore } from '../../../../../../../../store/serving'
  import { GROUP_RULE_OPS } from '../../../../../../../../constants';


  const { appData } = storeToRefs(useServingStore())

  const categoryListLoading = ref(false)
  const categoryList = ref<ICategoryTreeItem[]>([])

  onMounted(() => {
    getAllGroupData()
  })

  const getAllGroupData = async() => {
    categoryListLoading.value = true
    const params: IGroupCategoriesQuery = {
      mode: ECategoryType.Custom,
      start: 0,
      limit: 100 // @todo 确认分页方式
    }
    const res = await getAllGroupList(<number>appData.value.id, params)
    const list: ICategoryTreeItem[] = []
    res.details.forEach((category: IAllCategoryGroupItem) => {
      const { group_category_id, group_category_name, groups } = category
      const groupsList = groups.map(item => {
        const { id, spec } = item
        const rules = <IGroupRuleItem[]>spec.selector.labels_and || spec.selector.labels_or
        const rulesWithName: { key: string; opName: string; value: string|number}[] = rules.map(item => {
          const { key, op, value } = item
          const opType = GROUP_RULE_OPS.find(typeItem => typeItem.id === op)
          console.log(opType)
          return { key, value, opName: <string>opType?.name }
        })
        return {
          id,
          rules: rulesWithName,
          label: spec.name,
        }
      })
      list.push({
        id: group_category_id,
        label: group_category_name,
        count: groupsList.length,
        children: groupsList
      })
    })

    categoryList.value = list
    categoryListLoading.value = false
  }

</script>
<template>
  <div class="group-tree-select">
    <h3 class="title">回滚相关分组</h3>
    <bk-input class="group-search-input" placeholder="请输入">
      <template #suffix>
        <Search class="search-input-icon" />
      </template>
    </bk-input>
    <div class="tree-wrapper">
      <bk-tree
        :data="categoryList"
        :show-checkbox="true"
        :show-node-type-icon="false">
        <template #node="node">
          <div class="node-label">
            <span class="label">{{ node.label }}</span>
            <template v-if="node.count">
              <span>（{{ node.count }}）</span>
            </template>
            <span class="rules" v-if="node.rules">
              <span> | </span>
              <span v-for="(rule, index) in node.rules" :key="index" class="rule">{{ `${rule.key} ${rule.opName} ${rule.value}` }}<span> ; </span></span>
            </span>
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
      padding: 0 8px;
      color: #63656e;
      font-size: 12px;
    }
    .rules {
      color: #979ba5;
    }
  }
</style>