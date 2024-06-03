<template>
  <section class="section">
    <bk-popover
      trigger="manual"
      ext-cls="search-selector"
      :is-show="isShowPopover"
      :arrow="false"
      placement="bottom-start"
      theme="light"
      :offset="{ alignmentAxis: menuOffset, mainAxis: 6 }"
      @after-show="handleGetSearchList('recent')">
      <div
        class="search-wrap"
        :data-placeholder="inputPlacehoder"
        v-bk-tooltips="{ content: inputPlacehoder, disabled: locale === 'zh-cn' || !inputPlacehoder }"
        @click="handleClickSearch">
        <bk-date-picker
          ref="datePickerRef"
          :model-value="dateTime"
          type="datetimerange"
          ext-popover-cls="selector-date-picker"
          append-to-body
          disable-date
          @change="handleDateChange"
          @open-change="handleDatePickerOpenChange"
          @pick-success="handleConfirmSelectTime">
          <template #trigger>
            <span></span>
          </template>
        </bk-date-picker>
        <div class="search-condition-list">
          <div
            v-for="(condition, index) in searchConditionList"
            :key="condition.key"
            style="margin-right: 6px"
            class="search-condition-item">
            <bk-tag
              v-if="!condition.isEdit"
              closable
              @close="handleConditionClose(index)"
              @click="handleConditionClick($event, condition)">
              {{ condition.content }}
            </bk-tag>
            <input
              v-else
              v-model="editSearchStr"
              ref="editInputRef"
              class="input"
              placeholder=" "
              @blur="handleConditionEdit(condition)"
              @keydown="handleEnterConditionEdit($event, condition)"
              @compositionstart="isComposing = true"
              @compositionend="isComposing = false" />
          </div>
          <div v-if="isShowSearchInput" class="search-container-input" ref="inputWrapRef">
            <input
              v-model="searchStr"
              ref="inputRef"
              class="input"
              placeholder=" "
              @focus="inputFocus = true"
              @blur="handleConfirmConditionItem"
              @keydown="handleEnterAddConditionItem"
              @compositionstart="isComposing = true"
              @compositionend="isComposing = false" />
          </div>
        </div>
        <div
          v-if="searchConditionList.length && isClientSearch"
          :class="['set-used', { light: isCommonlyUsedBtnLight }]"
          v-bk-tooltips="{
            content: highlightCommonlySearchName ? `${t('已收藏为')}: ${highlightCommonlySearchName}` : t('设为常用'),
          }"
          @click.stop="handleOpenSetCommonlyDialg(true)">
          <span class="bk-bscp-icon icon-star-fill"></span>
        </div>
      </div>
      <template #content>
        <div v-if="!showChildSelector" v-click-outside="() => (isShowPopover = false)" class="menu-wrap">
          <div class="search-condition">
            <div class="title">{{ t('查询条件') }}</div>
            <div v-for="item in selectorData" :key="item.value" class="search-item" @click="handleSelectParent(item)">
              {{ item.name }}
            </div>
          </div>
          <div class="resent-search">
            <div class="title">{{ t('最近查询') }}</div>
            <bk-loading :loading="resentSearchListLoading">
              <div
                v-for="item in recentSearchList"
                :key="item.id"
                class="search-item"
                @click="handleSelectRecentSearch(item)">
                <bk-overflow-title type="tips">{{ item.spec.search_name }}</bk-overflow-title>
              </div>
            </bk-loading>
          </div>
        </div>
        <div v-else class="children-menu-wrap" v-click-outside="handleChildSelectorClickOutside">
          <div v-for="item in childSelectorData" :key="item.value" class="search-item" @click="handleSelectChild(item)">
            {{ item.name }}
          </div>
        </div>
      </template>
    </bk-popover>
    <div v-if="isClientSearch" class="commonly-wrap">
      <template v-for="(item, index) in commonlySearchList" :key="item.id">
        <CommonlyUsedTag
          v-if="index < 5"
          :commonly-search-item="item"
          @update="handleOpenSetCommonlyDialg(false, item)"
          @click="searchConditionList = cloneDeep(item.search_condition)"
          @delete="handleOpenDeleteCommonlyDialog(item)" />
      </template>
      <bk-popover
        trigger="manual"
        ext-cls="all-commonly-search-popover"
        placement="bottom-start"
        theme="light"
        :is-show="isShowAllCommonSearchPopover"
        :arrow="false">
        <bk-button theme="primary" text @click="isShowAllCommonSearchPopover = !isShowAllCommonSearchPopover">
          {{ t('全部常用查询') }}
        </bk-button>
        <template #content>
          <div
            v-for="item in commonlySearchList"
            :key="item.id"
            class="search-item"
            v-click-outside="() => (isShowAllCommonSearchPopover = false)"
            @click="handleSelectCommonSearch(item)">
            <div class="name">
              <bk-overflow-title>{{ item.spec.search_name }}</bk-overflow-title>
            </div>
            <div class="action-icon" v-if="item.spec.creator !== 'system'">
              <span class="bk-bscp-icon icon-edit-line edit" @click.stop="handleOpenSetCommonlyDialg(false, item)" />
              <span class="bk-bscp-icon icon-close-line close" @click.stop="handleOpenDeleteCommonlyDialog(item)" />
            </div>
          </div>
        </template>
      </bk-popover>
    </div>
    <SetCommonlyDialog
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :is-show="isShowSetCommonlyDialog"
      :is-create="isCreateCommonlyUsed"
      :name="selectedCommomlyItem?.spec.search_name"
      @create="handleConfirmCreateCommonlyUsed"
      @update="handleConfirmUpdateCommonlyUsed"
      @close="isShowSetCommonlyDialog = false" />
    <bk-dialog
      :is-show="isShowDeleteCommonlyDialog"
      :ext-cls="'delete-commonly-dialog'"
      :width="400"
      @closed="isShowDeleteCommonlyDialog = false">
      <div class="head">{{ t('确认删除该常用查询?') }}</div>
      <div class="body">
        <span class="label">{{ t('名称') }} : </span>
        <span class="name">{{ selectedDeleteCommonlyItem?.spec.search_name }}</span>
      </div>
      <div class="footer">
        <div class="btns">
          <bk-button theme="danger" @click="handleConfirmDeleteCommonlyUsed">{{ t('删除') }}</bk-button>
          <bk-button @click="isShowDeleteCommonlyDialog = false">{{ t('取消') }}</bk-button>
        </div>
      </div>
    </bk-dialog>
  </section>
