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
      title: $i18n.t('确认离开当前页?'),
      subTitle: $i18n.t('离开将会导致未保存信息丢失'),
      clsName: 'custom-info-confirm default-info',
      okText: $i18n.t('离开'),
      cancelText: $i18n.t('取消'),
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
