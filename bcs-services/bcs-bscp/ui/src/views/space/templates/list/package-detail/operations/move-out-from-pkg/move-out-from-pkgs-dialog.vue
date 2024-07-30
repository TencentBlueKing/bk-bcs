<template>
  <bk-dialog
    ext-cls="move-out-from-pkgs-dialog"
    header-align="center"
    footer-align="center"
    :title="t('确认从配置套餐中移出该配置文件?')"
    :width="600"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @confirm="handleConfirm"
    @closed="close">
    <div style="margin-bottom: 8px">
      {{ t('配置文件') }}: <span style="color: #313238; font-weight: 600">{{ name }}</span>
    </div>
    <div class="service-table">
      <bk-loading style="min-height: 100px" :loading="loading">
        <bk-table
          v-if="!loading"
          :data="citedList"
          :max-height="maxTableHeight"
          :checked="checkedPkgs"
          :is-row-select-enable="isRowSelectEnable"
          show-overflow-tooltip
          @selection-change="handleSelectionChange"
          @select-all="handleSelectAll">
          <bk-table-column v-if="citedList.length > 1" type="selection" min-width="30" width="40" />
          <bk-table-column :label="t('所在模板套餐')">
            <template #default="{ row }">
              <div class="pkg-name">
                <span v-overflow-title class="name-text">{{ row.name }}</span>
                <span v-if="props.currentPkg === row.id" class="tag">{{ t('当前') }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column show-overflow-tooltip :label="t('使用此套餐的服务')" prop="appNames"></bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
    <p v-if="citedList.length === 1 || selectedPkgs.length === citedList.length" class="tips">
      <Warn class="warn-icon" />
      {{ t('移出后配置文件将不存在任一套餐。你仍可在「全部配置文件」或「未指定套餐」分类下找回。') }}
    </p>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" :disabled="selectedPkgs.length === 0" @click="handleConfirm">
          {{ t('确认移出') }}
        </bk-button>
        <bk-button @click="close">{{ t('取消') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { Warn } from 'bkui-vue/lib/icon';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import { IPackagesCitedByApps } from '../../../../../../../../types/template';
  import {
    getPackagesByTemplateIds,
    getUnNamedVersionAppsBoundByPackages,
    moveOutTemplateFromPackage,
  } from '../../../../../../../api/template';

  interface ICitedItem {
    template_id: number;
    id: number;
    name: string;
    appNames: string;
  }

  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, isAcrossChecked } = storeToRefs(useTemplateStore());
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    id: number;
    name: string;
    currentPkg: number | string;
  }>();

  const emits = defineEmits(['update:show', 'movedOut']);

  const selectedPkgs = ref<number[]>([]);
  const citedList = ref<ICitedItem[]>([]);
  const loading = ref(false);
  const pending = ref(false);

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight;
    return windowHeight * 0.6 - 200;
  });

  const checkedPkgs = computed(() => {
    return citedList.value.filter((pkg) => selectedPkgs.value.includes(pkg.id));
  });

  watch(
    () => props.show,
    (val) => {
      if (val) {
        selectedPkgs.value = [];
        getCitedData();
      }
    },
  );

  const getCitedData = async () => {
    loading.value = true;
    const citedPkgsRes = await getPackagesByTemplateIds(spaceId.value, currentTemplateSpace.value, [props.id]);
    if (citedPkgsRes.details.length === 1) {
      const pkgs = citedPkgsRes.details[0].map((item) => item.template_set_id);
      const params = {
        start: 0,
        all: true,
      };
      let list: ICitedItem[] = [];
      const citedAppsRes = await getUnNamedVersionAppsBoundByPackages(
        spaceId.value,
        currentTemplateSpace.value,
        pkgs,
        params,
      );
      citedPkgsRes.details[0].forEach((item) => {
        // console.log(item, 'item');
        const { template_set_id, template_set_name, template_id } = item;
        const appNames: string =
          citedAppsRes.details
            .filter((appItem: IPackagesCitedByApps) => appItem.template_set_id === template_set_id)
            .map((appItem: IPackagesCitedByApps) => appItem.app_name)
            .join(',') || '--';
        list.push({
          template_id,
          id: template_set_id,
          name: template_set_name,
          appNames,
        });
      });
      const index = list.findIndex((item) => item.id === props.currentPkg);
      const currentPkgData = list.splice(index, 1);
      list = currentPkgData.concat(list);
      citedList.value = list;
    }

    if (typeof props.currentPkg === 'number') {
      selectedPkgs.value = [props.currentPkg];
    } else if (props.currentPkg === 'all' && citedList.value.length === 1) {
      selectedPkgs.value = [citedList.value[0].id];
    }

    loading.value = false;
  };

  const handleSelectionChange = ({ checked, row }: { checked: boolean; row: ICitedItem }) => {
    if (checked) {
      if (!selectedPkgs.value.includes(row.id)) {
        selectedPkgs.value.push(row.id);
      }
    } else {
      const index = selectedPkgs.value.findIndex((id) => id === row.id);
      if (index > -1) {
        selectedPkgs.value.splice(index, 1);
      }
    }
  };

  const handleSelectAll = ({ checked }: { checked: boolean }) => {
    if (checked) {
      selectedPkgs.value = citedList.value.map((item) => item.id);
    } else {
      if (typeof props.currentPkg === 'number') {
        selectedPkgs.value = [props.currentPkg];
      } else if (props.currentPkg === 'all' && citedList.value.length === 1) {
        selectedPkgs.value = [citedList.value[0].id];
      }
    }
  };

  const isRowSelectEnable = ({ row, isCheckAll }: any) => {
    return isCheckAll || props.currentPkg !== row.id;
  };
  const handleConfirm = async () => {
    try {
      pending.value = true;
      const appIds = checkedPkgs.value.map((item) => item.template_id);
      await moveOutTemplateFromPackage(
        spaceId.value,
        currentTemplateSpace.value,
        [props.id],
        selectedPkgs.value,
        isAcrossChecked.value,
        typeof props.currentPkg === 'string' ? 0 : props.currentPkg,
        props.currentPkg === 'no_specified',
        appIds,
      );
      emits('movedOut');
      close();
      Message({
        theme: 'success',
        message: t('移出套餐成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .pkg-name {
    display: flex;
    align-items: center;
    .name-text {
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .tag {
      flex-shrink: 0;
      margin-left: 4px;
      padding: 0 8px;
      height: 22px;
      line-height: 22px;
      font-size: 12px;
      color: #3a84ff;
      background: #edf4ff;
      border: 1px solid #3a84ff4d;
      border-radius: 2px;
    }
  }
  .tips {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #63656e;
    .warn-icon {
      margin-right: 4px;
      font-size: 14px;
      color: #ff9c05;
    }
  }
  .actions-wrapper {
    padding-bottom: 20px;
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
<style lang="scss">
  .move-out-from-pkgs-dialog.bk-modal-wrapper {
    .bk-dialog-header {
      line-height: normal !important;
      .bk-dialog-title {
        white-space: pre-wrap !important;
      }
    }

    .bk-modal-footer {
      background: #ffffff;
      border-top: none;
      .bk-button {
        min-width: 88px;
      }
    }
  }
</style>
