<template>
  <div v-show="selectType !== CheckType.Uncheck" class="selections-style">
    {{ t('已选择', { count: isFullDataMode ? selectionsLength : showSelectedLength }) }},
    <span
      class="checked-em"
      v-show="(isFullDataMode ? selectionsLength : showSelectedLength) < dataLength"
      @click="selectTypeChange">
      {{ t('选择所有', { count: dataLength }) }}
    </span>
    <span
      class="checked-em"
      v-show="(isFullDataMode ? selectionsLength : showSelectedLength) === dataLength"
      @click="clearSelection">
      {{ t('取消选择所有数据') }}
    </span>
  </div>
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import CheckType from '../../types/across-checked';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      dataLength: number;
      selectionsLength: number;
      isFullDataMode: boolean;
      selectType: number;
      crossPageSelect: boolean;
      handleSelectTypeChange: Function;
      handleClearSelection: Function;
    }>(),
    {
      dataLength: 0,
      selectionsLength: 0,
      isFullDataMode: false,
      selectType: CheckType.Uncheck,
      crossPageSelect: true,
      handleSelectTypeChange: () => {},
      handleClearSelection: () => {},
    },
  );
  // 已选择数据长度展示
  const showSelectedLength = computed(() => {
    const { selectType, selectionsLength, dataLength } = props;
    return [CheckType.HalfChecked, CheckType.Checked].includes(selectType)
      ? selectionsLength
      : dataLength - selectionsLength;
  });

  // 根据是否提供全选/跨页全选功能，判断当前页全选/跨页全选
  const selectTypeChange = () => {
    if (props.crossPageSelect) {
      // 跨页全选
      props.handleSelectTypeChange(CheckType.AcrossChecked);
    } else {
      // 当前页全选
      props.handleSelectTypeChange(CheckType.Checked);
    }
  };
  const clearSelection = () => {
    props.handleClearSelection();
  };
</script>
<style lang="scss" scoped>
  .selections-style {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 30px;
    font-size: 12px;
    color: #63656e;
    background: #ebecf0;
    .checked-number {
      padding: 0 5px;
      font-weight: 700;
    }
    .checked-em {
      margin-left: 5px;
      color: #3a84ff;
      cursor: pointer;
      &:hover {
        color: #699df4;
      }
    }
  }
</style>
