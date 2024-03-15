<template>
  <div class="preview-panel">
    <h3 class="title">
      {{ t('上线预览') }}
      <span class="tips">
        {{ t('上线后，相关分组的') }}
        <span class="bold">{{ t('实例') }}</span>
        {{ t('将从以下各版本更新至当前版本') }}
      </span>
    </h3>
    <bk-exception v-if="previewData.length === 0" scene="part" type="empty">
      <div class="empty-tips">
        {{ t('暂无预览') }}
        <p>{{ t('请先从左侧选择待上线的分组范围') }}</p>
      </div>
    </bk-exception>
    <template v-else>
      <preview-section-item
        v-for="previewGroup in previewData"
        section-type="diff"
        :key="previewGroup.id"
        :preview-group="previewGroup"
        :release-type="props.releaseType"
        :released-groups="props.releasedGroups"
        :value="props.value"
        @diff="emits('diff', $event)"
        @change="emits('change', $event)">
      </preview-section-item>
    </template>
    <template v-if="excludeData.length > 0">
      <div class="split-line"></div>
      <h3 class="title">
        {{ t('已排除分组实例') }}
        <span class="tips">
          {{ t('本次上线版本对以下分组实例') }}<span class="bold">{{ t('不会产生影响') }}</span>
        </span>
      </h3>
      <preview-section-item
        v-for="group in excludeData"
        section-type="exclude"
        :key="group.id"
        :preview-group="group"
        :release-type="props.releaseType"
        :released-groups="props.releasedGroups"
        :value="props.value"
        @diff="emits('diff', $event)"
        @change="emits('change', $event)">
      </preview-section-item>
    </template>
  </div>
</template>
<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IGroupToPublish, IGroupPreviewItem } from '../../../../../../../../types/group';
  import { storeToRefs } from 'pinia';
  import useConfigStore from '../../../../../../../store/config';
  import PreviewSectionItem from './preview-section-item.vue';

  const versionStore = useConfigStore();
  const { versionData } = storeToRefs(versionStore);
  const { t } = useI18n();

  const defaultGroup = computed(() => {
    return props.groupList.find((group) => group.id === 0);
  });

  // 全部实例分组是否已上线
  const isDefaultGroupReleased = computed(() => {
    return defaultGroup.value && defaultGroup.value.release_id > 0;
  });

  // 全部实例分组是否已上线在当前版本
  const isDefaultGroupReleasedOnCrtVersion = computed(() => {
    return defaultGroup.value && defaultGroup.value.release_id === versionData.value.id;
  });

  // 聚合预览数据
  const aggregatePreviewData = () => {
    const list: IGroupPreviewItem[] = [];
    // 首次上线
    // 1. 全部实例分组：当前选中全部实例上线方式，且全部实例分组未在其他线上版本
    // 2. 普通分组：当前选中选择分组上线方式，且全部实例分组未在其他线上版本
    const initialRelease: IGroupPreviewItem = { id: 0, name: t('首次上线'), type: 'plain', children: [] };
    // 变更版本
    // 1. 全部实例分组：当前选中全部实例上线方式，且全部实例分组已在其他线上版本
    // 2. 普通分组：
    //    a. 当前选中选择分组上线方式，且当前分组在其他线上版本或全部实例分组已在其他线上版本
    //    b. 当前选中全部实例上线方式，且全部实例分组已在线上版本时，取消排除的分组
    const modifyReleases: IGroupPreviewItem[] = [];
    props.value
      .filter((group) => !props.releasedGroups.includes(group.id))
      .forEach((group) => {
        // 全部实例分组
        if (group.id === 0) {
          if (props.releaseType === 'all') {
            if (!isDefaultGroupReleased.value) {
              // 首次上线：当前选中全部实例上线方式，且全部实例分组未在其他线上版本
              initialRelease.children.push(group);
            } else if (isDefaultGroupReleased.value && !isDefaultGroupReleasedOnCrtVersion.value) {
              // 变更版本：当前选中全部实例上线方式，且全部实例分组已在其他线上版本
              pushItemToAggegateData(group, group.release_name, 'modify', modifyReleases);
            }
          }
        } else {
          // 普通分组
          if (props.releaseType === 'select') {
            if (!isDefaultGroupReleased.value) {
              // 首次上线：当前选中选择分组上线方式，且全部实例分组未在其他线上版本
              if (group.release_id === 0) {
                initialRelease.children.push(group);
              } else {
                pushItemToAggegateData(group, group.release_name, 'modify', modifyReleases);
              }
            } else if (
              (group.release_id > 0 && group.release_id !== versionData.value.id) ||
              !isDefaultGroupReleasedOnCrtVersion.value
            ) {
              // 变更版本：当前选中选择分组上线方式，当前分组在其他线上版本或全部实例分组已在其他线上版本
              const name =
                group.release_id === 0 ? (defaultGroup.value as IGroupToPublish).release_name : group.release_name;
              pushItemToAggegateData(group, name, 'modify', modifyReleases);
            }
          } else if (props.releaseType === 'all' && group.release_id > 0 && group.release_id !== versionData.value.id) {
            // 变更版本：当前选中全部实例上线方式，当前分组在其他线上版本或全部实例分组已在其他线上版本
            const name =
              group.release_id === 0 ? (defaultGroup.value as IGroupToPublish).release_name : group.release_name;
            pushItemToAggegateData(group, name, 'modify', modifyReleases);
          }
        }
      });
    list.push(...modifyReleases);
    if (initialRelease.children.length > 0) {
      list.unshift(initialRelease);
    }
    previewData.value = list;
  };

  // 聚合排除数据
  const aggregateExcludedData = () => {
    const list: IGroupPreviewItem[] = [];
    if (props.releaseType === 'all') {
      const groupsOnOtherRelease = props.groupList.filter(
        (group) =>
          group.release_id > 0 &&
          group.release_id !== versionData.value.id &&
          props.value.findIndex((item) => item.id === group.id) === -1,
      );
      groupsOnOtherRelease.forEach((group) => {
        pushItemToAggegateData(group, group.release_name, 'retain', list);
      });
    }
    excludeData.value = list;
  };

  const pushItemToAggegateData = (
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

  const props = withDefaults(
    defineProps<{
      groupList: IGroupToPublish[];
      releaseType: string;
      releasedGroups?: number[]; // 已上线版本的分组
      value: IGroupToPublish[]; // 当前选中的分组
    }>(),
    {
      releasedGroups: () => [],
    },
  );

  const emits = defineEmits(['diff', 'change']);

  const previewData = ref<IGroupPreviewItem[]>([]);
  const excludeData = ref<IGroupPreviewItem[]>([]);

  watch(
    () => props.value,
    () => {
      aggregatePreviewData();
      aggregateExcludedData();
    },
    { immediate: true },
  );
</script>
<style lang="scss" scoped>
  .preview-panel {
    height: 100%;
    overflow: auto;
  }
  .title {
    margin: 0 0 16px;
    padding: 0 24px;
    line-height: 19px;
    font-size: 14px;
    font-weight: 700;
    color: #63656e;
    .tips {
      margin-left: 16px;
      line-height: 20px;
      color: #979ba5;
      font-size: 12px;
      font-weight: 400;
      .bold {
        font-weight: 700;
      }
    }
  }
  .empty-tips {
    font-size: 14px;
    color: #63656e;
    & > p {
      margin: 8px 0 0;
      color: #979ba5;
      font-size: 12px;
    }
  }
  .split-line {
    margin: 32px 24px 8px;
    height: 1px;
    background: #dcdee5;
  }
</style>
