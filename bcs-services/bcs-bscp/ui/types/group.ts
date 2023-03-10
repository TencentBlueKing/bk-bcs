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

export interface IGroupCategoriesQuery {
  mode: ECategoryType;
  start: number;
  limit: number;
}

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

export interface IGroupItem {
  id: number;
  spec: {
    name: string;
    mode: string;
    uid: string;
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

export interface IGroupEditing {
  id?: number;
  group_category_id: number|string;
  name: string;
  mode: string;
  rules: Array<IGroupRuleItem>;
  rule_logic: string;
  uid?: string;
}

export interface IGroupRuleItem {
  key: string;
  op: EGroupRuleType;
  value: string|number
}

// 分组新建、编辑提交到接口参数
export interface IGroupEditArg {
  name: string;
  group_category_id: number;
  mode: ECategoryType;
  selector?: {
    labels_and?: Array<IGroupRuleItem>;
    labels_or?: Array<IGroupRuleItem>
  };
  uid?: string
}