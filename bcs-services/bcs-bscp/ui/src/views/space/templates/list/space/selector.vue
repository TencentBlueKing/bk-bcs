<template>
  <div class="space-selector">
    <bk-select
      ref="selectorRef"
      :search-placeholder="t('搜索空间')"
      filterable
      :input-search="false"
      :model-value="currentTemplateSpace"
      :popover-options="{ theme: 'light bk-select-popover template-space-selector-popover' }"
      @toggle="selectorOpen = $event"
      @change="handleSelect">
      <template #trigger>
        <div class="select-trigger">
          <h5 class="space-name" :title="templateSpaceDetail.name">{{ templateSpaceDetail.name }}</h5>
          <bk-overflow-title type="tips" class="space-desc">{{ templateSpaceDetail.memo || '--' }}</bk-overflow-title>
          <DownShape :class="['triangle-icon', { up: selectorOpen }]" />
        </div>
      </template>
      <bk-option v-for="item in spaceList" :key="item.id" :value="item.id" :label="item.spec.name">
        <div class="space-option-item">
          <div class="name-text">{{ item.spec.name }}</div>
          <div class="actions">
            <template v-if="!['default_space', '默认空间'].includes(item.spec.name)">
              <i class="bk-bscp-icon icon-edit-small" @click.stop="handleEditOpen(item)" />
              <Del class="delete-icon" @click.stop="handleDelete(item)" />
            </template>
          </div>
        </div>
      </bk-option>
      <template #extension>
        <div class="create-space-extension" @click="handleCreateOpen">
          <i class="bk-bscp-icon icon-add"></i>
          {{ t('创建空间') }}
        </div>
      </template>
    </bk-select>
  </div>
  <Create v-model:show="isShowCreateDialog" @created="handleCreated" />
  <Edit v-model:show="editingData.open" :data="editingData.data" @edited="loadList" />
  <DeleteConfirmDialog
    v-model:is-show="isDeleteTemplateSpaceDialogShow"
    :title="t('确认删除该配置模板空间？')"
    @confirm="handleDeleteTemplateSpaceConfirm">
    <div style="margin-bottom: 8px">
      {{ t('配置模板空间') }}:
      <span style="color: #313238; font-weight: 600">{{ deleteTemplateSpaceItem?.spec.name }}</span>
    </div>
    <div>{{ t('一旦删除，该操作将无法撤销，请谨慎操作') }}</div>
  </DeleteConfirmDialog>
