<template>
  <bk-dialog
    quick-close
    width="1105"
    dialog-type="show"
    :is-show="props.isShow"
    :theme="'primary'"
    :show-mask="true"
    :show-footer="false"
    :scrollable="false"
    @closed="emits('update:isShow', false)"
    ext-cls="version-dialog">
    <div class="log-version">
      <div class="log-version-left">
        <ul class="left-list">
          <li
            v-for="(item, index) in props.logList"
            :class="['left-list-item', { 'item-active': index === active }]"
            :key="index"
            @click="active = index">
            <slot>
              <span class="item-title">{{ item.title }}</span>
              <span class="item-date">{{ item.date }}</span>
              <span v-if="index === 0" class="item-current">{{ t('当前版本') }}</span>
            </slot>
          </li>
        </ul>
      </div>
      <div class="log-version-right">
        <slot name="detail">
          <div class="markdown-theme-style" v-html="logList[active].detail"></div>
        </slot>
      </div>
    </div>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { IVersionLogItem } from '../../types/version-log';

  const { t } = useI18n();
  const props = withDefaults(
    defineProps<{
      logList: IVersionLogItem[];
      isShow: boolean;
    }>(),
    {},
  );
  const emits = defineEmits(['update:isShow']);
  const active = ref(0);
</script>

<style scoped lang="scss">
  .log-version {
    display: flex;
    max-height: calc(100vh - 300px);
    overflow: auto;
    &-left {
      flex: 0 0 180px;
      background-color: #fafbfd;
      border-right: 1px solid #dcdee5;
      padding: 40px 0;
      display: flex;
      font-size: 12px;
      overflow: hidden;
      .left-list {
        border-top: 1px solid #dcdee5;
        border-bottom: 1px solid #dcdee5;
        height: 520px;
        overflow: auto;
        display: flex;
        flex-direction: column;
        width: 100%;
        &-item {
          flex: 0 0 54px;
          display: flex;
          flex-direction: column;
          justify-content: center;
          padding-left: 30px;
          position: relative;
          border-bottom: 1px solid #dcdee5;
          &:hover {
            cursor: pointer;
            background-color: #ffffff;
          }
          .item-title {
            color: #313238;
            font-size: 16px;
          }
          .item-date {
            color: #979ba5;
          }
          .item-current {
            position: absolute;
            right: 20px;
            top: 8px;
            background-color: #699df4;
            border-radius: 2px;
            width: 58px;
            height: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #ffffff;
          }
          &.item-active {
            &::before {
              content: ' ';
              position: absolute;
              top: 0px;
              bottom: 0px;
              left: 0;
              width: 6px;
              background-color: #3a84ff;
            }
            background-color: #ffffff;
          }
        }
      }
    }
    &-right {
      flex: 1;
      padding: 25px 30px 50px 45px;
      .detail-container {
        overflow: auto;
      }
    }
  }
  .bk-modal-wrapper.bk-dialog-wrapper .bk-modal-content {
    padding: 0;
  }
</style>
