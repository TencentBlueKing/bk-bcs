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

// 文件下载
export const fileDownload = (content: string, name: string) => {
  const blob = new Blob([content], { type: 'text/plain;charset=UTF-8' })
  const eleLink = document.createElement('a')
  const blobURL = window.URL.createObjectURL(blob)
  eleLink.style.display = 'none'
  eleLink.href = blobURL
  eleLink.setAttribute('download', name)

  // hack HTML5 download attribute
  if (typeof eleLink.download === 'undefined') {
      eleLink.setAttribute('target', '_blank')
  }
  document.body.appendChild(eleLink)
  eleLink.click()
  document.body.removeChild(eleLink)
  window.URL.revokeObjectURL(blobURL)
}
