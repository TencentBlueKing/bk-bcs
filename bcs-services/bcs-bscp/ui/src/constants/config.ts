export const CONFIG_FILE_TYPE = [
  { id: 'text', name: '文本文件' },
  { id: 'binary', name: '二进制文件' }
]

export const CONFIG_STATUS_MAP = {
  'ADD': {
    text: '新增',
    color: '#3a84ff',
    bgColor: '#edf4ff'
  },
  'DELETE': {
    text: '删除',
    color: '#feebea',
    bgColor: '#ea3536'
  },
  'REVISE': {
    text: '修改',
    color: '#fe9c00',
    bgColor: '#fff1db'
  },
  'UNCHANGE': {
    text: '--',
    color: '',
    bgColor: ''
  }
}

export const VERSION_STATUS_MAP = {
  'not_released': '未上线',
  'partial_released': '灰度中',
  'full_released': '已上线'
}

export const GET_UNNAMED_VERSION_DATE = () => {
  return {
    id: 0,
    attachment: {
      app_id: 0,
      biz_id: 0
    },
    revision: {
      create_at: '',
      creator: ''
    },
    spec: {
      name: '未命名版本',
      memo: '',
      deprecated: false,
      publish_num: 0,
    },
    status: {
      publish_status: 'editing',
      released_groups: []
    }
  }
}
