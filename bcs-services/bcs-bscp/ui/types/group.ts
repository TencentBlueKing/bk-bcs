export const enum ECategoryType {
  Custom = 'custom',
  Debug = 'debug'
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
  app_id: number|string;
  group_category_id: number|string;
  name: string;
  mode: string;
  selector?: string;
  uid?: number;
}

export interface ICategoryGroup {
  config: ICategoryItem;
  groups: {
    count: number;
    data: Array<IGroupItem>
  }
}