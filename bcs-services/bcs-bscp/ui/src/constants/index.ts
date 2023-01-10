export type IPage = {
  count?: number,
  start: number,
  limit: number
}

export enum FilterOp {
  AND = 'and',
  OR = 'or'
}

export enum RuleOp {
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
