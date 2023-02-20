export type IPageFilter = {
  count: Boolean,
  start: number,
  limit: number
}

export const enum FilterOp {
  AND = 'and',
  OR = 'or'
}

export const enum RuleOp {
  eq = 'eq',
  neq = 'neq',
  gt = 'gt',
  gte = 'gte',
  lt = 'lt',
  lte = 'lte',
  in = 'in',
  nin = 'nin'
}

export type IRequestFilter = {
  op?: FilterOp,
  rules?: IRequestFilterRule[],
}

export type IRequestFilterRule = {
  field: string,
  op: RuleOp,
  value: any
}

export type IServingItem = {
  id?: number,
  biz_id: number,
  spec: {
    name: string,
    deploy_type: string,
    config_type: string,
    mode: string,
    memo: string,
    reload: {
      file_reload_spec: {
        reload_file_path: string
      },
      reload_type: string
    }
  },
  revision: {
    creator: string,
    reviser: string,
    create_at: string,
    update_at: string,
  }
}

export type IServingEditParams = {
  id?: number,
  biz_id?: number,
  app_id?: number,
  name: string,
  file_type: string,
  path?: string,
  file_mode?: string,
  user?: string,
  user_group?: string,
  privilege?: string
}