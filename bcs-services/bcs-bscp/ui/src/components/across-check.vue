<template>
  <div class="check">
    <bk-checkbox
      v-model="localValue"
      :indeterminate="indeterminate"
      :class="{
        'across-all-check': localValue === CheckType.AcrossChecked,
        'half-indeterminate': localValue === CheckType.HalfAcrossChecked,
      }"
      :disabled="disabled"
      :true-label="localValue === CheckType.AcrossChecked ? CheckType.AcrossChecked : CheckType.Checked"
      :false-label="setCheckBoxStatus"
      @change="handleCheckChange">
    </bk-checkbox>
    <bk-popover
      ref="popoverRef"
      theme="light"
      trigger="click"
      placement="bottom"
      :arrow="false"
      :offset="{ mainAxis: 10, crossAxis: 30 }"
      :disabled="disabled"
      @after-show="isDropDownShow = true"
      @after-hidden="isDropDownShow = false">
      <angle-down
        v-show="crossPageSelect"
        :class="['check-icon', { 'icon-angle-up': isDropDownShow }, { disabled: disabled }]" />
      <template #content>
        <ul class="dropdown-ul">
          <li
            class="dropdown-li"
            v-for="item in checkTypeList"
            :key="item.id"
            :class="{ active: localValue === item.id }"
            @click="handleCheckAll(item.id)">
            {{ item.name }}
          </li>
        </ul>
      </template>
    </bk-popover>
  </div>
</template>
<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import CheckType from '../../types/across-checked';
  import { AngleDown } from 'bkui-vue/lib/icon';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      value: number;
      disabled: boolean;
      crossPageSelect: boolean;
    }>(),
    {
      value: CheckType.Uncheck,
      disabled: false,
      crossPageSelect: true,
    },
  );

  const emits = defineEmits(['change']);

  const popoverRef = ref();
  const isDropDownShow = ref(false);
  const localValue = ref(0);
  const checkTypeList = ref([
    {
      id: CheckType.Checked,
      name: t('全选'),
    },
    {
      id: CheckType.AcrossChecked,
      name: t('跨页全选'),
    },
  ]);

  // 跨页半选和跨页全选
  // const allChecked = computed(() => [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(localValue.value));
  // 当前页半选和跨页半选
  const indeterminate = computed(() => [CheckType.HalfChecked, CheckType.HalfAcrossChecked].includes(localValue.value));
  // 当前页全选和跨页全选
  // const isChecked = computed(() => [CheckType.Checked, CheckType.AcrossChecked].includes(localValue.value));

  const setCheckBoxStatus = computed(() => {
    const status = props.value;
    switch (status) {
      case CheckType.HalfAcrossChecked:
        return CheckType.HalfAcrossChecked; // 跨页半选
      case CheckType.HalfChecked:
        return CheckType.HalfChecked; // 当前页半选
      case CheckType.Checked:
      case CheckType.AcrossChecked:
      default:
        return CheckType.Uncheck;
    }
  });

  watch(
    () => props.value,
    (newV) => {
      localValue.value = newV;
    },
  );

  const handleCheckChange = (id: number) => {
    localValue.value = id;
    emits('change', id);
  };
  const handleCheckAll = (id: number) => {
    handleCheckChange(id);
    popoverRef.value.hide();
  };
</script>
<style lang="scss" scoped>
  .check {
    text-align: left;
    .across-all-check {
      :deep(.bk-checkbox-input) {
        background-color: #fff;
        &::after {
          content: '';
          border-color: #3a84ff;
        }
      }
    }
    .half-indeterminate {
      :deep(.bk-checkbox-input) {
        background-color: #fff;
        &::after {
          content: '';
          background-color: #3a84ff;
        }
      }
    }
    &-icon {
      margin-left: 5px;
      font-size: 20px;
      cursor: pointer;
      color: #63656e;
      &.disabled {
        display: none;
        // color: #c4c6cc;
      }
      &.icon-angle-up {
        transform: rotate(180deg);
      }
    }
  }
  .dropdown-ul {
    margin: -12px;
    font-size: 12px;
    .dropdown-li {
      padding: 0 16px;
      min-width: 68px;
      font-size: 12px;
      text-align: center;
      line-height: 32px;
      cursor: pointer;
      &:hover {
        background: #e5efff;
        color: #3a84ff;
      }
      &.active {
        background: #e5efff;
        color: #3a84ff;
      }
    }
  }
</style>
