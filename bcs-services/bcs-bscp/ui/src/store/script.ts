// 服务实例的pinia数据
import { ref } from 'vue'
import { defineStore } from "pinia";

export const useScriptStore = defineStore('script', () => {
  // 脚本配置页面是否需要打开编辑版本面板
  const versionListPageShouldOpenEdit = ref(false)

  return { versionListPageShouldOpenEdit }
})