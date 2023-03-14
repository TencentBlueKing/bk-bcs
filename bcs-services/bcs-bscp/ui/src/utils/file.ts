import { byteUnitConverse } from './index'

// 将文本转换为二进制文件
export const transDataToFile = (data: string, name: string) => {
  const blob = new Blob([data], { type: 'application/octet-stream' })
  const file = new File([blob], name, { type: 'application/octet-stream' })
  return file
}

// 将File对象转换为json
export const transFileToObject = (file: File) => {
  const { name, size } = file
  return { name, size: byteUnitConverse(size) }
}
