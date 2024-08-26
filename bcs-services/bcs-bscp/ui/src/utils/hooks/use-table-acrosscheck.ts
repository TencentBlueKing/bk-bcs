import { Ref, ref } from 'vue';
import useTableAcrossCheckCommon from './use-table-acrosscheck-common';

export interface IAcrossCheckConfig {
  dataCount: Ref<number>; // 可选的数据总数，不含禁用状态
  curPageData: Ref<any[]>; // 当前页数据
  rowKey?: string[]; // 每行数据唯一标识；需要在每行数据的第一子层级，暂不支持递归查找
  crossPageSelect: Ref<boolean>; // 是否提供全选/跨页全选功能
}
// 表格跨页全选功能
export default function useTableAcrossCheck({
  dataCount,
  curPageData,
  rowKey = ['name', 'id'],
  crossPageSelect = ref(true),
}: IAcrossCheckConfig) {
  const result = useTableAcrossCheckCommon({
    dataSource: dataCount,
    curPageData,
    rowKey,
    crossPageSelect,
  });
  return {
    ...result,
  };
}
