<template>
  <bk-sideslider
    :title="t('从已有配置文件添加')"
    :width="640"
    :is-show="isShow"
    :before-close="handleBeforeClose"
    @closed="close">
    <div class="slider-content-container">
      <div class="package-configs-pick">
        <div class="search-wrapper">
          <SearchInput v-model="searchStr" :placeholder="t('配置文件名称/路径/描述')" @search="handleSearch" />
        </div>
        <div class="package-tables">
          <PackageTable
            v-if="packageGroupsOnShow.length"
            v-for="pkg in packageGroupsOnShow"
            v-model:selected-configs="selectedConfigs"
            :key="pkg.id"
            :pkg="pkg"
            :open="openedPkgTable === pkg.id"
            :config-list="pkg.configs"
            @change="isFormChange = true"
            @toggle-open="handleToggleOpenTable" />
          <TableEmpty v-else :is-search-empty="isSearchEmpty" @clear="clearSearchStr"></TableEmpty>
        </div>
      </div>
      <div class="selected-panel">
        <h5 class="title-text">
          {{ t('已选') }} <span class="num">{{ selectedConfigs.length }}</span> {{ t('个配置文件') }}
        </h5>
        <div class="selected-list">
          <div v-for="config in selectedConfigs" class="config-item" :key="config.id">
            <div class="name" :title="config.name">{{ config.name }}</div>
            <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(config.id)" />
          </div>
          <p v-if="selectedConfigs.length === 0" class="empty-tips">{{ t('请先从左侧选择配置文件') }}</p>
        </div>
      </div>
    </div>
    <div class="action-btns">
      <bk-button
        theme="primary"
        :loading="pending"
        :disabled="loading || selectedConfigs.length === 0"
        @click="handleAddConfigs">
        {{ t('添加') }}
      </bk-button>
      <bk-button @click="close">{{ t('取消') }}</bk-button>
    </div>
  </bk-sideslider>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../../store/template';
  import { ITemplateConfigItem } from '../../../../../../../../../types/template';
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation';
  import { addTemplateToPackage } from '../../../../../../../../api/template';
  import PackageTable from './package-table.vue';
  import SearchInput from '../../../../../../../../components/search-input.vue';
  import TableEmpty from '../../../../../../../../components/table/table-empty.vue';
  import { cloneDeep } from 'lodash';

  interface IPackageTableGroup {
    id: number | string;
    name: string;
    configs: ITemplateConfigItem[];
  }

  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    groups: IPackageTableGroup[];
  }>();

  const emits = defineEmits(['update:show', 'added']);

  const isShow = ref(false);
  const isFormChange = ref(false);
  const loading = ref(false);
  const packageGroups = ref<IPackageTableGroup[]>([]);
  const packageGroupsOnShow = ref<IPackageTableGroup[]>([]); // 实际展示的数据，处理搜索的场景
  const pending = ref(false);
  const searchStr = ref('');
  const openedPkgTable = ref<number | string>('');
  const selectedConfigs = ref<{ id: number; name: string }[]>([]);
  const isSearchEmpty = ref(false);

  watch(
    () => props.show,
    (val) => {
      isShow.value = val;
      if (val) {
        openedPkgTable.value = '';
        selectedConfigs.value = [];
        isFormChange.value = false;
      }
    },
  );

  watch(
    () => props.groups,
    () => {
      packageGroups.value = cloneDeep(props.groups);
      packageGroupsOnShow.value = cloneDeep(props.groups);
    },
    { immediate: true },
  );

  const handleSearch = () => {
    isSearchEmpty.value = searchStr.value === '';
    if (searchStr.value) {
      const list: IPackageTableGroup[] = [];
      packageGroups.value.forEach((pkg) => {
        const matchedConfigs = pkg.configs.filter((config) => {
          const { name, path, memo } = config.spec;
          const lowerSearchStr = searchStr.value.toLocaleLowerCase();
          return (
            name.toLocaleLowerCase().includes(lowerSearchStr) ||
            path.toLocaleLowerCase().includes(lowerSearchStr) ||
            memo.toLocaleLowerCase().includes(lowerSearchStr)
          );
        });
        if (matchedConfigs.length > 0) {
          const { id, name } = pkg;
          list.push({ id, name, configs: matchedConfigs });
        }
      });
      packageGroupsOnShow.value = list;
    } else {
      packageGroupsOnShow.value = packageGroups.value.slice();
    }
  };

  const handleToggleOpenTable = (id: string | number) => {
    console.log('1111');
    openedPkgTable.value = openedPkgTable.value === id ? '' : id;
  };

  const handleDeleteConfig = (id: number) => {
    const index = selectedConfigs.value.findIndex((item) => item.id === id);
    if (index > -1) {
      selectedConfigs.value.splice(index, 1);
    }
  };

  const handleAddConfigs = async () => {
    try {
      pending.value = true;
      const configIds = selectedConfigs.value.map((item) => item.id);
      await addTemplateToPackage(spaceId.value, currentTemplateSpace.value, configIds, [currentPkg.value as number]);
      emits('added');
      close();
      Message({
        theme: 'success',
        message: t('添加配置文件成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation();
      return result;
    }
    return true;
  };

  const close = () => {
    emits('update:show', false);
  };

  const clearSearchStr = () => {
    searchStr.value = '';
    handleSearch();
  };
</script>
<style lang="scss" scoped>
  .slider-content-container {
    display: flex;
    align-items: flex-start;
    height: calc(100vh - 101px);
    overflow: auto;
  }
  .search-wrapper {
    padding: 0 16px 0 24px;
    .search-input-icon {
      padding-right: 10px;
      color: #979ba5;
      background: #ffffff;
      font-size: 16px;
    }
  }
  .package-configs-pick {
    padding: 20px 0;
    width: 440px;
    height: 100%;
    .package-tables {
      padding: 16px 16px 0 24px;
      height: calc(100% - 32px);
      overflow: auto;
      .package-config-table:not(:last-of-type) {
        margin-bottom: 16px;
      }
    }
  }
  .selected-panel {
    padding: 20px 24px 20px 16px;
    width: 200px;
    height: 100%;
    background: #f5f7fa;
    .title-text {
      margin: 0;
      line-height: 16px;
      font-size: 12px;
      font-weight: normal;
      color: #63656e;
      .num {
        color: #3a84ff;
        font-weight: 700;
      }
    }
    .selected-list {
      padding-top: 16px;
      height: calc(100% - 16px);
      overflow: auto;
      .config-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 9px 0 12px;
        height: 32px;
        font-size: 12px;
        color: #63656e;
        background: #ffffff;
        border-radius: 2px;
        &:not(:last-of-type) {
          margin-bottom: 4px;
        }
        .name {
          text-overflow: ellipsis;
          overflow: hidden;
          white-space: nowrap;
        }
        .delete-icon {
          margin-left: 4px;
          font-size: 12px;
          cursor: pointer;
          &:hover {
            color: #3a84ff;
          }
        }
      }
      .empty-tips {
        margin: 56px 0 0;
        font-size: 12px;
        color: #979ba5;
        text-align: center;
      }
    }
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
