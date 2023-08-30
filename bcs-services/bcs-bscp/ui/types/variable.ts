
export interface IVariableItem {
  id: number;
  spec: {
    name: string;
    memo: string;
    type: string;
    default_val: string;
  };
  attachment: {
    biz_id: number;
  };
  revision: {
    creator: string;
    reviser: string;
    create_at: string;
    update_at: string;
  };
}

export interface IVariableEditParams {
  name: string;
  type: string;
  default_val: string;
  memo: string;
}

// 服务中使用的变量详情
export interface IConfigVariableItem {
  name: string;
  type: string;
  default_val: string;
  memo: string;
}
