import { h, Ref, ref } from 'vue';
import AcrossCheck from '../../components/across-check.vue';
import CheckType from '../../../types/across-checked';

export interface IAcrossCheckConfig {
  tableData: Ref<any[]>;
  curPageData: Ref<any[]>;
  rowKey?: string[];
  arrowShow?: Ref<boolean>;
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
    return h(AcrossCheck, {
      value: selectType.value,
      disabled: !tableData.value.length,
      arrowShow: !arrowShow?.value,
      onChange: handleSelectTypeChange,
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
      selectType.value !== CheckType.HalfAcrossChecked
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
    handleRowCheckChange, // 行的全选框操作
    handleSelectionAll, // 全选操作
    handleClearSelection, // 清空全选
  };
}
