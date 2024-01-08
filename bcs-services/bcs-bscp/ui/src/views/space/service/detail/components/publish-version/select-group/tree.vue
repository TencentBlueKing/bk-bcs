<template>
  <div class="group-tree-container">
    <div class="operate-actions">
      <div class="btns">
        <bk-button text theme="primary" @click="handleSelectAll">全选</bk-button>
        <bk-button text theme="primary" @click="handleClearAll">全不选</bk-button>
        <bk-select
          class="version-select"
          no-data-text="暂无已上线的可选版本"
          :multiple="true"
          :filterable="true"
          :input-search="false"
          :show-select-all="true"
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
          <!-- <bk-option v-if="props.versionList.length > 0" :value="0">全选</bk-option> -->
          <bk-option
            v-for="version in props.versionList"
            :key="version.id"
            :label="version.spec.name"
            :value="version.id" />
        </bk-select>
      </div>
      <bk-input v-model="searchStr" class="group-search-input" placeholder="搜索分组名称/标签key" :clearable="true">
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
        :selectable="false"
        :data="searchTreeData"
        :expand-all="false"
        :show-node-type-icon="false"
      >
        <template #node="node">
          <div class="node-item-wrapper">
            <bk-checkbox
              size="small"
              :model-value="node.checked"
              :disabled="node.disabled"
              :indeterminate="node.indeterminate"
              v-bk-tooltips="{ content: '已上线分组不可取消选择',disabled: !node.disabled }"
              @change="handleNodeCheckChange(node)">
            </bk-checkbox>
            <div class="node-label" @click="handleNodeCheckChange(node)">
              <div class="label">{{ node.name }}</div>
              <span v-if="node.count" class="count">({{ node.count }})</span>
              <template v-if="node.rules">
                <span class="split-line"> | </span>
                <div class="rules">
                  <bk-overflow-title type="tips">
                    <span v-for="(rule, index) in node.rules" :key="index" class="rule">
                      <span v-if="index > 0"> & </span>
                      <rule-tag class="tag-item" :rule="rule" />
                    </span>
                  </bk-overflow-title>
                </div>
              </template>
            </div>
          </div>
        </template>
        <template #empty>
          <tableEmpty :is-search-empty="isSearchEmpty" @clear="handleClearSearch"></tableEmpty>
        </template>
      </bk-tree>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Search, AngleDown, AngleUp } from 'bkui-vue/lib/icon';
import { IGroupToPublish } from '../../../../../../../../types/group';
import { IConfigVersion } from '../../../../../../../../types/config';
import RuleTag from '../../../../../groups/components/rule-tag.vue';
import tableEmpty from '../../../../../../../components/table/table-empty.vue';

interface IGroupNodeData extends IGroupToPublish {
  node_id: number;
  checked: boolean;
  disabled: boolean;
}

// 将全量分组数据按照规则key分组，并记录所有分组节点数据
const categorizingData = (groupList: IGroupToPublish[]) => {
  const nodeItemList: IGroupNodeData[] = [];
  groupList.forEach((group) => {
    if (group.id === 0) {
      // id为0表示默认分组，在分组节点树中不可选
      return;
    }
    const checked = props.value.findIndex(item => item.id === group.id) > -1;
    const disabled = props.releasedGroups.includes(group.id);
    nodeItemList.push({ ...group, node_id: group.id, checked, disabled });
  });

  allGroupNode.value = nodeItemList;
};

const props = withDefaults(defineProps<{
  groupListLoading: boolean;
  groupList: IGroupToPublish[];
  versionListLoading: boolean;
  versionList: IConfigVersion[];
  releasedGroups?: number[]; // 调整分组上线时，【选择分组上线】已选择分组不可取消
  value: IGroupToPublish[];
}>(), {
  releasedGroups: () => [],
});

const emits = defineEmits(['change']);

const allGroupNode = ref<IGroupNodeData[]>([]); // 树中所有的分组叶子节点
const versionSelectorOpen = ref(false);
const searchStr = ref('');
const treeRef = ref();
const isSearchEmpty = ref(false);
const selectedVersionIds = ref<number[]>([]);

// 节点搜索
const searchTreeData = computed(() => {
  if (searchStr.value === '') return allGroupNode.value;
  isSearchEmpty.value = true;
  return allGroupNode.value.filter(node => {
    const { name, rules } = node
    const searchText = searchStr.value.toLowerCase()
    return name.toLowerCase().includes(searchText) || rules.some(rule => rule.key.toLowerCase().includes(searchText))
  });
});

// 分组列表变更
watch(
  () => props.groupList,
  (val) => {
    categorizingData(val);
  },
  { immediate: true },
);

// 选中小组变更
watch(
  () => props.value,
  (val) => {
    const ids = val.map(item => item.id);
    allGroupNode.value.forEach(node => {
      node.checked = ids.includes(node.id);
    })
  },
);

// 全选
const handleSelectAll = () => {
  const groupList: IGroupToPublish[] = [];
  props.groupList.forEach((group) => {
    const hasGroupChecked = props.value.findIndex(item => item.id === group.id) > -1; // 分组在编辑前是否选中
    const hasAdded = groupList.findIndex(item => item.id === group.id) > -1; // 分组已添加
    const isDisabled = props.releasedGroups.includes(group.id);
    if (group.id !== 0 && !hasAdded && (!isDisabled || hasGroupChecked)) {
      groupList.push(group);
    }
  });
  emits('change', groupList);
};

// 全不选
const handleClearAll = () => {
  const hasCheckedGroups = props.groupList.filter((group) => {
    const res = props.releasedGroups.includes(group.id) && props.value.findIndex(item => item.id === group.id) > -1;
    return res;
  });
  emits('change', hasCheckedGroups);
};

// 按版本选择
const handleSelectVersion = (val: number[]) => {
  const list: IGroupToPublish[] = [];
  val.forEach(id => {
    const version = props.versionList.find(item => item.id === id);
    if (version) {
      version.status.released_groups.forEach((releaseItem) => {
        if (!list.find(item => releaseItem.id === item.id)) {
          const group = allGroupNode.value.find(groupItem => groupItem.id === releaseItem.id);
          if (group) {
            list.push(group);
          }
        }
      });
    }
  });
  // 调整分组上线时，当前版本已上线分组不可取消
  props.releasedGroups.forEach(id => {
    if (!list.find(item => item.id === id)) {
      const group = allGroupNode.value.find(groupItem => groupItem.id === id);
      if (group) {
        list.push(group);
      }
    }
  });
  emits('change', list);
};

// 选中/取消选中节点
const handleNodeCheckChange = (node: IGroupNodeData) => {
  if (node.disabled) {
    return;
  }
  const list = props.value.slice();
  // 叶子节点
  const group = props.groupList.find(group => group.id === node.id);
  if (group) {
    const index = list.findIndex(item => item.id === group.id);
    if (index === -1) {
      list.push(group);
    } else {
      list.splice(index, 1);
    }
  }
  emits('change', list);
};

// 清空筛选条件
const handleClearSearch = () => {
  searchStr.value = '';
  isSearchEmpty.value = false;
};
</script>
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
    flex: 1;
    padding: 0 8px;
    color: #63656e;
    font-size: 12px;
    overflow: hidden;
  }
  .count {
    margin-left: 4px;
  }
  .split-line {
    margin: 0 4px;
    color: #979ba5;
  }
  .rules {
    flex: 1;
    min-width: 0;
    color: #979ba5;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
