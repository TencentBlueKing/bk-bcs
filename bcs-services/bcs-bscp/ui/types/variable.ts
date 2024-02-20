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

// 服务中使用的变量详情
export interface IVariableEditParams {
  name: string;
  type: string;
  default_val: string;
  memo: string;
}

// 变量被配置文件引用详情
export interface IVariableCitedByConfigDetailItem {
  variable_name: string;
  references: {
    id: number;
    template_revision_id: number;
    name: string;
    path: string;
  }[];
}

// 批量导入参数
export interface IVariableImportParams {
  separator: string;
  variables: string;
}
