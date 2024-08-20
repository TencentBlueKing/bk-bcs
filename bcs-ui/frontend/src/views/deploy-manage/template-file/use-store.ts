import { reactive, set } from 'vue';

import useVariable from '../variable/use-variable';

import { IListTemplateMetadataItem, ITemplateSpaceData } from '@/@types/cluster-resource-patch';
import { TemplateSetService } from '@/api/modules/new-cluster-resource';

export const store = reactive<{
  spaceList: ITemplateSpaceData[]
  spaceLoading: boolean
  loadingSpaceIDs: string[]
  fileListMap: Record<string, IListTemplateMetadataItem[]>
  editMode: 'yaml'|'form'|undefined
  isFormModeDisabled: boolean
  varList: any[],
  showDeployBtn: boolean,
}>({
  spaceLoading: false, // 空间加载状态
  spaceList: [], // 空间列表
  loadingSpaceIDs: [], // 空间下文件加载状态
  fileListMap: {},
  editMode: undefined,
  isFormModeDisabled: false, // 详情模式下是否禁用表单模式
  varList: [],
  showDeployBtn: true, // 是否显示部署按钮
});

// 更新空间列表
export async function updateListTemplateSpaceList() {
  store.spaceLoading = true;
  store.spaceList = await TemplateSetService.ListTemplateSpace().catch(() => []);
  store.spaceLoading = false;
}

// 更新空间下文件列表
export async function updateTemplateMetadataList(spaceID: string) {
  if (!spaceID) return;
  store.loadingSpaceIDs = [spaceID];
  const list = await TemplateSetService.ListTemplateMetadata({
    $templateSpaceID: spaceID,
  });
  store.loadingSpaceIDs = [];
  set(store.fileListMap, spaceID, list);
}

// 更新变量列表
const { getVariableDefinitions } = useVariable();
export async function updateVarList() {
  const { results } = await getVariableDefinitions({
    limit: 0,
    offset: 0,
    all: true,
    scope: '',
    searchKey: '',
  });
  store.varList = results;
}

// 更新是否显示部署按钮
export async function updateShowDeployBtn(state: boolean) {
  store.showDeployBtn = state;
}
