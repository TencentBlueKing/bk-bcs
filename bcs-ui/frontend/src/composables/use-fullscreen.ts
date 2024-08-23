import { onBeforeMount, onMounted, ref } from 'vue';

export default function useFullScreen() {
// 全屏
  const contentRef = ref();
  const isFullscreen = ref(false);
  const handleFullscreenChange = () => {
    isFullscreen.value = !!document.fullscreenElement;
  };
  const switchFullScreen = () => {
    if (!document.fullscreenElement) {
      contentRef.value?.requestFullscreen();
    } else {
      document.exitFullscreen();
    }
  };

  onMounted(() => {
    document.addEventListener('fullscreenchange', handleFullscreenChange);
  });

  onBeforeMount(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange);
  });

  return {
    contentRef,
    isFullscreen,
    switchFullScreen,
  };
}
