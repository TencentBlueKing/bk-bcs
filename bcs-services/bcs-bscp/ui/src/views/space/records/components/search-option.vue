<template>
  <section class="search-option">
    <bk-checkbox v-model="publishVersionConfig" @change="changePublishStatus"> 仅看上线操作 </bk-checkbox>
    <bk-checkbox v-model="failure" @change="changeFailedStatus"> 仅看失败操作 </bk-checkbox>
    <!-- <bk-checkbox-group v-model="checkboxGroupValue" @change="changeOption">
      <bk-checkbox label="publishAction"> 仅看上线操作 </bk-checkbox>
      <bk-checkbox label="failedStatus"> 仅看失败操作 </bk-checkbox>
    </bk-checkbox-group> -->
    <div class="search-input__wrap">
      <bk-search-select
        v-model="searchValue"
        :placeholder="$t('所属服务/资源类型/操作行为/资源实例/状态/操作人/操作途径')"
        :data="searchData"
        unique-select
        :max-height="32"
        @update:model-value="change" />
    </div>
  </section>
</template>

<script setup lang="ts">
  import { computed, ref, Ref } from 'vue';
  import { debounce } from 'lodash';
  import { RECORD_RES_TYPE, ACTION, STATUS, FILTER_KEY } from '../../../../constants/record';

  enum StatusId {
    resType,
    action,
    status,
  }

  interface ISearchValueItem {
    id: string;
    name: string;
    values: { id: string; name: string }[];
  }

  const publishVersionConfig = ref(false);
  const failure = ref(false);
  const searchValue = ref<ISearchValueItem[]>([]);

  // 资源类型
  const resType = computed(() => {
    return Object.entries(RECORD_RES_TYPE).map(([key, value]) => ({
      name: value,
      id: key,
    }));
  });
  // 操作行为
  const action = computed(() => {
    return Object.entries(ACTION).map(([key, value]) => ({
      name: value,
      id: key,
    }));
  });
  // 状态
  const status = computed(() => {
    return Object.entries(STATUS).map(([key, value]) => ({
      name: value,
      id: key,
    }));
  });
  const searchData = [
    {
      name: '资源类型',
      id: StatusId.resType,
      multiple: true,
      children: [...resType.value],
      async: false,
    },
    {
      name: '操作行为',
      id: StatusId.action,
      multiple: true,
      children: [...action.value],
      async: false,
    },
    {
      name: '状态',
      id: StatusId.status,
      multiple: true,
      children: [...status.value],
      async: false,
    },
  ];

  // 搜索框值变化时 两个“仅看”选项联动
  const change = (data: ISearchValueItem[]) => {
    const optionIdArr = data.map((item) => item.values.map((i) => i.id));
    const statusMap: { [key: string]: Ref<boolean> } = {
      [FILTER_KEY.PublishVersionConfig]: publishVersionConfig,
      [FILTER_KEY.Failure]: failure,
    };
    Object.keys(statusMap).forEach((id) => {
      statusMap[id].value = !!optionIdArr.length && optionIdArr.some((item) => item.every((itemId) => itemId === id));
    });
  };

  const changeStatus = debounce(
    (id, name, values, status) => {
      const actionObj = { id, name, values };
      if (status) {
        const index = searchValue.value.findIndex((item) => item.id === id);
        if (index > -1) {
          // 去除已有数据
          searchValue.value.splice(index, 1);
        }
        searchValue.value.push(actionObj); // 添加
      } else {
        searchValue.value = searchValue.value.filter((option) => option.id !== id); // 删除
      }
      console.log(searchValue.value, 'searchValue.value');
    },
    300,
    { leading: true },
  );

  // 仅看上线操作
  const changePublishStatus = (status: boolean) => {
    changeStatus(StatusId.action, '操作行为', [{ id: 'PublishVersionConfig', name: '上线版本配置' }], status);
  };
  // 仅看失败操作
  const changeFailedStatus = (status: boolean) => {
    changeStatus(StatusId.status, '状态', [{ id: 'Failure', name: '失败' }], status);
  };
</script>

<style lang="scss" scoped>
  .search-option {
    margin-left: auto;
    display: flex;
    align-items: center;
    .search-input__wrap {
      margin-left: 16px;
      width: 400px;
    }
  }
</style>
