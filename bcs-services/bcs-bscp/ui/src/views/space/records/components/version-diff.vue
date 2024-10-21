<template>
  <div v-if="!loading">
    <!-- 去审批抽屉 -->
    <PublishVersionDiff
      :btn-loading="btnLoading"
      :is-approval-mode="true"
      :bk-biz-id="spaceId"
      :app-id="appId"
      :show="show"
      :current-version="versionData"
      :base-version-id="baseVersionId"
      :version-list="diffableVersionList()"
      :current-version-groups="groupsPendingtoPublish()"
      @publish="handleConfirm"
      @reject="RejectDialogShow = true"
      @close="handleClose" />
  </div>
  <DialogReject
    v-model:show="RejectDialogShow"
    :space-id="spaceId"
    :app-id="appId"
    :release-id="versionData.id"
    :release-name="versionData.spec.name"
    @reject="handleReject" />
</template>

<script setup lang="ts">
  import { ref, watch, computed } from 'vue';
  import PublishVersionDiff from '../../service/detail/components/publish-version-diff.vue';
  import { getServiceGroupList } from '../../../../api/group';
  import { getConfigVersionList } from '../../../../api/config';
  import { getAppDetail } from '../../../../api';
  import { approve } from '../../../../api/record';
  import { IGroupItemInService } from '../../../../../types/group';
  import { IConfigVersion, IReleasedGroup } from '../../../../../types/config';
  import { GET_UNNAMED_VERSION_DATA } from '../../../../constants/config';
  import { APPROVE_STATUS } from '../../../../constants/record';
  import useServiceStore from '../../../../store/service';
  import DialogReject from './dialog-reject.vue';
  import BkMessage from 'bkui-vue/lib/message';
  import { useI18n } from 'vue-i18n';

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    appId: number;
    releaseId: number;
    releasedGroups: number[];
  }>();

  const emits = defineEmits(['update:show', 'close']);

  const { t } = useI18n();

  const serviceStore = useServiceStore();

  const loading = ref(true);
  const btnLoading = ref(false);
  const versionData = ref<IConfigVersion>(GET_UNNAMED_VERSION_DATA());
  const versionList = ref<IConfigVersion[]>([]);
  const baseVersionId = ref(0);
  const groupList = ref<IGroupItemInService[]>([]);
  const RejectDialogShow = ref(false);

  // 需要对比的分组id集合
  const releasedGroups = computed(() => {
    if (!props.releasedGroups.length) {
      // 全部分组上线
      return groupList.value
        .filter((group) => versionList.value.some((version) => group.release_id === version.id))
        .map((group) => group.group_id);
    }
    // 指定分组上线
    return props.releasedGroups;
  });

  watch(
    () => props.show,
    (newV) => {
      newV ? init() : serviceStore.$reset();
    },
  );

  const init = async () => {
    const { spaceId, appId, releaseId } = props;
    try {
      loading.value = true;
      const [versionRes, groupRes, appDetailRes] = await Promise.all([
        getConfigVersionList(spaceId, appId, { start: 0, all: true }), // 所有版本
        getServiceGroupList(spaceId, appId), // 所有分组
        getAppDetail(spaceId, appId), // 服务详情数据
      ]);
      // 已上线版本列表
      versionList.value = versionRes.data.details.filter((item: IConfigVersion) => {
        const { id, status } = item;
        return id !== releaseId && status.publish_status !== 'not_released';
      });
      // 存放当前版本数据
      const currVersion = versionRes.data.details.find((item: IConfigVersion) => item.id === props.releaseId);
      if (currVersion !== undefined) {
        versionData.value = currVersion;
      }
      // 处理全部分组
      groupList.value = groupRes.details;
      // 当前的服务数据
      serviceStore.$patch((state) => {
        state.appData = appDetailRes;
      });
      loading.value = false;
    } catch (e) {
      console.log(e);
    }
  };

  // 审批 S

  // 需要对比的线上版本集合
  const diffableVersionList = () => {
    const list = [] as IConfigVersion[];
    versionList.value.forEach((version) => {
      version.status.released_groups.some((group) => {
        if (releasedGroups.value.includes(group.id)) {
          list.push(version);
        }
        // 全量分组上线的版本中 含有 待上线版本的的分组, 也需要对比
        if (version.status.fully_released) {
          releasedGroups.value.some((group) => {
            return groupList.value.some((item) => {
              if (item.group_id === group && item.release_id === 0) {
                list.push(version);
                return true;
              }
              return false;
            });
          });
        }
        return false;
      });
    });
    return list;
  };

  // 待上线分组实例和数据组装
  const groupsPendingtoPublish = () => {
    const list: IReleasedGroup[] = [];
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

  // 审批通过
  const handleConfirm = async () => {
    btnLoading.value = true;
    try {
      // await approve(props.spaceId, props.appId, props.releaseId, {
      //   publish_status: APPROVE_STATUS.PendPublish,
      // });
      // BkMessage({
      //   theme: 'success',
      //   message: t('操作成功'),
      // });
      // emits('close', 'refresh');
    } catch (e) {
      console.log(e);
      await approve(props.spaceId, props.appId, props.releaseId, {
        publish_status: APPROVE_STATUS.PendPublish,
      });
      BkMessage({
        theme: 'success',
        message: t('操作成功'),
      });
      emits('close', 'refresh');
    } finally {
      btnLoading.value = true;
    }
  };

  // 审批拒绝
  const handleReject = () => {
    RejectDialogShow.value = false;
    emits('close', 'refresh');
  };

  const handleClose = () => {
    btnLoading.value = false;
    emits('close');
  };
</script>

<style lang="scss" scoped></style>
