import { InfoBox } from 'bkui-vue';
import { ITemplateConfigItem } from '../../../types/template';
import { deleteTemplate } from '../../api/template';
import { localT } from '../../i18n';
const useDeleteTemplateConfigs = (space_id: string, current_template_space: number, configs: ITemplateConfigItem[]) => {
  if (!configs || configs.length === 0) return;
  return new Promise((resolve) => {
    InfoBox({
      title: `${localT('确认彻底删除')}${
        configs.length > 1
          ? `${configs.length}${localT('条配置文件')}`
          : `${localT('配置文件')}【${configs[0].spec.name}】`
      }？`,
      subTitle: localT('删除后不可找回，请谨慎操作。'),
      confirmText: localT('确认删除'),
      infoType: 'warning',
      onConfirm: async () => {
        const ids = configs.map((config) => config.id);
        return deleteTemplate(space_id, current_template_space, ids);
      },
      onClosed: () => resolve(false),
    });
  });
};

export default useDeleteTemplateConfigs;