</template>

<script lang="ts" setup>
  import { nextTick, ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
  import { storeToRefs } from 'pinia';
  import {
    CLIENT_SEARCH_DATA,
    CLIENT_STATISTICS_SEARCH_DATA,
    CLIENT_STATUS_MAP,
    CLIENT_ERROR_CATEGORY_MAP,
    CLIENT_COMPONENT_TYPES_MAP,
  } from '../../../../constants/client';
  import { ISelectorItem, ISearchCondition, ICommonlyUsedItem, IClinetCommonQuery } from '../../../../../types/client';
  import {
    getClientSearchRecord,
    createClientSearchRecord,
    updateClientSearchRecord,
    deleteClientSearchRecord,
  } from '../../../../api/client';
  import { getTimeRange, datetimeFormat } from '../../../../utils';
  import useClientStore from '../../../../store/client';
  import SetCommonlyDialog from './set-commonly-dialog.vue';
  import CommonlyUsedTag from './commonly-used-tag.vue';
  import { Message } from 'bkui-vue';
  import { cloneDeep } from 'lodash';
  import { useRoute } from 'vue-router';
  import { useI18n } from 'vue-i18n';

  const { t, locale } = useI18n();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const route = useRoute();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const isShowPopover = ref(false);
  const searchConditionList = ref<ISearchCondition[]>([]);
  const showChildSelector = ref(false);
  const childSelectorData = ref<ISelectorItem[]>();
  const searchStr = ref('');
  const inputRef = ref();
  const parentSelecte = ref<ISelectorItem>();
  const recentSearchList = ref<ICommonlyUsedItem[]>([]);
  const resentSearchListLoading = ref(false);
  const commonlySearchList = ref<ICommonlyUsedItem[]>([]);
  const inputFocus = ref(false);
  const isShowSetCommonlyDialog = ref(false);
  const isCreateCommonlyUsed = ref(true);
  const selectedCommomlyItem = ref<ICommonlyUsedItem>();
  const isShowDeleteCommonlyDialog = ref(false);
  const selectedDeleteCommonlyItem = ref<ICommonlyUsedItem>();
  const isShowAllCommonSearchPopover = ref(false);
  const editSearchStr = ref('');
  const editInputRef = ref();
  const editConditionItem = ref<ISearchCondition>();
  const menuOffset = ref(0);
  const inputWrapRef = ref();
  const dateTime = ref(getTimeRange(1));
  const datePickerRef = ref();
  const highlightCommonlySearchName = ref('');
  const isComposing = ref(false); // 是否使用输入法
  const isShowSearchInput = ref(false);

  const inputPlacehoder = computed(() => {
    if (searchConditionList.value.length || searchStr.value || inputFocus.value) return '';
    if (isClientSearch.value) {
      return t(
        'UID/IP/标签/当前配置版本/最近一次拉取配置状态/在线状态/客户端组件类型/客户端组件版本/配置拉取时间范围/错误类别',
      );
    }
    return t('标签/当前配置版本/最近一次拉取配置状态/在线状态/客户端组件类型/客户端组件版本');
  });

  const isClientSearch = computed(() => route.name === 'client-search');

  const selectorData = computed(() => (isClientSearch.value ? CLIENT_SEARCH_DATA : CLIENT_STATISTICS_SEARCH_DATA));

  const isCommonlyUsedBtnLight = computed(() => {
    const item = commonlySearchList.value.find((commonlySearchItem) => {
      if (commonlySearchItem.search_condition.length !== searchConditionList.value.length) return false;
      return commonlySearchItem.search_condition.every((commonlySearchConditionList) => {
        const { key, value } = commonlySearchConditionList;
        return searchConditionList.value.findIndex((item) => item.key === key && item.value === value) > -1;
      });
    });
    if (item) {
      highlightCommonlySearchName.value = item.spec.search_name;
      return true;
    }
    highlightCommonlySearchName.value = '';
    return false;
  });

  watch(
    () => searchConditionList.value,
    () => {
      // 搜索框和查询条件都为空时不需要转换查询参数
      if (searchConditionList.value.length === 0 && Object.keys(searchQuery.value.search!).length === 0) return;
      handleSearchConditionChangeQuery();
    },
    { deep: true },
  );

  watch(
    () => props.appId,
    () => {
      handleGetSearchList('common');
      searchConditionList.value = [];
    },
  );

  watch(
    () => searchQuery.value.search,
    (val) => {
      if (Object.keys(val!).length === 0) {
        searchConditionList.value = [];
      } else {
        handleAddRecentSearch();
      }
    },
  );

  watch(
    () => isShowPopover.value,
    (val) => {
      if (val && !searchStr.value && !editSearchStr.value) {
        showChildSelector.value = false;
        parentSelecte.value = undefined;
      }
    },
  );

  onMounted(() => {
    handleGetSearchList('common');
    const entries = Object.entries(route.query);
    if (entries.length === 0) return;
    const { name, value, children } = CLIENT_SEARCH_DATA.find((item) => item.value === entries[0][0])!;

    if (value === 'pull_time') {
      searchConditionList.value.push({
        content: `${name} : ${entries[0][1]} 00:00:00 - ${entries[0][1]} 23:59:59`,
        value: `${entries[0][1]} 00:00:00 - ${entries[0][1]} 23:59:59`,
        key: value,
        isEdit: false,
      });
    } else if (value === 'label') {
      const labels = JSON.parse(entries[0][1] as string);
      Object.keys(labels).forEach((key) => {
        searchConditionList.value.push({
          content: `标签: ${key}=${labels[key]}`,
          value: `${key}=${labels[key]}`,
          key: value,
          isEdit: false,
        });
      });
    } else {
      let content = `${entries[0][1]}`;
      if (children) {
        content = children.find((item) => item.value === entries[0][1])!.name;
      }
      searchConditionList.value.push({
        content: `${name} : ${content}`,
        value: entries[0][1] as string,
        key: value,
        isEdit: false,
      });
    }
  });

  onBeforeUnmount(() => {
    clientStore.$patch((state) => {
      state.searchQuery.search = {};
    });
  });

  // 选择父选择器
  const handleSelectParent = (parentSelectorItem: ISelectorItem) => {
    isShowSearchInput.value = true;
    parentSelecte.value = parentSelectorItem;
    // 如果有子选择项就展示 没有就用户手动输入
    if (parentSelectorItem.value === 'pull_time') {
      isShowPopover.value = false;
      nextTick(() => datePickerRef.value.handleFocus());
    } else if (parentSelectorItem?.children) {
      childSelectorData.value = parentSelectorItem.children;
      nextTick(() => (menuOffset.value = inputWrapRef.value.offsetLeft));
      showChildSelector.value = true;
    } else {
      isShowPopover.value = false;
      nextTick(() => inputRef.value.focus());
    }
    searchStr.value = `${parentSelectorItem?.name} : `;
  };

  // 选择子选择器
  const handleSelectChild = (childrenSelectoreItem: ISelectorItem) => {
    showChildSelector.value = false;
    isShowPopover.value = false;
    // 重复的查询项去重
    const index = searchConditionList.value.findIndex(
      (item) => item.key === parentSelecte.value?.value && item.key !== 'label',
    );
    if (index > -1) handleConditionClose(index);
    searchConditionList.value.push({
      key: parentSelecte.value!.value,
      value: childrenSelectoreItem.value,
      content: `${parentSelecte.value?.name} : ${childrenSelectoreItem.name}`,
      isEdit: false,
    });
    searchStr.value = '';
    parentSelecte.value = undefined;
    menuOffset.value = 0;
  };

  const handleConfirmConditionItem = () => {
    const conditionValue = parentSelecte.value ? searchStr.value.split(' : ', 2)[1] : searchStr.value;
    inputFocus.value = false;
    isShowSearchInput.value = false;
    if (!conditionValue) {
      searchStr.value = '';
      return;
    }
    // 添加默认查询条件ip
    if (!parentSelecte.value?.value) {
      parentSelecte.value = selectorData.value.find((item) => {
        return isClientSearch.value ? item.value === 'ip' : item.value === 'current_release_name';
      })!;
    }
    // 重复的查询项去重
    const index = searchConditionList.value.findIndex(
      (item) => item.key === parentSelecte.value?.value && item.key !== 'label',
    );
    if (index > -1) handleConditionClose(index);
    searchConditionList.value.push({
      key: parentSelecte.value!.value,
      value: conditionValue,
      content: `${parentSelecte.value?.name} : ${conditionValue}`,
      isEdit: false,
    });
    searchStr.value = '';
    isShowPopover.value = false;
    inputRef.value.blur();
    parentSelecte.value = undefined;
  };

  const handleEnterAddConditionItem = (e: any) => {
    if (e.keyCode === 13) {
      if (isComposing.value) {
        e.preventDefault();
      } else {
        handleConfirmConditionItem();
      }
    }
  };

  const handleDateChange = (val: string[]) => {
    dateTime.value = val;
  };

  const handleConfirmSelectTime = () => {
    const index = searchConditionList.value.findIndex((item) => item.key === 'pull_time');
    if (index > -1) handleConditionClose(index);
    searchStr.value = '';
    searchConditionList.value.push({
      key: parentSelecte.value!.value,
      value: `${dateTime.value[0]} - ${dateTime.value[1]}`,
      content: `${parentSelecte.value?.name} : ${dateTime.value[0]} - ${dateTime.value[1]}`,
      isEdit: false,
    });
  };

  // 获取最近搜索记录和常用搜索记录
  const handleGetSearchList = async (search_type: string) => {
    if (!props.appId) return;
    try {
      resentSearchListLoading.value = search_type === 'recent';
      const params: IClinetCommonQuery = {
        start: 0,
        limit: 10,
        search_type,
      };
      if (search_type === 'common') {
        params.all = true;
      } else {
        isClientSearch.value ? (params.search_type = 'query') : (params.search_type = 'statistic');
      }
      const res = await getClientSearchRecord(props.bkBizId, props.appId, params);
      const searchList = res.data.details;
      searchList.forEach((item: ICommonlyUsedItem) => handleQueryChangeSearchCondition(item));
      if (search_type === 'recent') {
        recentSearchList.value = searchList;
      } else {
        commonlySearchList.value = searchList;
      }
    } catch (error) {
      console.error(error);
    } finally {
      resentSearchListLoading.value = false;
    }
  };

  // 删除查询条件
  const handleConditionClose = (index: number) => {
    searchConditionList.value.splice(index, 1);
    showChildSelector.value = false;
  };

  // 添加最近查询
  const handleAddRecentSearch = async () => {
    await createClientSearchRecord(props.bkBizId, props.appId, {
      search_type: isClientSearch.value ? 'query' : 'statistic',
      search_condition: searchQuery.value.search!,
    });
  };

  // 设置常用查询
  const handleConfirmCreateCommonlyUsed = async (search_name: string) => {
    try {
      await createClientSearchRecord(props.bkBizId, props.appId, {
        search_condition: searchQuery.value.search!,
        search_type: 'common',
        search_name,
      });
      isShowSetCommonlyDialog.value = false;
      handleGetSearchList('common');
      Message({
        theme: 'success',
        message: t('常用查询添加成功'),
      });
    } catch (error) {
      console.error(error);
    }
  };

  // 更新常用查询
  const handleConfirmUpdateCommonlyUsed = async (search_name: string) => {
    try {
      await updateClientSearchRecord(props.bkBizId, props.appId, selectedCommomlyItem.value!.id, {
        search_condition: selectedCommomlyItem.value!.spec.search_condition,
        search_type: 'common',
        search_name,
      });
      isShowSetCommonlyDialog.value = false;
      handleGetSearchList('common');
      Message({
        theme: 'success',
        message: t('常用查询修改成功'),
      });
    } catch (error) {
      console.error(error);
    }
  };

  // 删除常用查询
  const handleOpenDeleteCommonlyDialog = (item: ICommonlyUsedItem) => {
    selectedDeleteCommonlyItem.value = item;
    isShowDeleteCommonlyDialog.value = true;
    isShowAllCommonSearchPopover.value = false;
  };

  const handleConfirmDeleteCommonlyUsed = async () => {
    try {
      await deleteClientSearchRecord(props.bkBizId, props.appId, selectedDeleteCommonlyItem.value!.id);
      isShowDeleteCommonlyDialog.value = false;
      handleGetSearchList('common');
      Message({
        theme: 'success',
        message: t('常用查询删除成功'),
      });
    } catch (error) {
      console.error(error);
    }
  };

  const handleOpenSetCommonlyDialg = (isCreate: boolean, item?: ICommonlyUsedItem) => {
    if (isCreate) {
      if (isCommonlyUsedBtnLight.value) return;
      isCreateCommonlyUsed.value = true;
    } else {
      isCreateCommonlyUsed.value = false;
      selectedCommomlyItem.value = item;
    }
    isShowSetCommonlyDialog.value = true;
    isShowAllCommonSearchPopover.value = false;
  };

  // 查询条件转换为查询参数
  const handleSearchConditionChangeQuery = () => {
    const query: { [key: string]: any } = {};
    const label: { [key: string]: any } = {};
    searchConditionList.value.forEach((item) => {
      if (item.key === 'label') {
        const labelValue = item.value.split('=', 2);
        label[labelValue[0]] = labelValue[1] || '';
        query[item.key] = label;
      } else if (item.key === 'online_status' || item.key === 'release_change_status') {
        if (query[item.key]) {
          query[item.key].push(item.value);
        } else {
          query[item.key] = [item.value];
        }
      } else if (item.key === 'pull_time') {
        const startTime = item.value.split(' - ')[0];
        const endTime = item.value.split(' - ')[1];
        query.start_pull_time = new Date(`${startTime.replace(' ', 'T')}+08:00`).toISOString();
        query.end_pull_time = new Date(`${endTime.replace(' ', 'T')}+08:00`).toISOString();
      } else {
        query[item.key] = item.value.trim();
      }
    });
    clientStore.$patch((state) => {
      state.searchQuery.search = query;
    });
  };

  // 查询参数转换为查询条件并获取查询名
  const handleQueryChangeSearchCondition = (item: ICommonlyUsedItem) => {
    const searchList: ISearchCondition[] = [];
    const searchName: string[] = [];
    const query: { [key: string]: any } = item.spec.search_condition;
    Object.keys(query).forEach((key) => {
      if (key === 'label') {
        const labelValue = query[key];
        Object.keys(labelValue).forEach((label) => {
          const value = labelValue[label] || '';
          const content = value ? `${t('标签')} : ${label}=${labelValue[label]}` : `${t('标签')} : ${label}`;
          searchList.push({
            key,
            value: `${label}=${value}`,
            content,
            isEdit: false,
          });
          searchName.push(content);
        });
      } else if (key === 'online_status' || key === 'release_change_status') {
        query[key].forEach((value: string) => {
          const content = `${selectorData.value.find((item) => item.value === key)?.name} : ${
            CLIENT_STATUS_MAP[value as keyof typeof CLIENT_STATUS_MAP]
          }`;
          searchList.push({
            key,
            value,
            content,
            isEdit: false,
          });
          searchName.push(content);
        });
      } else if (key === 'start_pull_time' || key === 'end_pull_time') {
        if (searchList.find((item) => item.key === 'pull_time')) return;
        const content = `${t('配置拉取时间范围')} : ${datetimeFormat(query.start_pull_time)} - ${datetimeFormat(
          query.end_pull_time,
        )}`;
        searchList.push({
          key: 'pull_time',
          value: `${datetimeFormat(query.start_pull_time)} - ${datetimeFormat(query.end_pull_time)}`,
          content,
          isEdit: false,
        });
        searchName.push(content);
      } else if (key === 'failed_reason') {
        const errorItem = CLIENT_ERROR_CATEGORY_MAP.find((item) => item.value === query[key]);
        const content = `${selectorData.value.find((item) => item.value === key)?.name} : ${errorItem?.name}`;
        searchList.push({
          key,
          value: query[key],
          content,
          isEdit: false,
        });
        searchName.push(content);
      } else if (key === 'client_type') {
        const clientType = CLIENT_COMPONENT_TYPES_MAP.find((item) => item.value === query[key]);
        const content = `${selectorData.value.find((item) => item.value === key)?.name} : ${clientType?.name}`;
        searchList.push({
          key,
          value: query[key],
          content,
          isEdit: false,
        });
        searchName.push(content);
      } else {
        const content = `${selectorData.value.find((item) => item.value === key)?.name} : ${query[key]}`;
        searchList.push({
          key,
          value: query[key],
          content,
          isEdit: false,
        });
        searchName.push(content);
      }
    });
    item.search_condition = searchList;
    item.spec.search_name = item.spec.search_name || searchName.join(';');
  };

  const handleSelectRecentSearch = (item: ICommonlyUsedItem) => {
    searchConditionList.value = cloneDeep(item.search_condition);
    isShowPopover.value = false;
  };

  const handleSelectCommonSearch = (item: ICommonlyUsedItem) => {
    searchConditionList.value = cloneDeep(item.search_condition);
    isShowAllCommonSearchPopover.value = false;
  };

  const handleClickSearch = () => {
    isShowPopover.value = !isShowPopover.value;
    isShowSearchInput.value = true;
    nextTick(() => inputRef.value.focus());
  };

  const handleConditionClick = (e: any, condition: ISearchCondition) => {
    e.preventDefault();
    e.stopPropagation();
    editConditionItem.value = condition;
    condition.isEdit = true;
    editSearchStr.value = condition.content;
    parentSelecte.value = selectorData.value.find((item) => item.value === condition.key);
    if (condition.key === 'pull_time') {
      nextTick(() => datePickerRef.value.handleFocus());
      isShowPopover.value = false;
    } else if (parentSelecte.value?.children) {
      childSelectorData.value = parentSelecte.value.children;
      showChildSelector.value = true;
      isShowPopover.value = true;
    } else {
      setTimeout(() => {
        editInputRef.value[0].focus();
      }, 200);
      isShowPopover.value = false;
    }
  };

  const handleConditionEdit = (condition: ISearchCondition) => {
    if (!condition.isEdit) return;
    const conditionValue = editSearchStr.value.split(' : ', 2)[1];
    if (conditionValue) {
      condition.value = conditionValue;
      condition.content = `${parentSelecte.value?.name} : ${conditionValue}`;
    }
    condition.isEdit = false;
    parentSelecte.value = undefined;
  };

  const handleEnterConditionEdit = (e: any, condition: ISearchCondition) => {
    if (e.keyCode === 13) {
      if (isComposing.value) {
        e.preventDefault();
      } else {
        handleConditionEdit(condition);
      }
    }
  };

  const handleDatePickerOpenChange = (open: boolean) => {
    if (open) {
      isShowPopover.value = false;
    } else {
      const condition = searchConditionList.value.find((item) => item.key === 'pull_time');
      if (condition) {
        condition.isEdit = false;
      }
    }
  };

  const handleChildSelectorClickOutside = () => {
    if (editSearchStr.value) {
      // 编辑态 取消编辑
      editConditionItem.value!.isEdit = false;
      editSearchStr.value = '';
    } else if (searchStr.value) {
      // 新增态 状态复原
      searchStr.value = '';
    }
    isShowPopover.value = false;
    showChildSelector.value = false;
  };
</script>

<style scoped lang="scss">
  .section {
    position: relative;
  }
  .search-wrap {
    position: relative;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    padding-left: 8px;
    width: 670px;
    min-height: 32px;
    background: #fff;
    border: 1px solid #c4c6cc;
    padding-right: 32px;
    &::after {
      position: absolute;
      width: calc(100% - 16px);
      content: attr(data-placeholder);
      color: #c4c6cc;
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .bk-date-picker {
      width: 0 !important;
    }
    .search-container-input {
      .input {
        min-width: fit-content;
        border: none;
        height: 100%;
        outline: none;
        box-shadow: none;
        color: #63656e;
      }
    }
    .search-condition-list {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      .search-condition-item {
        .input {
          border: none;
          height: 100%;
          outline: none;
          box-shadow: none;
        }
      }
    }
    .set-used {
      position: absolute;
      right: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: #f0f1f5;
      border-radius: 2px;
      color: #c4c6cc;
      font-size: 14px;
      &.light span {
        color: #ff9c01;
      }
    }
  }
  .menu-wrap {
    display: flex;
    justify-content: space-between;
    width: calc(670px - 16px);
    padding: 8px;
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
    .search-item {
      width: 319px;
      height: 32px;
      padding-left: 12px;
      line-height: 32px;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
  .children-menu-wrap {
    min-width: 200px;
    padding: 4px 0;
    div {
      display: flex;
      align-items: center;
      height: 32px;
      padding: 0 8px;
      color: #63656e;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
  .commonly-wrap {
    height: 26px;
    display: flex;
    align-items: center;
    margin: 6px 0;
  }
  .action-item {
    width: 58px;
    height: 32px;
    line-height: 32px;
    text-align: center;
    color: #63656e;
    cursor: pointer;
    &:hover {
      background: #f5f7fa;
    }
  }
</style>

<style lang="scss">
  .bk-popover.bk-pop2-content.search-selector {
    padding: 0;
  }
  .commonly-search-item-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
  }
  .all-commonly-search-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    max-height: 168px;
    overflow-y: auto;
    .search-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0 8px 0 12px;
      width: 238px;
      height: 32px;
      cursor: pointer;
      color: #63656e;
      &:hover {
        background-color: #f5f7fa;
        .action-icon {
          display: block;
        }
      }
      .name {
        max-width: 120px;
      }
      .action-icon {
        display: none;
        font-size: 16px;
        height: 32px;
        line-height: 32px;
        .bk-bscp-icon:hover {
          color: #3a84ff;
        }
        .edit {
          margin-right: 12px;
        }
      }
    }
  }

  .delete-commonly-dialog .bk-modal-body {
    .bk-modal-header {
      display: none;
    }
    .bk-modal-footer {
      display: none;
    }
    .bk-modal-content {
      padding: 48px 24px 0 24px;
      .head {
        font-size: 20px;
        color: #313238;
        text-align: center;
      }
      .body {
        min-height: 32px;
        line-height: 32px;
        background: #f5f7fa;
        padding-left: 16px;
        margin-top: 16px;
        .label {
          color: #63656e;
        }
        .name {
          color: #313238;
        }
      }
      .footer {
        display: flex;
        justify-content: space-around;
        .btns {
          margin-top: 24px;
          .bk-button {
            width: 88px;
          }
          .bk-button:nth-child(1) {
            margin-right: 8px;
          }
        }
      }
    }
  }
  .selector-date-picker {
    top: 130px !important;
    left: 723px !important;
  }
</style>
