<template>
  <div class="group-tree-container">
    <div class="operate-actions">
      <div class="btns">
        <bk-button text theme="primary" @click="handleSelectAll">{{ t('全选') }}</bk-button>
        <bk-button text theme="primary" @click="handleClearAll">{{ t('全不选') }}</bk-button>
      </div>
      <bk-input
        v-model="searchStr"
        class="group-search-input"
        :placeholder="t('搜索分组名称/标签key')"
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
        :selectable="false"
        :expand-all="true"
        :show-node-type-icon="false"
        :data="searchTreeData">
        <template #node="node">
          <div class="node-item-wrapper">
            <bk-checkbox
              size="small"
              :disabled="node.disabled"
              :model-value="getNodeCheckValue(node)"
              :indeterminate="getNodeIndeterminateValue(node)"
              v-bk-tooltips="{ content: t('已上线分组不可取消选择'), disabled: !node.disabled }"
              @change="(val: boolean) => handleNodeCheckChange(val, node)">
              <div class="node-label">
                <div class="label">{{ node.name }}</div>
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
            </bk-checkbox>
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
  import { useI18n } from 'vue-i18n';
  import { Search } from 'bkui-vue/lib/icon';
  import { IGroupRuleItem, IGroupToPublish } from '../../../../../../../../types/group';
  import { IConfigVersion } from '../../../../../../../../types/config';
  import RuleTag from '../../../../../groups/components/rule-tag.vue';
  import tableEmpty from '../../../../../../../components/table/table-empty.vue';

  // 分组叶子节点
  interface IGroupNode {
    id: number;
    name: string;
    disabled: boolean;
    release_id: number;
    release_name: string;
    published?: boolean;
    rules: IGroupRuleItem[];
  }

  // 父节点
  interface ITreeDataItem {
    id: number;
    name: string;
    children: IGroupNode[];
  }

  const { t } = useI18n();

  // 将全量分组数据按照版本分组
  const categorizingData = (groupList: IGroupToPublish[]) => {
    const list: ITreeDataItem[] = [];
    const unReleasedVersionNode: ITreeDataItem = {
      id: 0,
      name: '未上线版本的分组',
      children: [],
    };
    groupList.forEach((group) => {
      // id为0表示默认分组，在分组节点树中不可选
      if (group.id === 0) return;

      const disabled = props.releasedGroups.includes(group.id);
      const groupNode = { ...group, disabled };
      if (group.release_id > 0) {
        const parentNode = list.find((item) => item.id === group.release_id);
        if (parentNode) {
          parentNode.children.push(groupNode);
        } else {
          list.push({
            id: group.release_id,
            name: `已上线版本：${group.release_name}`,
            children: [groupNode],
          });
        }
      } else {
        unReleasedVersionNode.children.push(groupNode);
      }
    });

    if (unReleasedVersionNode.children.length > 0) {
      list.push(unReleasedVersionNode);
    }

    treeData.value = list;
  };

  const getNodeCheckValue = (node: IGroupNode | ITreeDataItem) => {
    if ('children' in node) {
      return node.children.every((childNode) => props.value.findIndex((group) => group.id === childNode.id) > -1);
    }
    return props.value.findIndex((group) => group.id === node.id) > -1;
  };

  const getNodeIndeterminateValue = (node: IGroupNode | ITreeDataItem) => {
    if ('children' in node) {
      let foundChecked = false;
      let foundUnChecked = false;
      for (const childNode of node.children) {
        if (props.value.findIndex((group) => group.id === childNode.id) > -1) {
          foundChecked = true;
        } else {
          foundUnChecked = true;
        }
      }
      return foundChecked && foundUnChecked;
    }
    return false;
  };

  const props = withDefaults(
    defineProps<{
      groupListLoading: boolean;
      groupList: IGroupToPublish[];
      versionListLoading: boolean;
      versionList: IConfigVersion[];
      releasedGroups?: number[]; // 调整分组上线时，【选择分组上线】已选择分组不可取消
      value: IGroupToPublish[];
    }>(),
    {
      releasedGroups: () => [],
    },
  );

  const emits = defineEmits(['change']);

  const treeData = ref<ITreeDataItem[]>([]);
  const searchStr = ref('');
  const treeRef = ref();
  const isSearchEmpty = ref(false);

  // 节点搜索
  const searchTreeData = computed(() => {
    if (searchStr.value === '') return treeData.value;
    isSearchEmpty.value = true;
    const searchText = searchStr.value.toLowerCase();
    const searchResults: ITreeDataItem[] = [];
    treeData.value.forEach((parentNode) => {
      if (parentNode.name.toLowerCase().includes(searchText)) {
        searchResults.push(parentNode);
      } else {
        const matchedGroups = parentNode.children.filter((group) => {
          const { name, rules } = group;
          return (
            name.toLowerCase().includes(searchText) || rules.some((rule) => rule.key.toLowerCase().includes(searchText))
          );
        });
        if (matchedGroups.length > 0) {
          searchResults.push({ ...parentNode, children: matchedGroups });
        }
      }
    });
    return searchResults;
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
  // watch(
  //   () => props.value,
  //   (val) => {
  //     const ids = val.map((item) => item.id);
  //     allGroupNode.value.forEach((node) => {
  //       node.checked = ids.includes(node.id);
  //     });
  //   },
  // );

  // 全选
  const handleSelectAll = () => {
    const groupList: IGroupToPublish[] = [];
    props.groupList.forEach((group) => {
      const hasGroupChecked = props.value.findIndex((item) => item.id === group.id) > -1; // 分组在编辑前是否选中
      const hasAdded = groupList.findIndex((item) => item.id === group.id) > -1; // 分组已添加
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
      const res = props.releasedGroups.includes(group.id) && props.value.findIndex((item) => item.id === group.id) > -1;
      return res;
    });
    emits('change', hasCheckedGroups);
  };

  // 选中/取消选中节点
  const handleNodeCheckChange = (val: boolean, node: any) => {
    const list = props.value.slice();

    if ('children' in node) {
      const parentNode = treeData.value.find((item) => item.id === node.id);
      if (parentNode) {
        // 父节点
        parentNode.children.forEach((childNode) => {
          if (childNode.disabled) return;
          const index = list.findIndex((item) => item.id === childNode.id);
          const hasGroupChecked = index > -1;
          if (val) {
            const group = props.groupList.find((group) => group.id === childNode.id);
            if (group && !hasGroupChecked) {
              list.push(group);
            }
          } else {
            if (index > -1) {
              list.splice(index, 1);
            }
          }
        });
      }
    } else {
      // 叶子节点
      if (node.disabled) return;
      if (val) {
        const group = props.groupList.find((group) => group.id === node.id);
        if (group) {
          list.push(group);
        }
      } else {
        const index = list.findIndex((item) => item.id === node.id);
        list.splice(index, 1);
      }
    }

    console.log(val, node);
    console.log(list);
    // debugger; // eslint-disable-line no-debugger
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
    max-height: 450px;
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
