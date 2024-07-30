// 服务实例的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';

export default defineStore('script', () => {
  // 脚本配置页面是否需要打开编辑版本面板
  const versionListPageShouldOpenEdit = ref(false);
  // 脚本配置页面是否需要打开查看版本面板
  const versionListPageShouldOpenView = ref(false);
  // 不含禁用的列表数据总数
  const filterDisableCount = ref(0);

  return { versionListPageShouldOpenEdit, versionListPageShouldOpenView, filterDisableCount };
});
