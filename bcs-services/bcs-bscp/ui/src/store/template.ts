// 服务实例的pinia数据
import { ref } from 'vue';
import { defineStore } from 'pinia';
import { ITemplatePackageItem, ITemplateSpaceItem } from '../../types/template';

export default defineStore('template', () => {
  // 模板空间列表
  const templateSpaceList = ref<ITemplateSpaceItem[]>([]);
  // 当前模板空间
  const currentTemplateSpace = ref();
  // 套餐列表
  const packageList = ref<ITemplatePackageItem[]>([]);
  // 模板空间下所有配置文件模板数量
  const CountOfAllTemplatesInSpace = ref(0);
  // 未指定套餐的配置文件数量
  const countOfTemplatesForNoSpecifiedPackage = ref(0);
  // 当前套餐
  const currentPkg = ref<string | number>('');
  // 标识是否需要刷新左侧套餐菜单栏是否需要刷新
  const needRefreshMenuFlag = ref(false);
  // 配置页面是否需要打开编辑版本面板
  const versionListPageShouldOpenEdit = ref(false);
  // 配置页面是否需要打开查看版本面板
  const versionListPageShouldOpenView = ref(false);
  // 置顶id
  const topIds = ref<number[]>([]);
  // 表格是否跨页选择
  const isAcrossChecked = ref(false);
  // 表格数据总数
  const dataCount = ref(0);

  return {
    templateSpaceList,
    currentTemplateSpace,
    packageList,
    currentPkg,
    CountOfAllTemplatesInSpace,
    countOfTemplatesForNoSpecifiedPackage,
    needRefreshMenuFlag,
    versionListPageShouldOpenEdit,
    versionListPageShouldOpenView,
    topIds,
    isAcrossChecked,
    dataCount,
  };
});
