import { ref } from 'vue';

export type Preset = 'QA' | 'KubernetesProfessor';

const isShowAssistant = ref(false);

export default function useAssistantStore() {
  function toggleAssistant(isShow: boolean) {
    isShowAssistant.value = isShow;
  }

  return {
    isShowAssistant,
    toggleAssistant,
  };
}
