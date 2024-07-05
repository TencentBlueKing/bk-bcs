<template>
  <template v-if="isFullDataMode">
    <div v-show="selectionsLength" class="selections-style">
      {{ t('已选择') }}
      <span class="checked-number">{{ selectionsLength }}</span>
      <template v-if="locale === 'zh-cn'">条数据</template>
      <span class="checked-em" v-show="!(selectionsLength === dataLength)" @click="selectTypeChange">
        {{ t('选择所有') }} {{ dataLength }}
        <template v-if="locale === 'zh-cn'">条</template>
      </span>
      <span class="checked-em" v-show="selectionsLength === dataLength" @click="clearSelection">
        {{ t('取消选择所有数据') }}
      </span>
    </div>
  </template>
  <template v-else>
    <div v-show="selectType !== CheckType.Uncheck" class="selections-style">
      {{ t('已选择') }}
      <span class="checked-number">
        {{ showSelectedLength }}
      </span>
      <template v-if="locale === 'zh-cn'">条数据</template>
      <span class="checked-em" v-show="!(showSelectedLength === dataLength)" @click="selectTypeChange">
        {{ t('选择所有') }} {{ dataLength }}
        <template v-if="locale === 'zh-cn'">条</template>
      </span>
      <span class="checked-em" v-show="showSelectedLength === dataLength" @click="clearSelection">
        {{ t('取消选择所有数据') }}
      </span>
    </div>
  </template>
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import CheckType from '../../types/across-checked';
  import { useI18n } from 'vue-i18n';

  const { t, locale } = useI18n();

  const props = defineProps({
    dataLength: {
      // 不含禁用的数据总数
      type: Number,
      default: 0,
    },
    selectionsLength: {
      type: Number,
      default: 0,
    },
    isFullDataMode: {
      // 是否全量数据模式
      type: Boolean,
      default: false,
    },
    selectType: {
      type: Number,
      default: CheckType.Uncheck,
    },
    arrowShow: {
      type: Boolean,
      default: true,
    },
    handleSelectTypeChange: {
      type: Function,
      default: () => {},
    },
    handleClearSelection: {
      type: Function,
      default: () => {},
    },
  });

  // 已选择数据长度展示
  const showSelectedLength = computed(() => {
    const { selectType, selectionsLength, dataLength } = props;
    return [CheckType.HalfChecked, CheckType.Checked].includes(selectType)
      ? selectionsLength
      : dataLength - selectionsLength;
  });

  // 根据是否提供全选/跨页全选功能，判断当前页全选/跨页全选
  const selectTypeChange = () => {
    if (props.arrowShow) {
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
