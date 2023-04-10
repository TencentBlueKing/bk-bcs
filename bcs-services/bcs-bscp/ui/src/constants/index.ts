export const GROUP_RULE_OPS = [
  { id: 'eq', name: '=' },
  { id: 'ne', name: '!=' },
  { id: 'gt', name: '>' },
  { id: 'ge', name: '>=' },
  { id: 'lt', name: '<' },
  { id: 'le', name: '<=' },
  { id: 'in', name: 'IN' },
  { id: 'nin', name: 'NOT IN' }
]

export const CONFIG_FILE_TYPE = [
  { id: 'text', name: 'Text' },
  { id: 'binary', name: '二进制文件' }
]

export const CONFIG_STATUS_MAP = {
  'ADD': '增加',
  'DELETE': '删除',
  'REVISE': '修改',
  'UNCHANGE': '--'
}
