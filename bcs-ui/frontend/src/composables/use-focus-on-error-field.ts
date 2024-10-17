import { ref } from 'vue';

export function useFocusOnErrorField() {
  const errorClasses = ref(['form-error-tip', 'error-tips', 'is-error']);

  const focusOnErrorField = async () => {
    // 自动滚动到第一个错误的位置
    const firstErrDom = findFirstErrorElement();
    firstErrDom?.scrollIntoView({
      block: 'center',
      behavior: 'smooth',
    });
  };

  const findFirstErrorElement = () => {
    for (const className of errorClasses.value) {
      const elements = document.getElementsByClassName(className);
      if (elements.length > 0) {
        return elements[0];
      }
    }
    return null;
  };

  return {
    focusOnErrorField,
  };
}