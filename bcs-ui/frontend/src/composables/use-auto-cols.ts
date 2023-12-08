import { throttle } from 'lodash';
import { onBeforeUnmount, onMounted, Ref, ref } from 'vue';;

interface IRange {
  min: number
  max: number
  cols: number
}
// 自适应布局
export default function useAutoCols(el: Ref<any>, range: IRange[]) {
  const cols = ref(1);
  const setCols = throttle(() => {
    const width = el.value?.clientWidth || 0;
    const item = range.find(item => width <= item.max && width > item.min);
    cols.value = item?.cols || 1;
  }, 200);
  const observer = new ResizeObserver(() => {
    window.requestAnimationFrame(() => {
      setCols();
    });
  });

  onMounted(() => {
    observer.observe(el.value);
  });

  onBeforeUnmount(() => {
    observer.disconnect();
  });

  return {
    cols,
  };
}
