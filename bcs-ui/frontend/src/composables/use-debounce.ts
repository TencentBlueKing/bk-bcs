import { customRef } from 'vue';

export default function useDebouncedRef<T>(value, delay = 200) {
  let timeout;
  let innerValue = value;
  return customRef<T>((track, trigger) => ({
    get() {
      track();
      return innerValue;
    },
    set(newValue: any) {
      clearTimeout(timeout);
      if (newValue === undefined || newValue === '') {
        innerValue = newValue;
        trigger();
      } else {
        timeout = setTimeout(() => {
          innerValue = newValue;
          trigger();
        }, delay);
      }
    },
  }));
}
