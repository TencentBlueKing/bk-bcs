import { Ref, watch, ref } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';

export interface IConfig {
  watchOnce?: boolean
}
export default function useSideslider(data: Ref<any> = ref(''), config: IConfig = { watchOnce: false }) {
  const isChanged = ref(false);

  const watchOnce = watch(data, () => {
    isChanged.value = true;
    config.watchOnce && watchOnce();
  }, { deep: true });

  const handleBeforeClose = () => new Promise((resolve, reject) => {
    if (!isChanged.value) {
      resolve(true);
      return;
    };
    $bkInfo({
      title: $i18n.t('generic.msg.info.exitTips.text'),
      subTitle: $i18n.t('generic.msg.info.exitTips.subTitle'),
      clsName: 'custom-info-confirm default-info',
      okText: $i18n.t('generic.button.exit'),
      cancelText: $i18n.t('generic.button.cancel'),
      confirmFn() {
        resolve(true);
      },
      cancelFn() {
        reject(false);
      },
    });
  });

  const reset = () => {
    setTimeout(() => {
      isChanged.value = false;
    });
  };

  const setChanged = (v: boolean) => {
    isChanged.value = v;
  };

  return {
    isChanged,
    reset,
    setChanged,
    handleBeforeClose,
  };
}
