import { localT } from '../../../../../../i18n';
import { IGroupToPublish, IGroupPreviewItem } from '../../../../../../../types/group';

const pushItemToAggregateData = (
  group: IGroupToPublish,
  releaseName: string,
  type: string,
  data: IGroupPreviewItem[],
) => {
  const release = data.find((item) => item.id === group.release_id);
  if (release) {
    release.children.push(group);
  } else {
    data.push({
      id: group.release_id,
      name: releaseName,
      type,
      children: [group],
    });
  }
};

// 聚合预览数据
export const aggregatePreviewData = (
  selectedGroups: IGroupToPublish[],
  groupList: IGroupToPublish[],
  releasedIds: number[],
  releaseType: string,
  versionId: number,
) => {
  let list: IGroupPreviewItem[] = [];
  // 首次上线
  // 1. 全部实例分组：当前选中全部实例上线方式，且全部实例分组未在其他线上版本
  // 2. 普通分组：当前选中选择分组上线方式，且全部实例分组未在其他线上版本
  const initialRelease: IGroupPreviewItem = { id: 0, name: localT('首次上线'), type: 'plain', children: [] };
  // 变更版本
  // 1. 全部实例分组：当前选中全部实例上线方式，且全部实例分组已在其他线上版本
  // 2. 普通分组：
  //    a. 当前选中选择分组上线方式，且当前分组在其他线上版本或全部实例分组已在其他线上版本
  //    b. 当前选中全部实例上线方式，且全部实例分组已在线上版本时，取消排除的分组
  const modifyReleases: IGroupPreviewItem[] = [];
  const defaultGroup = groupList.find((item) => item.id === 0) as IGroupToPublish;
  const isDefaultGroupReleased = defaultGroup.release_id > 0;
  const isDefaultGroupReleasedOnCrtVersion = defaultGroup.release_id === versionId;
  selectedGroups
    .filter((group) => !releasedIds.includes(group.id))
    .forEach((group) => {
      // 全部实例分组
      if (group.id === 0) {
        if (releaseType === 'all') {
          if (!isDefaultGroupReleased) {
            // 首次上线：当前选中全部实例上线方式，且全部实例分组未在其他线上版本
            initialRelease.children.push(group);
          } else if (isDefaultGroupReleased && !isDefaultGroupReleasedOnCrtVersion) {
            // 变更版本：当前选中全部实例上线方式，且全部实例分组已在其他线上版本
            pushItemToAggregateData(group, group.release_name, 'modify', modifyReleases);
          }
        }
      } else {
        // 普通分组
        if (releaseType === 'select') {
          if (!isDefaultGroupReleased) {
            // 首次上线：当前选中选择分组上线方式，且全部实例分组未在其他线上版本
            if (group.release_id === 0) {
              initialRelease.children.push(group);
            } else {
              pushItemToAggregateData(group, group.release_name, 'modify', modifyReleases);
            }
          } else if ((group.release_id > 0 && group.release_id !== versionId) || !isDefaultGroupReleasedOnCrtVersion) {
            // 变更版本：当前选中选择分组上线方式，当前分组在其他线上版本或全部实例分组已在其他线上版本
            const name = group.release_id === 0 ? defaultGroup.release_name : group.release_name;
            pushItemToAggregateData(group, name, 'modify', modifyReleases);
          }
        } else if (releaseType === 'all' && group.release_id > 0 && group.release_id !== versionId) {
          // 变更版本：当前选中全部实例上线方式，当前分组在其他线上版本或全部实例分组已在其他线上版本
          const name = group.release_id === 0 ? defaultGroup.release_name : group.release_name;
          pushItemToAggregateData(group, name, 'modify', modifyReleases);
        }
      }
    });
  list.push(...modifyReleases);
  if (initialRelease.children.length > 0) {
    list.unshift(initialRelease);
  }
  // 按照版本id倒序排列
  list = list.sort((crt, next) => next.id - crt.id);
  const includesDefaultGroupIndex = list.findIndex((item) => item.children.some((g) => g.id === 0));
  // 将全部实例置顶
  if (includesDefaultGroupIndex > -1) {
    const previewData = list[includesDefaultGroupIndex];
    // 全部实例增加说明
    const defaultGroup = previewData.children.find((item) => item.id === 0);
    if (defaultGroup) {
      defaultGroup.desc = list.length > 1 ? localT('（已上线分组之外的所有实例）') : '';
    }
    list.splice(includesDefaultGroupIndex, 1);
    list.unshift(previewData);
  }
  return list;
};

// 聚合排除数据
export const aggregateExcludedData = (
  selectedGroups: IGroupToPublish[],
  groupList: IGroupToPublish[],
  releaseType: string,
  versionId: number,
) => {
  const list: IGroupPreviewItem[] = [];
  if (releaseType === 'all') {
    const groupsOnOtherRelease = groupList.filter(
      (group) =>
        group.release_id > 0 &&
        group.release_id !== versionId &&
        selectedGroups.findIndex((item) => item.id === group.id) === -1,
    );
    groupsOnOtherRelease.forEach((group) => {
      pushItemToAggregateData(group, group.release_name, 'retain', list);
    });
  }
  // 按照版本id倒序排列
  return list.sort((crt, next) => next.id - crt.id);
};
