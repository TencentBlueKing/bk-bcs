export const enum ECategoryType {
  Custom = 'custom',
  Debug = 'debug'
}

export const enum EGroupRuleType {
  '=' = 'eq',
  '!=' = 'ne',
  '>' = 'gt',
  '>=' = 'ge',
  '<' = 'lt',
  '<=' = 'le',
  'IN' = 'in',
  'NOT IN' = 'nin',
}

// 分组分类查询接口参数
export interface IGroupCategoriesQuery {
  mode: ECategoryType;
  start: number;
  limit: number;
}

// 分类详情
export interface ICategoryItem {
  id: number;
  spec: {
    name: string;
  };
  attachment: {
    biz_id: number;
    app_id: number;
    group_category_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  }
}

// 分组详情
export interface IGroupItem {
  id: number;
  spec: {
    name: string;
    mode: string;
    uid: string;
    selector: {
      labels_and?: IGroupRuleItem[];
      labels_or?: IGroupRuleItem[];
    }
  };
  attachment: {
    biz_id: number;
    app_id: number;
    group_category_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  }
}

// 分组编辑数据
export interface IGroupEditing {
  id?: number;
  group_category_id: number|string;
  name: string;
  mode: string;
  rules: IGroupRuleItem[];
  rule_logic: string;
  uid?: string;
}

// 分组规则
export interface IGroupRuleItem {
  key: string;
  op: EGroupRuleType;
  value: string|number
}

// 分组新建、编辑提交到接口参数
export interface IGroupEditArg {
  id?: number,
  name: string;
  group_category_id?: number;
  mode?: ECategoryType;
  selector?: {
    labels_and?: Array<IGroupRuleItem>;
    labels_or?: Array<IGroupRuleItem>
  };
  uid?: string
}

// 全量分类下分组列表单个详情
export interface IAllCategoryGroupItem {
  group_category_id: number;
  group_category_name: string;
  groups: IGroupItem[];
}

// 分组选择树单个详情
export interface ICategoryTreeItem {
  id: number;
  label: string;
  count: number;
  children: IGroupTreeItem[];
}

// 分组选择树分组节点单个详情
export interface IGroupTreeItem {
  id: number;
  label: string;
  rules: {key: string; opName: string; value: string|number}[];
}
