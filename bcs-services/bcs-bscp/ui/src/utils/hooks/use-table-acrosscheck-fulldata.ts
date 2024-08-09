import { Ref, ref } from 'vue';
import useTableAcrossCheckCommon from './use-table-acrosscheck-common';

export interface IAcrossCheckConfig {
  tableData: Ref<any[]>; // 全量数据
  curPageData: Ref<any[]>; // 当前页数据
  rowKey?: string[]; // 每行数据唯一标识
  crossPageSelect: Ref<boolean>; // 是否提供全选/跨页全选功能
}
// 表格跨页全选功能
export default function useTableAcrossCheck({
  tableData,
  curPageData,
  rowKey = ['name', 'id'],
  crossPageSelect = ref(true),
}: IAcrossCheckConfig) {
  const result = useTableAcrossCheckCommon({
    dataSource: tableData,
    curPageData,
    rowKey,
    crossPageSelect,
  });

  return {
    ...result,
  };
}
