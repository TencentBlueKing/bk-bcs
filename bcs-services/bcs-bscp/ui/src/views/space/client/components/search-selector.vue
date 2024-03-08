<template>
  <div>
    <bk-popover
      trigger="click"
      ext-cls="search-selector"
      :is-show="isShow"
      :arrow="false"
      placement="bottom"
      theme="light">
      <div class="search-wrap">
        <div class="search-condition-list">
          <bk-tag
            v-for="(condition, index) in searchConditionList"
            :key="index"
            closable
            @close="handleConditionClose(index)">
            {{ condition.content }}
          </bk-tag>
        </div>
        <div class="search-container-input">
          <bk-input
            v-model="searchStr"
            ref="inputRef"
            class="input"
            :placeholder="inputPlacehoder"
            @blur="handleConfirmConditionItem"
            @enter="handleConfirmConditionItem" />
        </div>
      </div>
      <template #content>
        <div class="menu-wrap">
          <div class="search-condition">
            <div class="title">查询条件</div>
            <div v-if="!showChildSelector">
              <div
                v-for="item in CLIENT_SEARCH_DATA"
                :key="item.value"
                class="search-item"
                @click="handleSelectParent(item)">
                {{ item.name }}
              </div>
            </div>
            <div v-else>
              <div
                v-for="item in childSelectorData"
                :key="item.value"
                class="search-item"
                @click="handleSelectChild(item)">
                {{ item.name }}
              </div>
            </div>
          </div>
          <div class="resent-search">
            <div class="title">最近查询</div>
          </div>
        </div>
      </template>
    </bk-popover>
  </div>
</template>

<script lang="ts" setup>
  import { nextTick, ref, computed } from 'vue';
  import { CLIENT_SEARCH_DATA } from '../../../../constants/client';
  import { ISelectorItem, ISearchCondition } from '../../../../../types/client.ts';

  const isShow = ref(false);
  const searchConditionList = ref<ISearchCondition[]>([]);
  const showChildSelector = ref(false);
  const childSelectorData = ref<ISelectorItem[]>();
  const searchStr = ref('');
  const inputRef = ref();
  const parentSelecte = ref<ISelectorItem>();

  const inputPlacehoder = computed(() => {
    return searchConditionList.value.length
      ? ' '
      : 'UID/IP/标签/当前配置版本/目标配置版本/最近一次拉取配置状态/附加信息/在线状态/客户端组件版本';
  });

  // 选择父选择器
  const handleSelectParent = (parentSelectorItem: ISelectorItem) => {
    parentSelecte.value = parentSelectorItem;
    // 如果有子选择项就展示 没有就用户手动输入
    if (parentSelectorItem?.children) {
      childSelectorData.value = parentSelectorItem.children;
      showChildSelector.value = true;
    } else {
      nextTick(() => inputRef.value.focus());
    }
    searchStr.value = `${parentSelectorItem?.name}:`;
  };

  // 选择子选择器
  const handleSelectChild = (childrenSelectoreItem: ISelectorItem) => {
    showChildSelector.value = false;
    searchConditionList.value.push({
      key: parentSelecte.value!.value,
      value: childrenSelectoreItem.value,
      content: `${parentSelecte.value?.name} : ${childrenSelectoreItem.name}`,
    });
    searchStr.value = '';
  };

  // 手动输入确认搜索项
  const handleConfirmConditionItem = () => {
    const conditionValue = searchStr.value.split(':', 2);
    if (!conditionValue[1]) return;
    searchConditionList.value.push({
      key: parentSelecte.value!.value,
      value: conditionValue[1],
      content: `${parentSelecte.value?.name} : ${conditionValue[1]}`,
    });
    searchStr.value = '';
  };

  // 删除查询条件
  const handleConditionClose = (index: number) => {
    searchConditionList.value.splice(index, 1);
  };
</script>

<style scoped lang="scss">
  .search-wrap {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    padding-left: 8px;
    width: 670px;
    min-height: 32px;
    background: #fff;
    border: 1px solid #c4c6cc;
    .search-container-input {
      min-width: 40px;
      .input {
        border: none;
        height: 100%;
        outline: none;
        box-shadow: none;
      }
    }
    .search-condition-list {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
    }
  }
  .menu-wrap {
    display: flex;
    justify-content: space-between;
    width: calc(670px - 16px);
    .title {
      width: 319px;
      height: 24px;
      background: #eaebf0;
      border-radius: 2px;
      padding-left: 8px;
      color: #313238;
      line-height: 24px;
      margin-bottom: 8px;
    }
    .search-condition {
      .search-item {
        height: 32px;
        padding-left: 12px;
        line-height: 32px;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>

<style lang="scss">
  .bk-popover.bk-pop2-content.search-selector {
    padding: 8px;
  }
</style>
../../../../../types/appp
