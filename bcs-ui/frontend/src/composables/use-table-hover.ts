import { copyText } from '@/common/util';
import { ref } from 'vue';
import $i18n from '@/i18n/i18n-setup';
import $bkMessage from '@/common/bkmagic';

export default function useTableHover() {
  const ativeIndex = ref(-1);
  const handleMouseEnter = (index) => {
    ativeIndex.value = index;
  };
  const handleMouseLeave = () => {
    ativeIndex.value = -1;
  };
  const handleCopyContent = (value) => {
    copyText(value);
    $bkMessage({
      theme: 'success',
      message: $i18n.t('复制成功'),
    });
  };

  return {
    ativeIndex,
    handleMouseEnter,
    handleMouseLeave,
    handleCopyContent,
  };
}
