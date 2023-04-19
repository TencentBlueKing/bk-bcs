// 服务实例的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";

interface IAppData {
  id: number|string;
  spec: {
      name: string;
  }
}

export const useServiceStore = defineStore('service', () => {
  // 服务详情数据
  const appData = ref<IAppData>({
    id: '',
    spec: {
      name: ''
    }
  })

  return { appData }
})
