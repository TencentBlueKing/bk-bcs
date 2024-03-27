<template>
  <bk-popover
    ext-cls="released-group-viewer"
    theme="light"
    :placement="props.placement"
    :disabled="props.disabled"
    @after-show="popoverOpen">
    <slot name="default"></slot>
    <template #content>
      <bk-loading :loading="loading" :opacity="1" class="groups-content-wrapper">
        <div class="header-area">
          <h3 class="title">{{ props.isPending ? t('待上线实例') : t('已上线实例') }}</h3>
          <template v-if="hasDefaultGroup">
            <div class="tips">{{ t('除以下分组之外的所有实例') }}</div>
          </template>
        </div>
        <div class="group-list">
          <div v-for="group in groupList" class="group-item" :key="group.id">
            <div class="group-name">
              <i class="bk-bscp-icon icon-resources-fill" />
              {{ group.name }}
            </div>
            <div class="rules">
              <template v-for="(rule, index) in group.new_selector?.labels_and" :key="index">
                <template v-if="index > 0"> & </template>
                <rule-tag class="tag-item" :rule="rule" />
              </template>
            </div>
          </div>
        </div>
      </bk-loading>
    </template>
  </bk-popover>
</template>
<script setup lang="ts">
  import { ref, withDefaults, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { getServiceGroupList } from '../../../../../../api/group';
  import { IReleasedGroup } from '../../../../../../../types/config';
  import { IGroupItemInService } from '../../../../../../../types/group';
  import RuleTag from '../../../../groups/components/rule-tag.vue';

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      bkBizId: string;
      appId: number;
      disabled?: boolean;
      placement?: string;
      groups: IReleasedGroup[]; // 当前版本上线的分组实例
      isPending?: boolean; // 是否为待上线
    }>(),
    {
      placement: 'bottom-end',
      groups: () => [],
      isPending: false,
    },
  );

  const loading = ref(false);
  const groupList = ref<IReleasedGroup[]>([]);

  const hasDefaultGroup = computed(() => props.groups.some((item) => item.id === 0));

  const popoverOpen = () => {
    if (hasDefaultGroup.value) {
      getExcludeGroups();
    } else {
      groupList.value = props.groups;
    }
  };

  // 获取默认分组下的排除分组
  const getExcludeGroups = async () => {
    loading.value = true;
    const res = await getServiceGroupList(props.bkBizId, props.appId);
    groupList.value = res.details
      .filter((item: IGroupItemInService) => {
        return (
          item.group_id > 0 &&
          item.release_id > 0 &&
          props.groups.findIndex((group) => group.id === item.group_id) === -1
        );
      })
      .map((item: IGroupItemInService) => ({ ...item, name: item.group_name, id: item.group_id }));
    loading.value = false;
  };
</script>
<style scoped lang="scss">
  .groups-content-wrapper {
    min-width: 220px;
    font-size: 12px;
    line-height: 16px;
    max-height: 400px;
    color: #63656e;
    background: #ffffff;
    overflow: auto;
    .header-area {
      margin: 10px 16px 0;
      padding-bottom: 12px;
      border-bottom: 1px solid #dcdee5;
    }
    .title {
      margin: 0;
      font-size: 12px;
      line-height: 20px;
      color: #313238;
    }
    .group-name {
      display: flex;
      align-items: center;
      line-height: 20px;
      color: #63656e;
      .bk-bscp-icon {
        margin-right: 4px;
        color: #979ba5;
        font-size: 16px;
      }
    }
    .tips {
      line-height: 20px;
      color: #979ba5;
    }
    .group-list {
      padding: 12px 16px 7px;
      max-width: 400px;
    }
    .group-item {
      padding-bottom: 12px;
      margin-bottom: 8px;
      border-bottom: 1px solid #dcdee5;
      &:last-child {
        margin-bottom: 0;
        border-bottom: none;
      }
    }
    .rules {
      margin-top: 4px;
      padding: 0 8px;
      width: 100%;
      color: #979ba5;
      white-space: normal;
      .tag-item {
        display: inline;
      }
    }
  }
</style>
<style lang="scss">
  .released-group-viewer.bk-popover.bk-pop2-content {
    padding: 0;
  }
</style>
