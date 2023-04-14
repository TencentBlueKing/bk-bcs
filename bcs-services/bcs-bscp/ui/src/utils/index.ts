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

// 字符串内容的字节大小
// @notice：edge 79版本才开始支持，发布时间2020-01-15 https://developer.mozilla.org/zh-CN/docs/Web/API/TextEncode
export const stringLengthInBytes = (content: string) => {
  return (new TextEncoder().encode(content)).length
}