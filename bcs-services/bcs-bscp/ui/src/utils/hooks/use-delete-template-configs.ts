import { InfoBox } from "bkui-vue"
import { ITemplateConfigItem } from "../../../types/template"
import { deleteTemplate } from '../../api/template'

const useDeleteTemplateConfigs = (space_id: string, current_template_space: number, configs: ITemplateConfigItem[]) => {
  if (!configs || configs.length === 0) return

  return new Promise((resolve) => {
    InfoBox({
      title: `确认彻底删除${configs.length > 1 ? configs.length + '条配置项' : '配置项【' + configs[0].spec.name + '】'}？`,
      subTitle:'删除后不可找回，请谨慎操作。',
      confirmText: '确认删除',
      infoType: 'warning',
      onConfirm: async() => {
        const ids = configs.map(config => config.id)
        return deleteTemplate(space_id, current_template_space, ids)
      },
      onClosed: () => {
        return resolve(false)
      }
    })
  } )
}

export default useDeleteTemplateConfigs
