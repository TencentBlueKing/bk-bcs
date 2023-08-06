import { InfoBox } from "bkui-vue"

const useModalCloseConfirmation = (title?: string, subTitle?: string) => {
  return new Promise((resolve) => {
    InfoBox({
      title: title || '确认离开当前页？',
      subTitle: subTitle || '离开会导致未保存信息丢失',
      confirmText: '离开',
      onConfirm: () => {
        return resolve(true)
      },
      onClosed: () => {
        return resolve(false)
      }
    })
  } )
}

export default useModalCloseConfirmation
