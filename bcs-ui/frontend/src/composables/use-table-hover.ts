import { ref } from 'vue';

import $bkMessage from '@/common/bkmagic';
import { copyText } from '@/common/util';
import $i18n from '@/i18n/i18n-setup';

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
      message: $i18n.t('generic.msg.success.copy'),
    });
  };

  return {
    ativeIndex,
    handleMouseEnter,
    handleMouseLeave,
    handleCopyContent,
  };
}
