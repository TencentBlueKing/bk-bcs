import InfoBox from 'bkui-vue/lib/info-box';

const useModalCloseConfirmation = (title?: string, subTitle?: string) => new Promise((resolve) => {
  InfoBox({
    title: title || '确认离开当前页？',
    subTitle: subTitle || '离开会导致未保存信息丢失',
    confirmText: '离开',
    onConfirm: () => {
      resolve(true);
    },
    onClosed: () => {
      resolve(false);
    },
  });
});

export default useModalCloseConfirmation;