</template>
<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { DownShape, Del } from 'bkui-vue/lib/icon';
  import { InfoBox, Message } from 'bkui-vue';
  import useGlobalStore from '../../../../../store/global';
  import useTemplateStore from '../../../../../store/template';
  import { ITemplateSpaceItem } from '../../../../../../types/template';
  import { ICommonQuery } from '../../../../../../types/index';
  import {
    getTemplateSpaceList,
    getTemplatesBySpaceId,
    deleteTemplateSpace,
    getTemplatePackageList,
  } from '../../../../../api/template';
  import DeleteConfirmDialog from '../../../../../components/delete-confirm-dialog.vue';

  import Create from './create.vue';
  import Edit from './edit.vue';

  const router = useRouter();
  const { spaceId } = storeToRefs(useGlobalStore());
  const templateStore = useTemplateStore();
  const { currentTemplateSpace, templateSpaceList } = storeToRefs(templateStore);
  const { t } = useI18n();

  const loading = ref(false);
  const spaceList = ref<ITemplateSpaceItem[]>([]);
  const selectorOpen = ref(false);
  const selectorRef = ref();
  const isShowCreateDialog = ref(false);
  const templatesLoading = ref(false);
  const isDeleteTemplateSpaceDialogShow = ref(false);
  const deleteTemplateSpaceItem = ref<ITemplateSpaceItem>();
  const editingData = ref({
    open: false,
    data: { id: 0, name: '', memo: '' },
  });

  const templateSpaceDetail = computed(() => {
    const item = templateSpaceList.value.find((item) => item.id === currentTemplateSpace.value);
    if (item) {
      const { name, memo } = item.spec;
      return { name, memo };
    }
    return { name: '', memo: '' };
  });

  onMounted(() => {
    initData();
  });

  const initData = async () => {
    await loadList();
    if (!currentTemplateSpace.value) {
      let tplSpaceId = 0;
      const lastAccessedTplSpaceDetail = localStorage.getItem('lastAccessedTplSpaceDetail');
      const id = lastAccessedTplSpaceDetail ? JSON.parse(lastAccessedTplSpaceDetail)?.id : 0;
      // 当前业务下是否存在改模板空间
      const isTplSpaceExist = spaceList.value.some((item) => item.id === id);
      if (isTplSpaceExist) {
        tplSpaceId = id;
      } else if (spaceList.value.length > 0) {
        tplSpaceId = spaceList.value[0].id;
      }
      if (tplSpaceId) {
        // url中没有模版空间id，且空间列表不为空时，默认选中第一个空间
        setTemplateSpace(tplSpaceId);
        updateRouter(tplSpaceId);
      }
    } else {
      setTemplateSpace(currentTemplateSpace.value);
    }
  };

  const loadList = async () => {
    loading.value = true;
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };
    const res = await getTemplateSpaceList(spaceId.value, params);
    const index = (res.details as ITemplateSpaceItem[]).findIndex((item) =>
      ['默认空间', 'default_space'].includes(item.spec.name),
    );
    if (index > -1) {
      // 默认空间放到首位
      spaceList.value = res.details.splice(index, 1).concat(res.details);
      spaceList.value[0].spec.memo = t(
        '空间可将业务下不同使用场景的配置模板文件隔离，每个空间内的配置文件路径+配置文件名是唯一的，每个业务下会自动创建一个默认空间',
      );
    } else {
      spaceList.value = res.details;
    }
    templateStore.$patch((state) => {
      state.templateSpaceList = spaceList.value;
    });
    loading.value = false;
  };

  const handleCreateOpen = () => {
    isShowCreateDialog.value = true;
    selectorRef.value.hidePopover();
  };

  const handleEditOpen = (space: ITemplateSpaceItem) => {
    const { id, spec } = space;
    editingData.value = {
      open: true,
      data: {
        id,
        name: spec.name,
        memo: spec.memo,
      },
    };
    selectorRef.value.hidePopover();
  };

  const handleCreated = (id: number) => {
    const { href } = router.resolve({ name: 'templates-list', params: { templateSpaceId: id } });
    window.location.href = href;
  };

  const handleDelete = async (space: ITemplateSpaceItem) => {
    templatesLoading.value = true;
    const params = {
      start: 0,
      limit: 1,
      all: true,
    };
    const packageParams = {
      start: 0,
      all: true,
    };
    const res = await getTemplatesBySpaceId(spaceId.value, space.id, params);
    const packageRes = await getTemplatePackageList(spaceId.value, String(space.id), packageParams);
    if (res.count > 0) {
      InfoBox({
        title: `${t('未能删除')}【${space.spec.name}】`,
        'ext-cls': 'info-box-style',
        subTitle: t('请先确认删除此空间下所有配置文件'),
        dialogType: 'confirm',
        confirmText: t('我知道了'),
      } as any);
    } else if (packageRes.count > 0) {
      InfoBox({
        title: `${t('未能删除')}【${space.spec.name}】`,
        'ext-cls': 'info-box-style',
        subTitle: t('请先确认删除此空间下所有配置套餐'),
        dialogType: 'confirm',
        confirmText: t('我知道了'),
      } as any);
    } else {
      deleteTemplateSpaceItem.value = space;
      isDeleteTemplateSpaceDialogShow.value = true;
    }
    selectorRef.value.hidePopover();
  };

  const handleDeleteTemplateSpaceConfirm = async () => {
    await deleteTemplateSpace(spaceId.value, deleteTemplateSpaceItem.value!.id);
    if (deleteTemplateSpaceItem.value!.id === currentTemplateSpace.value) {
      templateStore.$patch((state) => {
        state.currentTemplateSpace = '';
      });
      initData();
    } else {
      loadList();
    }
    Message({
      theme: 'success',
      message: t('删除空间成功'),
    });
    isDeleteTemplateSpaceDialogShow.value = false;
  };

  const handleSelect = (id: number) => {
    setTemplateSpace(id);
    setCurrentPackage();
    updateRouter(id);
  };

  const setTemplateSpace = (id: number) => {
    localStorage.setItem('lastAccessedTplSpaceDetail', JSON.stringify({ id }));
    templateStore.$patch((state) => {
      state.currentTemplateSpace = id;
    });
  };

  const setCurrentPackage = () => {
    templateStore.$patch((state) => {
      state.currentPkg = '';
    });
  };

  const updateRouter = (id: number) => {
    router.push({ name: 'templates-list', params: { templateSpaceId: id } });
  };
</script>
<style lang="scss" scoped>
  .select-trigger {
    position: relative;
    margin: 0 16px;
    padding: 8px;
    height: 58px;
    background: #f5f7fa;
    border-radius: 2px;
    cursor: pointer;
    .space-name {
      margin: 0;
      width: calc(100% - 20px);
      font-size: 14px;
      color: #313238;
      line-height: 22px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .space-desc {
      width: calc(100% - 20px);
      font-size: 12px;
      color: #979ba5;
      line-height: 20px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .triangle-icon {
      position: absolute;
      top: 50%;
      right: 9px;
      font-size: 12px;
      color: #979ba5;
      transform: translateY(-50%);
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      &.up {
        transform: translateY(-50%) rotate(-180deg);
      }
    }
  }
  .space-option-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    width: 100%;
    &:hover {
      .actions {
        display: flex;
      }
    }
    .name-text {
      width: calc(100% - 34px);
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .actions {
      display: none;
      align-items: center;
      width: 34px;
      .icon-edit-small {
        margin-right: 4px;
        font-size: 18px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
      .delete-icon {
        font-size: 12px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
  .create-space-extension {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
      .icon-add {
        color: #3a84ff;
      }
    }
    .icon-add {
      margin-right: 5px;
      font-size: 14px;
      color: #979ba5;
    }
  }
</style>
<style lang="scss">
  .template-space-selector-popover.bk-popover.bk-pop2-content.bk-select-popover
    .bk-select-content-wrapper
    .bk-select-option {
    padding: 0;
  }
</style>
