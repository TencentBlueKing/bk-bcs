export const enum ECategoryType {
  Custom = 'custom',
  Debug = 'debug',
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

// 全量分类下分组列表单个详情
export interface IAllCategoryGroupItem {
  group_category_id: number;
  group_category_name: string;
  groups: IGroupItem[];
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
  };
}

// 分组详情
export interface IGroupItem {
  id: number;
  name: string;
  public: boolean;
  bind_apps: { name: string; id: number }[];
  released_apps_num: number;
  mode?: string;
  uid?: string;
  selector: {
    labels_and?: IGroupRuleItem[];
    labels_or?: IGroupRuleItem[];
  };
}

// 分组按规则key分类数据
export interface IGroupCategory {
  name: string;
  show: boolean;
  fold: boolean;
  children: IGroupCategoryItem[];
}

// 分组按规则key分类的单条表格数据
export interface IGroupCategoryItem {
  IS_CATEORY_ROW?: boolean;
  CATEGORY_NAME?: string;
  fold?: boolean;
  id?: number;
  name?: string;
  public?: boolean;
  bind_apps: { name: string; id: number }[];
  released_apps_num?: number;
  selector?: {
    labels_and?: IGroupRuleItem[];
    labels_or?: IGroupRuleItem[];
  };
}

// 分组编辑数据
export interface IGroupEditing {
  id?: number;
  name: string;
  public: boolean;
  bind_apps: number[];
  rules: IGroupRuleItem[];
  rule_logic: string;
  uid?: string;
}

// 分组规则
export interface IGroupRuleItem {
  key: string;
  op: EGroupRuleType | string;
  value: string | number | string[];
}

// 分组新建、编辑提交到接口参数
export interface IGroupEditArg {
  id?: number;
  name: string;
  group_category_id?: number;
  mode?: ECategoryType;
  selector?: {
    labels_and?: Array<IGroupRuleItem>;
    labels_or?: Array<IGroupRuleItem>;
  };
  uid?: string;
}

// 选择上线的分组
export interface IGroupToPublish {
  id: number;
  name: string;
  release_id: number;
  release_name: string;
  published?: boolean;
  desc?: string;
  rules: IGroupRuleItem[];
}

// 分组绑定的已上线服务
export interface IGroupBindService {
  app_id: number;
  app_name: string;
  release_id: number;
  release_name: string;
  edited: boolean;
}

// 服务下的分组数据
export interface IGroupItemInService {
  group_id: number;
  group_name: string;
  release_id: number;
  release_name: string;
  old_selector: {
    labels_or?: IGroupRuleItem[];
    labels_and?: IGroupRuleItem[];
  };
  new_selector: {
    labels_or?: IGroupRuleItem[];
    labels_and?: IGroupRuleItem[];
  };
  edited: boolean;
}

// 上线预览分组按照版本聚合
export interface IGroupPreviewItem {
  id: number;
  name: string;
  type: string;
  children: IGroupToPublish[];
}
