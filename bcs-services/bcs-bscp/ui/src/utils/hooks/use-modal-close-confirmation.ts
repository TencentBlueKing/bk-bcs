import { InfoBox } from 'bkui-vue';
import { localT } from '../../i18n';
const useModalCloseConfirmation = (title?: string, subTitle?: string) => new Promise((resolve) => {
  InfoBox({
    title: title || localT('确认离开当前页？'),
    subTitle: subTitle || localT('离开会导致未保存信息丢失'),
    confirmText: localT('离开'),
    onConfirm: () => {
      resolve(true);
    },
    onClosed: () => {
      resolve(false);
    },
  });
});

export default useModalCloseConfirmation;
