import { ref } from 'vue';

export function useFocusOnErrorField(rootElement?: HTMLElement | null) {
  const dom = rootElement || document;
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
      const elements = dom.getElementsByClassName(className);
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
