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
        {{ isDefaultGroupReleasedOnCrtVersion ? t('全部实例已上线') : t('暂无预览') }}
        <p>
          {{
            isDefaultGroupReleasedOnCrtVersion
              ? t('除以下分组之外的所有实例已上线当前版本')
              : t('请先从左侧选择待上线的分组范围')
          }}
        </p>
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
  import { aggregatePreviewData, aggregateExcludedData } from '../../hooks/aggegate-groups';
  import PreviewSectionItem from './preview-section-item.vue';

  const versionStore = useConfigStore();
  const { versionData } = storeToRefs(versionStore);
  const { t } = useI18n();

  const defaultGroup = computed(() => {
    return props.groupList.find((group) => group.id === 0);
  });

  // 全部实例分组是否已上线在当前版本
  const isDefaultGroupReleasedOnCrtVersion = computed(() => {
    return defaultGroup.value && defaultGroup.value.release_id === versionData.value.id;
  });

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
      previewData.value = aggregatePreviewData(
        props.value,
        props.groupList,
        props.releasedGroups,
        props.releaseType,
        versionData.value.id,
      );
      excludeData.value = aggregateExcludedData(props.value, props.groupList, props.releaseType, versionData.value.id);
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
      display: inline-flex;
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
