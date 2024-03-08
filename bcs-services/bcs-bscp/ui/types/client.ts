// 搜索项
export interface ISelectorItem {
  name: string;
  value: string;
  children?: ISelectorItem[];
}

// 查询条件
export interface ISearchCondition {
  key: string;
  value: string;
  content: string;
}
