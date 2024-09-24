// 操作记录列表参数
export interface IRecordQuery {
  id?: number; // 记录id
  app_id?: number;
  start_time?: string;
  end_time?: string;
  operate?: string;
  start?: number;
  limit?: number;
  all?: boolean;
  name?: string;
  resource_type?: string;
  action?: string;
  res_instance?: string;
  status?: string;
  operator?: string;
  operate_way?: string;
}

// 审批操作：撤销/驳回/通过/手动上线
export interface IDialogData {
  service?: string;
  version?: string;
  group?: string;
}

// 列表每行的数据
export interface IRowData {
  audit: {
    id: number;
    spec: {
      res_type: string;
      action: string;
      rid?: string;
      app_code?: string;
      is_compare: boolean;
      detail?: string;
      operator: string;
      res_instance: string;
      operate_way: string;
      status: string;
    };
    attachment: {
      biz_id: number;
      app_id: number;
      res_id: number;
    };
    revision: {
      created_at: string;
    };
  };
  strategy: {
    publish_type: string;
    publish_time: string;
    publish_status: string;
    reject_reason: string;
    approver: string;
    approver_progress: string;
    updated_at: string;
    reviser: string;
    release_id: number;
    scope: {
      groups: [
        {
          id: number;
          spec: {
            name: string;
            public: boolean;
            bind_apps: [];
            mode: string;
            selector: {
              labels_and: [
                {
                  key: string;
                  op: string;
                  value: string;
                },
              ];
            };
            uid: string;
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
        },
        {
          id: 2;
          spec: {
            name: string;
            public: true;
            bind_apps: [];
            mode: string;
            selector: {
              labels_and: [
                {
                  key: string;
                  op: string;
                  value: string;
                },
              ];
            };
            uid: string;
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
        },
      ];
    };
  };
  app: {
    name: string;
    creator: string;
  };
}
