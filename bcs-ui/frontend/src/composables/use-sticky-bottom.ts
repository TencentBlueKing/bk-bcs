import { onBeforeUnmount, onMounted, ref } from 'vue';

/**
 * 吸底效果 hooks
 * @param {string} stickyClass - 吸底时应用的 CSS 类名
 * @returns {Object} 返回 isSticky 状态和元素引用
 */
export function useStickyBottom(stickyClass = 'sticky bottom-0 bg-white py-[10px] z-[999] border-t border-[#e4e4e4]') {
  const targetRef = ref<any>(null);
  const isSticky = ref(false);
  const observer = ref<IntersectionObserver | null>(null);
  const sentinel = ref<HTMLElement | null>(null);

  onMounted(() => {
    console.log('@@@@@@@', targetRef.value);
    if (!targetRef.value) return;

    const element = targetRef.value?.$el || targetRef.value;
    // 创建并插入伪元素（占位元素）
    sentinel.value = document.createElement('div');
    sentinel.value.style.height = '1px';
    sentinel.value.style.width = '100%';
    element.parentNode?.insertBefore(sentinel.value, element.nextSibling);

    // 初始化Intersection Observer
    observer.value = new IntersectionObserver(
      ([entry]) => {
        isSticky.value = !entry.isIntersecting;
        console.log('@@@@@@@', isSticky.value);
      },
      {
        root: null,
        rootMargin: '0px',
        threshold: 1.0,
      },
    );
    observer.value.observe(sentinel.value);
  });

  onBeforeUnmount(() => {
    if (observer.value) {
      observer.value.disconnect();
    }
    if (sentinel.value) {
      sentinel.value.remove();
    }
  });

  return {
    targetRef,
    isSticky,
    stickyClass,
  };
}
