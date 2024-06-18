<template>
  <bk-tree
    v-if="pkgTreeData.length"
    ref="treeRef"
    label="name"
    node-key="nodeId"
    :data="pkgTreeData"
    :search="{ value: props.searchStr }"
    :show-node-type-icon="false">
    <template #node="node">
      <div class="node-item-wrapper">
        <bk-checkbox
          size="small"
          :model-value="node.checked"
          :disabled="node.disabled"
          :indeterminate="node.indeterminate"
          @change="handleNodeCheckChange(node, $event)" />
        <div
          class="node-name-text"
          v-bk-tooltips="{
            content: checkboxTooltips(node.checked),
            disabled: !node.disabled,
          }">
          {{ node.name }}
        </div>
        <span v-if="node.children" class="num">({{ node.children.length }})</span>
      </div>
    </template>
  </bk-tree>
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ITemplateBoundByAppData } from '../../../../../../../../../../types/config';
  import { IAllPkgsGroupBySpaceInBiz, IPkgTreeItem } from '../../../../../../../../../../types/template';
  interface ISpaceTreeItem extends IPkgTreeItem {
    isOpen: boolean;
    children: IPkgTreeItem[];
  }

  const { t } = useI18n();
  const props = defineProps<{
    bkBizId: string;
    pkgList: IAllPkgsGroupBySpaceInBiz[];
    imported: ITemplateBoundByAppData[]; // 编辑状态下已经选中的套餐，该列表下数据不可取消
    value: ITemplateBoundByAppData[]; // 当前选中的套餐信息
    searchStr: string;
  }>();

  const emits = defineEmits(['change']);

  const checkboxTooltips = (isImport: boolean) => {
    return isImport
      ? t('导入配置模板套餐时无法移除已有配置模板套餐')
      : t('该套餐中没有可用配置文件，无法被导入到服务配置中');
  };
  const pkgTreeData = computed(() => {
    return props.pkgList.map((item) => {
      const { template_space_id, template_space_name, template_sets } = item;
      const nodeId = `space_${template_space_id}`;
      let checked = false;
      let indeterminate = false;
      const isOpen = template_sets.some((set) => {
        const index = props.value.findIndex((item) => item.template_set_id === set.template_set_id);
        return index > -1;
      });

      if (template_sets.length > 0) {
        checked = template_sets.every((pkgItem) => isPkgNodeChecked(pkgItem.template_set_id));
        if (!checked) {
          indeterminate = template_sets.some((pkgItem) => isPkgNodeChecked(pkgItem.template_set_id));
        }
      }

      const group: ISpaceTreeItem = {
        id: template_space_id,
        nodeId,
        name: template_space_name,
        isOpen,
        children: [],
        checked,
        indeterminate,
        disabled: false,
      };

      const notEmptyTplPkgNodes: IPkgTreeItem[] = [];
      const emptyTplPkgNodes: IPkgTreeItem[] = [];

      template_sets.forEach((pkg) => {
        const isEmpty = pkg.template_ids.length === 0;
        const node = {
          id: pkg.template_set_id,
          nodeId: `pkg_${pkg.template_set_id}`,
          name: pkg.template_set_name,
          checked: isPkgNodeChecked(pkg.template_set_id),
          disabled: isPkgImported(pkg.template_set_id) || isEmpty,
          indeterminate: false,
          parentName: template_space_name,
        };
        if (isEmpty) {
          emptyTplPkgNodes.push(node);
        } else {
          notEmptyTplPkgNodes.push(node);
        }
      });
      group.children = [...notEmptyTplPkgNodes, ...emptyTplPkgNodes];
      return group;
    });
  });

  const handleNodeCheckChange = (node: ISpaceTreeItem, val: boolean) => {
    const list = props.value.slice();
    if (node.children) {
      // 空间节点
      const pkgNodes = node.children;
      if (val) {
        pkgNodes.forEach((pkg) => {
          if (!pkg.disabled && !isPkgNodeChecked(pkg.id)) {
            list.push({
              template_set_id: pkg.id,
              template_revisions: [],
              template_set_name: `${node.name} / ${pkg.name}`,
            });
          }
        });
      } else {
        pkgNodes.forEach((pkg) => {
          if (isPkgImported(pkg.id)) return;
          const index = list.findIndex((item) => item.template_set_id === pkg.id);
          if (index > -1) {
            list.splice(index, 1);
          }
        });
      }
    } else {
      // 套餐节点
      if (isPkgImported(node.id)) return;
      if (val) {
        if (!isPkgNodeChecked(node.id)) {
          list.push({
            template_set_id: node.id,
            template_revisions: [],
            template_set_name: `${node.parentName}-${node.name}`,
          });
        }
      } else {
        const index = list.findIndex((item) => item.template_set_id === node.id);
        if (index > -1) {
          list.splice(index, 1);
        }
      }
    }
    emits('change', list);
  };

  const isPkgImported = (id: number) => props.imported.some((pkg) => pkg.template_set_id === id);

  const isPkgNodeChecked = (id: number) => isPkgImported(id) || props.value.some((pkg) => pkg.template_set_id === id);
</script>
<style lang="scss" scoped>
  .manage-templates {
    width: 420px;
    height: 40px;
    background: #fafbfd;
    border: 1px solid #dcdee5;
    border-radius: 0 0 2px 2px;
  }

  .packages-tree {
    margin-top: 8px;
    padding: 0 24px;
    height: calc(100% - 87px);
    overflow: auto;
    .node-item-wrapper {
      display: flex;
      align-items: center;
    }
  }
  .node-item-wrapper {
    display: flex;
    align-items: center;
    .node-name-text {
      padding: 0 4px;
      font-size: 12px;
      color: #63656e;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .num {
      flex-shrink: 0;
      font-size: 12px;
      color: #63656e;
    }
  }
</style>
