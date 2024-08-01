import { h, Ref, ref } from 'vue';
import AcrossCheck from '../../components/across-check.vue';
import TableTip from '../../components/across-check-table-tip.vue';
import CheckType from '../../../types/across-checked';

export interface IAcrossCheckConfig {
  tableData: Ref<any[]>; // 全量数据
  curPageData: Ref<any[]>; // 当前页数据
  rowKey?: string[]; // 每行数据唯一标识
  arrowShow?: Ref<boolean>; // 多选下拉菜单展示
}
// 表格跨页全选功能
export default function useTableAcrossCheck({
  tableData,
  curPageData,
  rowKey = ['name', 'id'],
  arrowShow,
}: IAcrossCheckConfig) {
  const selectType = ref(CheckType.Uncheck);
  const selections = ref<any[]>([]);
  const renderSelection = () => {
    // 渲染表头
    return h(AcrossCheck, {
      value: selectType.value,
      disabled: !tableData.value.length,
      arrowShow: arrowShow?.value,
      onChange: handleSelectTypeChange,
    });
  };
  const renderTableTip = () => {
    // 表格中间数据提示
    return h(TableTip, {
      dataLength: tableData.value.length,
      selectionsLength: selections.value.length,
      isFullDataMode: true,
      arrowShow: !arrowShow?.value,
      handleSelectTypeChange,
      handleClearSelection,
    });
  };

  // 表头全选事件
  const handleSelectTypeChange = (value: number) => {
    switch (value) {
      case CheckType.Uncheck:
        handleClearSelection();
        break;
      case CheckType.Checked:
        handleSelectCurrentPage();
        break;
      case CheckType.AcrossChecked:
        handleSelectionAll();
        break;
    }
  };
  // 当前页全选
  const handleSelectCurrentPage = () => {
    selectType.value = CheckType.Checked;
    selections.value = [...curPageData.value];
  };
  // 跨页全选
  const handleSelectionAll = () => {
    selectType.value = CheckType.AcrossChecked;
    selections.value = [...tableData.value];
  };
  // 清空全选
  const handleClearSelection = () => {
    selectType.value = CheckType.Uncheck;
    selections.value = [];
  };
  // 表格行勾选后重置状态
  const handleSetSelectType = () => {
    if (selections.value.length === 0) {
      selectType.value = CheckType.Uncheck;
    } else if (
      selections.value.length < curPageData.value.length &&
      [CheckType.Checked, CheckType.Uncheck].includes(selectType.value)
    ) {
      // 从当前页全选 -> 当前页半选
      selectType.value = CheckType.HalfChecked;
    } else if (
      selections.value.length === curPageData.value.length &&
      ![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value)
    ) {
      selectType.value = CheckType.Checked;
    } else if (selections.value.length < tableData.value.length && selectType.value === CheckType.AcrossChecked) {
      // 从跨页全选 -> 跨页半选
      selectType.value = CheckType.HalfAcrossChecked;
    } else if (selections.value.length === tableData.value.length) {
      selectType.value = CheckType.AcrossChecked;
    }
  };
  // 当前行选中事件
  const handleRowCheckChange = (value: boolean, row: any) => {
    const index = selections.value.findIndex((item) => rowKey.every((key) => item[key] === row[key]));
    if (value && index === -1) {
      selections.value.push(row);
    } else if (!value && index > -1) {
      selections.value.splice(index, 1);
    }
    handleSetSelectType();
  };

  return {
    selectType, // 选择状态
    selections, // 选中的数据
    renderSelection, // 渲染全选框组件
    renderTableTip, // 渲染表格中间数据提示
    handleRowCheckChange, // 行的选择框操作
    handleSelectionAll, // 全选操作
    handleClearSelection, // 清空全选
  };
}
