import { CONFIG_FILE_TYPE } from '../constants/index'

// 字节数转换为对应的显示单位
export const byteUnitConverse = (size: number) => {
  if (0 <= size && size < 1024) {
    return `${size}B`
  } else if (1024 <= size && size < 1024 * 1024) {
    return `${Math.ceil(size / 1024)}KB`
  } else if (1024 * 1024 <= size && size < 1024 * 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)}MB`
  } else if (1024 * 1024 * 1024 <= size) {
    return `${(size / (1024 * 1024 * 1024)).toFixed(1)}GB`
  }
}

// 查询配置文件类型名称
export const getConfigTypeName = (type: string) => {
  const fileType = CONFIG_FILE_TYPE.find(item => item.id === type)
  return fileType ? fileType.name : '未知格式'
}