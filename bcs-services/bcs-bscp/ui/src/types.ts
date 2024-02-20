export type IPageFilter = {
  count: Boolean;
  start: number;
  limit: number;
};

export const enum FilterOp {
  AND = 'and',
  OR = 'or',
}

export const enum RuleOp {
  eq = 'eq',
  neq = 'neq',
  gt = 'gt',
  gte = 'gte',
  lt = 'lt',
  lte = 'lte',
  in = 'in',
  nin = 'nin',
}

export type IRequestFilter = {
  op: FilterOp;
  rules: IRequestFilterRule[];
};

export type IRequestFilterRule = {
  field: string;
  op: RuleOp;
  value: any;
};
