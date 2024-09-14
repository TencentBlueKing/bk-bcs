<template>
  <div v-if="!loading">
    <!-- 去审批抽屉 -->
    <PublishVersionDiff
      :bk-biz-id="spaceId"
      :app-id="appId"
      :show="show"
      :current-version="versionData"
      :base-version-id="baseVersionId"
      :version-list="diffableVersionList()"
      :current-version-groups="groupsPendingtoPublish()"
      @publish="emits('update:show', false)"
      @close="emits('update:show', false)" />
  </div>
</template>

<script setup lang="ts">
  // import { storeToRefs } from 'pinia';
  import { ref, watch, computed } from 'vue';
  import PublishVersionDiff from '../../service/detail/components/publish-version-diff.vue';
  import { getServiceGroupList } from '../../../../api/group';
  import { getConfigVersionList } from '../../../../api/config';
  import { getAppDetail } from '../../../../api';
  // import { IGroupToPublish, IGroupItemInService } from '../../../../../types/group';
  import { IGroupItemInService } from '../../../../../types/group';
  import { IConfigVersion, IReleasedGroup } from '../../../../../types/config';
  import { GET_UNNAMED_VERSION_DATA } from '../../../../constants/config';
  import useServiceStore from '../../../../store/service';

  const props = withDefaults(
    defineProps<{
      show: boolean;
      spaceId: string;
      appId: number;
      releasesId: number;
      releasedGroups: number[];
    }>(),
    {},
  );

  const emits = defineEmits(['update:show']);

  const serviceStore = useServiceStore();

  const loading = ref(true);
  // --
  const versionData = ref<IConfigVersion>(GET_UNNAMED_VERSION_DATA());
  const versionList = ref<IConfigVersion[]>([]);
  const baseVersionId = ref(0);
  const groupList = ref<IGroupItemInService[]>([]);
  // --

  watch(
    () => props.show,
    (newV) => {
      newV ? init() : serviceStore.$reset();
    },
  );

  const init = async () => {
    console.log('init 111');
    try {
      loading.value = true;
      const [versionRes, groupRes, appDetailRes] = await Promise.all([
        getConfigVersionList(props.spaceId, props.appId, { start: 0, all: true }), // 所有版本
        getServiceGroupList(props.spaceId, props.appId), // 所有分组
        getAppDetail(props.spaceId, props.appId), // 服务详情数据
      ]);
      // 已上线版本列表
      versionList.value = versionRes.data.details.filter((item: IConfigVersion) => {
        const { id, status } = item;
        return id !== props.releasesId && status.publish_status !== 'not_released';
      });
      // 存放当前版本数据
      const currVersion = versionRes.data.details.find((item: IConfigVersion) => item.id === props.releasesId);
      if (currVersion !== undefined) {
        versionData.value = currVersion;
      }
      // 处理全部分组
      groupList.value = groupRes.details;
      // console.log(props.releasesId, versionList.value[0].id, '是否找到版本');
      // 当前的服务数据
      serviceStore.$patch((state) => {
        state.appData = appDetailRes;
      });
      loading.value = false;
    } catch (e) {
      console.log(e);
    }
    // console.log(versionList.value, '所有已在线版本列表');
    // console.log(versionData.value, 'curr版本列表');
    // console.log(props.releasedGroups, '待上线版本的分组集合');
    // console.log(groupList.value, '所有分组');
  };

  // 审批 S
  // 需要对比的分组id集合
  const releasedGroups = computed(() => {
    console.log('releasedGroups 222');
    if (!props.releasedGroups.length) {
      // 全部分组上线
      return groupList.value
        .filter((group) => versionList.value.some((version) => group.release_id === version.id))
        .map((group) => group.group_id);
    }
    // 指定分组上线
    return props.releasedGroups;
  });

  // 需要对比的线上版本集合
  const diffableVersionList = () => {
    console.log('diffableVersionList 333');
    const list = [] as IConfigVersion[];
    versionList.value.forEach((version) => {
      version.status.released_groups.some((item) => {
        console.log(item.id, 'item-id');
        if (releasedGroups.value.includes(item.id)) {
          list.push(version);
        }
        return false;
      });
      if (version.status.fully_released && props.releasedGroups.length) {
        list.push(version);
      }
    });
    return list;
  };

  /**
   * @description 直接上线或先对比再上线
   * 所有分组都为首次上线，则直接上线，反之先对比再上线
   */
  //   const handlePublishOrOpenDiff = () => {
  //     if (diffableVersionList.value.length) {
  //       baseVersionId.value = diffableVersionList.value[0].id;
  //       isDiffSliderShow.value = true;
  //       return;
  //     }
  //   };
  // 待上线分组实例和数据组装
  const groupsPendingtoPublish = () => {
    console.log('groupsPendingtoPublish 444');
    const list: IReleasedGroup[] = [];
    // console.log(diffableVersionList.value, '待上线版本++++++++++++++++++++++++++++++++++++');
    releasedGroups.value.forEach((groupId) => {
      const group = groupList.value.find((g) => g.group_id === groupId);
      if (group) {
        const { group_id, group_name, new_selector, old_selector, edited } = group;
        list.push({
          id: group_id,
          name: group_name,
          new_selector,
          old_selector,
          edited,
          uid: '',
          mode: '',
        });
      }
    });
    // 全部分组上线(当前操作记录接口全部分组上线时 数组返回为空。对比组件做法：在已上线的分组基础上 添加一个id为0的分组实例)
    if (!props.releasedGroups.length) {
      list.push({
        id: 0,
        name: '全部实例',
        new_selector: {},
        old_selector: {},
        edited: false,
        uid: '',
        mode: '',
      });
    }
    return list;
  };
  // 审批 E
</script>

<style lang="scss" scoped></style>
