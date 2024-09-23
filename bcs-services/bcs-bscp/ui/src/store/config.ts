// 配置管理的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IConfigVersion } from '../../types/config';
import { GET_UNNAMED_VERSION_DATA } from '../constants/config';

export default defineStore('config', () => {
  // 非套餐配置和模板配置文件总数量 kv服务非套餐配置总数
  const allConfigCount = ref(0);

  // 非套餐配置和模板配置文件总数量 kv服务非套餐配置总数 （除去删除状态总数）
  const allExistConfigCount = ref(0);

  // 当前选中版本, 用id为0表示未命名版本
  const versionData = ref<IConfigVersion>(GET_UNNAMED_VERSION_DATA());

  // 上线版本ID, 上线成功切换到刚刚上线的版本
  const publishedVersionId = ref(0);

  // 是否为版本详情视图，切换服务后保持当前视图
  const versionDetailView = ref(false);

  // 是否需要刷新版本列表标识，配置生成版本、发布版本、调整分组上线之后需要更新版本列表
  const refreshVersionListFlag = ref(false);

  // 非配置模板中存在的冲突文件个数
  const conflictFileCount = ref(0);

  // 是否只查看冲突配置项
  const onlyViewConflict = ref(false);

  return {
    allConfigCount,
    versionData,
    versionDetailView,
    refreshVersionListFlag,
    publishedVersionId,
    conflictFileCount,
    allExistConfigCount,
    onlyViewConflict,
  };
});
