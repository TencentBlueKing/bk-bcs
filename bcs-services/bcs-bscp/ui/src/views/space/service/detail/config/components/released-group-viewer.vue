<template>
  <bk-popover
    ext-cls="released-group-viewer"
    theme="light"
    placement="bottom-end"
    :disabled="props.disabled"
    @after-show="popoverOpen">
    <slot name="default"></slot>
    <template #content>
      <bk-loading :loading="loading" :opacity="1" class="groups-content-wrapper">
        <div class="header-area">
          <h3 class="title">已上线实例</h3>
          <template v-if="props.isDefaultGroup">
            <div class="default-group-name">默认分组</div>
            <div class="tips">除以下分组之外的所有实例</div>
          </template>
        </div>
        <div class="group-list">
          <div v-for="group in groupList" class="group-item">
            <div class="group-name">{{ group.name }}</div>
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
  import { ref, withDefaults } from 'vue';
  import { getServiceGroupList } from '../../../../../../api/group';
  import { IReleasedGroup } from "../../../../../../../types/config";
  import { IGroupItemInService } from '../../../../../../../types/group';
  import RuleTag from '../../../../groups/components/rule-tag.vue';

  const props = withDefaults(defineProps<{
    bkBizId: string;
    appId: number;
    isDefaultGroup?: boolean;
    disabled?: boolean;
    groups: IReleasedGroup[];
  }>(), {
    groups: () => [],
  });

  const loading = ref(false);
  const groupList = ref<IReleasedGroup[]>([]);

  const popoverOpen = () => {
    if (props.isDefaultGroup) {
      getExcludeGroups();
    } else {
      groupList.value = props.groups;
    }
  }

// 获取默认分组下的排除分组
const getExcludeGroups = async () => {
  loading.value = true;
  const res = await getServiceGroupList(props.bkBizId, props.appId);
  groupList.value = res.details
    .filter((item: IGroupItemInService) => item.group_id > 0 && item.release_id > 0)
    .map((item: IGroupItemInService) => {
      return { ...item, name: item.group_name, id: item.group_id };
    });
  loading.value = false;
};

</script>
<style scoped lang="scss">
  .groups-content-wrapper {
    min-width: 220px;
    font-size: 12px;
    line-height: 16px;
    color: #63656e;
    background: #ffffff;
    .header-area {
      padding-bottom: 12px;
      border-bottom: 1px solid #dcdee5;
    }
    .title {
      margin: 0;
      padding: 7px 14px 0;
      font-size: 12px;
      line-height: 20px;
      color: #313238;
    }
    .default-group-name {
      line-height: 20px;
      color: #63656E;
    }
    .tips {
      line-height: 20px;
      color: #979BA5;
    }
    .group-list {
      padding: 12px 14px 7px;
      max-height: 300px;
      max-width: 520px;
      overflow: auto;
    }
    .group-item {
      margin-bottom: 8px;
    }
    .rules {
      margin-top: 4px;
      padding: 5px 8px;
      width: 100%;
      white-space: normal;
      background: #f5f7fa;
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
