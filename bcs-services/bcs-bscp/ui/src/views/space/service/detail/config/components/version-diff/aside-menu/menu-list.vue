<template>
  <div class="diff-menu-list">
    <div class="title-area">
      <div class="title">{{ props.title }}</div>
      <div class="title-extend">
        <slot name="title"></slot>
      </div>
    </div>
    <div class="list-wrapper">
      <div
        v-for="item in props.list"
        :class="['menu-item', { active: selected === item.id }]"
        :key="item.id"
        @click="handleSelect(item.id)">
        <i v-if="item.type" :class="['status-icon', item.type]"></i>
        <div class="name">{{ item.name }}</div>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';

  const props = defineProps<{
    title: string;
    list: { id: number | string; name: string; type: string }[];
    value?: string | number;
  }>();

  const emits = defineEmits(['selected']);

  const selected = ref<number | string | undefined>(props.value);

  const handleSelect = (id: number | string) => {
    selected.value = id;
    emits('selected', id);
  };

  watch(
    () => props.value,
    (val) => {
      selected.value = val;
    },
    {
      immediate: true,
    },
  );
</script>
<style lang="scss" scoped>
  .diff-menu-list {
    height: 100%;
  }
  .title-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    height: 49px;
    color: #979ba5;
    font-size: 12px;
    border-bottom: 1px solid #dcdee5;
    background-color: #ffffff;
    .title {
      font-size: 14px;
      font-weight: 700;
      color: #63656e;
    }
    .count {
      color: #313238;
    }
  }
  .list-wrapper {
    .menu-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: relative;
      padding: 0 16px 0 32px;
      height: 41px;
      color: #313238;
      background: #ffffff;
      border-bottom: 1px solid #dcdee5;
      cursor: pointer;
      &:hover {
        background: #e1ecff;
        color: #3a84ff;
      }
      &.active {
        background: #e1ecff;
        color: #3a84ff;
      }
      .status-icon {
        position: absolute;
        top: 18px;
        left: 16px;
        width: 4px;
        height: 4px;
        border-radius: 50%;
        &.add {
          background: #3a84ff;
        }
        &.delete {
          background: #ea3536;
        }
        &.modify {
          background: #fe9c00;
        }
      }
      .name {
        width: calc(100% - 24px);
        line-height: 16px;
        font-size: 12px;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
      .arrow-icon {
        position: absolute;
        top: 50%;
        right: 5px;
        transform: translateY(-60%);
        font-size: 12px;
        color: #3a84ff;
      }
    }
  }
</style>
