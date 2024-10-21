<template>
  <section class="search-option">
    <bk-checkbox v-model="publishVersionConfig" @change="changePublishStatus"> {{ $t('仅看上线操作') }} </bk-checkbox>
    <bk-checkbox v-model="failure" @change="changeFailedStatus"> {{ $t('仅看失败操作') }} </bk-checkbox>
    <div class="search-input__wrap">
      <bk-search-select
        v-model="searchValue"
        :placeholder="$t('资源类型/操作行为/资源实例/状态/操作人/操作途径')"
        :data="searchData"
        unique-select
        :max-height="32"
        @update:model-value="change" />
    </div>
  </section>
</template>

<script setup lang="ts">
  import { computed, onBeforeMount, ref, Ref } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { debounce } from 'lodash';
  import { useI18n } from 'vue-i18n';
  import { RECORD_RES_TYPE, ACTION, STATUS, FILTER_KEY, SEARCH_ID } from '../../../../constants/record';

  interface ISearchValueItem {
    id: string;
    name: string;
    values: { id: string; name: string }[];
  }

  const emits = defineEmits(['sendSearchData']);

  const { t } = useI18n();
  const router = useRouter();
  const route = useRoute();

  const publishVersionConfig = ref(false);
  const failure = ref(false);
  const searchValue = ref<ISearchValueItem[]>([]);

  const searchData = [
    // {
    //   name: '所属服务',
    //   id: SEARCH_ID.service,
    //   multiple: false,
    //   async: false,
    // },
    {
      name: t('资源类型'),
      id: SEARCH_ID.resource_type,
      multiple: false,
      children: Object.entries(RECORD_RES_TYPE).map(([key, value]) => ({
        name: value,
        id: key,
      })),
      async: false,
    },
    {
      name: t('操作行为'),
      id: SEARCH_ID.action,
      multiple: false,
      children: Object.entries(ACTION).map(([key, value]) => ({
        name: value,
        id: key,
      })),
      async: false,
    },
    {
      name: t('资源实例'),
      id: SEARCH_ID.res_instance,
      multiple: false,
      async: false,
    },
    {
      name: t('状态'),
      id: SEARCH_ID.status,
      multiple: false,
      children: Object.entries(STATUS).map(([key, value]) => ({
        name: value,
        id: key,
      })),
      async: false,
    },
    {
      name: t('操作人'),
      id: SEARCH_ID.operator,
      multiple: false,
      async: false,
    },
    {
      name: t('操作途径'),
      id: SEARCH_ID.operate_way,
      multiple: false,
      async: false,
    },
  ];

  const routeSearchValue = computed(() => {
    return searchValue.value.reduce<{ [key: string]: string }>((acc, item) => {
      const key = item.id; // 获取 id 作为键
      const value = item.values.map((v) => v.id).join(';'); // 获取 values 的 id 并用分号连接
      acc[key] = value; // 将结果放入累加器
      return acc;
    }, {});
  });

  onBeforeMount(() => {
    // 获取地址栏参数
    formatUrlParams();
  });

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
    setUrlParams();
    sendSearchData();
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
      setUrlParams();
      sendSearchData();
    },
    300,
    { leading: true },
  );

  // 仅看上线操作
  const changePublishStatus = (status: boolean) => {
    changeStatus(SEARCH_ID.action, t('操作行为'), [{ id: 'PublishVersionConfig', name: t('上线版本配置') }], status);
  };
  // 仅看失败操作
  const changeFailedStatus = (status: boolean) => {
    changeStatus(SEARCH_ID.status, t('状态'), [{ id: 'Failure', name: t('失败') }], status);
  };

  // 设置地址栏参数 http://dev.bscp.sit.bktencent.com:5174/space/2/records/all?action=PublishVersionConfig&status=Failure&service=abc
  const setUrlParams = () => {
    const searchIdArr = Object.keys(SEARCH_ID);
    // 非当前组件搜索参数
    const otherParmas = Object.keys(route.query)
      .filter((key) => !searchIdArr.includes(key))
      .reduce((obj: Record<string, any>, key) => {
        obj[key] = route.query[key];
        return obj;
      }, {});
    router.replace({
      query: {
        ...otherParmas,
        ...routeSearchValue.value,
      },
    });
  };
  // 获取地址栏参数并放入搜搜选项
  const formatUrlParams = () => {
    const searchId = Object.keys(SEARCH_ID); // 搜索id名
    Object.keys(route.query).forEach((routeKey) => {
      // 遍历路由参数的id
      if (searchId.includes(routeKey)) {
        // 过滤当前组件选项的id
        const index = searchData.findIndex((item) => item.id === routeKey);
        const name = searchData[index].name;
        let values = null;
        // 设置子元素
        values = String(route.query[routeKey])
          .split(';')
          .map((valItem) => {
            if (searchData[index].children) {
              const getObj = searchData[index].children?.find((item) => item.id === route.query[routeKey]);
              return {
                id: valItem,
                name: getObj!.name, // searchData正确配置肯定会有name
              };
            }
            return {
              id: valItem,
              name: valItem,
            };
          });
        searchValue.value.push({
          id: routeKey,
          name,
          values,
        });
      }
    });
    // 关联复选框选中
    change(searchValue.value);
  };

  // 发送数据
  const sendSearchData = () => {
    emits('sendSearchData', routeSearchValue.value);
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
