// useTableAcrossCheckCommon.ts
import { h, Ref, ref } from 'vue';
import AcrossCheck from '../../components/across-check.vue';
import TableTip from '../../components/across-check-table-tip.vue';
import CheckType from '../../../types/across-checked';

export interface IAcrossCheckConfig {
  dataSource: Ref<number | any[]>;
  curPageData: Ref<any[]>;
  rowKey?: string[]; // 每行数据唯一标识
  crossPageSelect: Ref<boolean>;
}
export default function useTableAcrossCheckCommon({
  dataSource,
  curPageData,
  rowKey = ['name', 'id'],
  crossPageSelect,
}: IAcrossCheckConfig) {
  const selectType = ref(CheckType.Uncheck);
  const selections = ref<{ [key: string]: any }[]>([]);

  const getDataLength = () => {
    // 根据传入的是数字还是数组，返回不同的长度
    if (typeof dataSource.value === 'number') {
      return dataSource.value;
    }
    return dataSource.value.length;
  };

  const renderSelection = () => {
    // 渲染表头
    return h(AcrossCheck, {
      value: selectType.value,
      disabled: !getDataLength(),
      crossPageSelect: crossPageSelect.value,
      onChange: handleSelectTypeChange,
    });
  };
  const renderTableTip = () => {
    // 表格中间数据提示
    return h(TableTip, {
      dataLength: getDataLength(),
      selectionsLength: selections.value.length,
      isFullDataMode: typeof dataSource.value !== 'number',
      selectType: selectType.value,
      crossPageSelect: crossPageSelect.value,
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
    // number非全量，反向传递
    selections.value = typeof dataSource.value === 'number' ? [] : [...dataSource.value];
  };
  // 清空全选
  const handleClearSelection = () => {
    selectType.value = CheckType.Uncheck;
    selections.value = [];
  };

  // 表格行勾选后重置状态
  const handleSetSelectType = () => {
    // 全量数据
    if (typeof dataSource.value !== 'number') {
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
      } else if (selections.value.length < dataSource.value.length && selectType.value === CheckType.AcrossChecked) {
        // 从跨页全选 -> 跨页半选
        selectType.value = CheckType.HalfAcrossChecked;
      } else if (selections.value.length === dataSource.value.length) {
        selectType.value = CheckType.AcrossChecked;
      }
    } else {
      // 非全量数据
      if (
        (selections.value.length === 0 && [CheckType.Checked, CheckType.HalfChecked].includes(selectType.value)) ||
        (selections.value.length === dataSource.value &&
          [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value))
      ) {
        // 取消全选状态
        selectType.value = CheckType.Uncheck;
        selections.value = [];
      } else if (
        selections.value.length < curPageData.value.length &&
        [CheckType.Checked, CheckType.Uncheck].includes(selectType.value)
      ) {
        // 从当前页全选/空数据 -> 当前页半选
        selectType.value = CheckType.HalfChecked;
      } else if (
        selections.value.length === curPageData.value.length &&
        ![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value)
      ) {
        // 当前页全选
        selectType.value = CheckType.Checked;
      } else if (selections.value.length < dataSource.value && selectType.value === CheckType.AcrossChecked) {
        // 跨页半选
        selectType.value = CheckType.HalfAcrossChecked;
      } else if (selections.value.length === 0 && [CheckType.HalfAcrossChecked].includes(selectType.value)) {
        // 跨页全选
        selectType.value = CheckType.AcrossChecked;
      }
    }
  };

  // 当前行选中事件
  const handleRowCheckChange = (value: boolean, row: any) => {
    const index = selections.value.findIndex((item: { [key: string]: any }) =>
      rowKey.every((key) => item[key] === row[key]),
    );
    // 全量数据
    if (typeof dataSource.value !== 'number') {
      if (value && index === -1) {
        selections.value.push(row);
      } else if (!value && index > -1) {
        selections.value.splice(index, 1);
      }
    } else {
      if (value && index === -1 && [CheckType.Uncheck, CheckType.HalfChecked].includes(selectType.value)) {
        // 非跨页选择时，勾选数据正常push
        selections.value.push(row);
      } else if (!value && index > -1 && [CheckType.Checked, CheckType.HalfChecked].includes(selectType.value)) {
        // 非跨页选择时，取消勾选数据正常splice
        selections.value.splice(index, 1);
      } else if (
        !value &&
        index === -1 &&
        [CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value)
      ) {
        // 跨页全选/半选时，取消勾选数据push
        selections.value.push(row);
      } else if (
        value &&
        index > -1 &&
        [CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value)
      ) {
        // 跨页半选时，勾选数据splice
        selections.value.splice(index, 1);
      }
    }
    handleSetSelectType();
  };

  return {
    selectType,
    selections,
    renderSelection,
    renderTableTip,
    handleSelectionAll,
    handleClearSelection,
    handleRowCheckChange,
  };
}
