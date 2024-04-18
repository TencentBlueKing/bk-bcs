<template>
  <bk-dialog
    width="960"
    ext-cls="import-template-config-dialog"
    :is-show="props.show"
    :close-icon="false"
    :show-head="false"
    :quick-close="false">
    <div v-bkloading="{ loading: pkgListLoading }" class="import-template-wrapper">
      <div class="template-packages">
        <PkgTree
          :bk-biz-id="props.bkBizId"
          :pkg-list="pkgList"
          :imported="importedPkgs"
          :value="selectedPkgs"
          @change="handlePkgsChange" />
      </div>
      <div class="selected-result-wrapper">
        <h4>{{ t('结果预览') }}</h4>
        <p class="tips-text">
          {{ t('已选择导入') }}
          <span class="num">{{ importedPkgs.length + selectedPkgs.length }}</span>
          {{ t('个模板套餐，可按需要切换模板版本') }}
        </p>
        <div v-if="!pkgListLoading" class="packages-list">
          <template v-if="importedPkgs.length + selectedPkgs.length > 0">
            <PkgTemplatesTable
              v-for="pkg in selectedPkgs"
              :key="pkg.template_set_id"
              :bk-biz-id="props.bkBizId"
              :pkg-list="pkgList"
              :pkg-id="pkg.template_set_id"
              :selected-versions="pkg.template_revisions"
              :conflict-tpls="conflictDetail[pkg.template_set_id] || []"
              @delete="handleDeletePkg"
              @select-version="handleSelectTplVersion(pkg.template_set_id, $event, 'new')"
              @update-versions="handleUpdateTplsVersions(pkg.template_set_id, $event, 'new')" />
            <PkgTemplatesTable
              v-for="pkg in importedPkgs"
              :key="pkg.template_set_id"
              :bk-biz-id="props.bkBizId"
              :pkg-list="pkgList"
              :disabled="true"
              :pkg-id="pkg.template_set_id"
              :selected-versions="pkg.template_revisions"
              @select-version="handleSelectTplVersion(pkg.template_set_id, $event, 'imported')"
              @update-versions="handleUpdateTplsVersions(pkg.template_set_id, $event, 'imported')" />
          </template>
          <bk-exception v-else scene="part" type="empty">
            <div class="empty-tips">
              {{ t('暂无预览') }}
              <p>{{ t('请先从左侧选择导入的模板套餐') }}</p>
            </div>
          </bk-exception>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="footer-wrapper">
        <p v-if="hasConflict" class="conflict-tip">{{ t('检测到模板冲突，无法导入') }}</p>
        <div class="action-btns">
          <bk-button theme="primary" :disabled="isImportBtnDisabled" :loading="pending" @click="handleImportConfirm">
            {{ t('导入') }}
          </bk-button>
          <bk-button @click="close">{{ t('取消') }}</bk-button>
        </div>
      </div>
    </template>
  </bk-dialog>
</template>
<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter } from 'vue-router';
  import Message from 'bkui-vue/lib/message';
  import { ITemplateBoundByAppData } from '../../../../../../../../../types/config';
  import { IAllPkgsGroupBySpaceInBiz } from '../../../../../../../../../types/template';
  import {
    checkAppTemplateBinding,
    importTemplateConfigPkgs,
    updateTemplateConfigPkgs,
  } from '../../../../../../../../api/config';
  import { getAllPackagesGroupBySpace, getAppPkgBindingRelations } from '../../../../../../../../api/template';
  import PkgTree from './pkg-tree.vue';
  import PkgTemplatesTable from './pkg-templates-table.vue';

  const route = useRoute();
  const router = useRouter();
  const { t } = useI18n();

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>();

  const emits = defineEmits(['update:show', 'imported']);

  const pkgListLoading = ref(false);
  const pkgList = ref<IAllPkgsGroupBySpaceInBiz[]>([]);
  const bindingId = ref(0); // 模板和服务的绑定关系id，不为0表示绑定关系已经存在，编辑时需要调用编辑接口
  const importedPkgsLoading = ref(false);
  const importedPkgs = ref<ITemplateBoundByAppData[]>([]);
  const selectedPkgs = ref<ITemplateBoundByAppData[]>([]);
  const pending = ref(false);
  const conflictDetail = ref<{ [key: number]: number[] }>([]);

  const hasConflict = computed(() => selectedPkgs.value.some((pkg) => conflictDetail.value[pkg.template_set_id]));

  const isImportBtnDisabled = computed(() => {
    const res =
      hasConflict.value || pkgListLoading.value || importedPkgs.value.length + selectedPkgs.value.length === 0;
    return res;
  });

  watch(
    () => props.show,
    async (val) => {
      if (val) {
        bindingId.value = 0;
        conflictDetail.value = [];
        importedPkgs.value = [];
        selectedPkgs.value = [];
        getImportedPkgsData();
        await getPkgList();
        if (route.query.pkg_id && /\d+/.test(route.query.pkg_id as string)) {
          const id = Number(route.query.pkg_id);
          pkgList.value.some((spaceGroup) =>
            spaceGroup.template_sets.some((pkg) => {
              if (
                pkg.template_set_id === id &&
                importedPkgs.value.findIndex((item) => item.template_set_id === id) === -1
              ) {
                selectedPkgs.value.push({
                  template_set_id: id,
                  template_revisions: [],
                });
                return true;
              }
              return undefined;
            }),
          );
        }
      }
    },
  );

  const getPkgList = async () => {
    pkgListLoading.value = true;
    const res = await getAllPackagesGroupBySpace(props.bkBizId, { app_id: props.appId });
    pkgList.value = res.details;
    pkgListLoading.value = false;
  };

  const getImportedPkgsData = async () => {
    importedPkgsLoading.value = true;
    const res = await getAppPkgBindingRelations(props.bkBizId, props.appId);
    if (res.details.length === 1) {
      bindingId.value = res.details[0].id;
      importedPkgs.value = res.details[0].spec.bindings;
    } else {
      bindingId.value = 0;
      importedPkgs.value = [];
    }
    importedPkgsLoading.value = false;
  };

  const handlePkgsChange = (pkgs: ITemplateBoundByAppData[]) => {
    selectedPkgs.value = pkgs;
  };

  const handleDeletePkg = (id: number) => {
    const index = selectedPkgs.value.findIndex((item) => item.template_set_id === id);
    if (index > -1) {
      selectedPkgs.value.splice(index, 1);
    }
  };

  // 更新套餐下某个模板选择的版本
  const handleSelectTplVersion = (
    pkgId: number,
    version: { template_id: number; template_revision_id: number; is_latest: boolean },
    type: string,
  ) => {
    const pkgs = type === 'imported' ? importedPkgs.value : selectedPkgs.value;
    const pkgData = pkgs.find((item) => item.template_set_id === pkgId);
    if (pkgData) {
      const index = pkgData.template_revisions.findIndex((item) => item.template_id === version.template_id);
      if (index > -1) {
        pkgData.template_revisions.splice(index, 1, version);
      } else {
        pkgData.template_revisions.push(version);
      }
    }
  };

  // 批量更新套餐下所有模板所选择的版本
  const handleUpdateTplsVersions = (
    pkgId: number,
    versionsData: { template_id: number; template_revision_id: number; is_latest: boolean }[],
    type: string,
  ) => {
    const pkgs = type === 'imported' ? importedPkgs.value : selectedPkgs.value;
    const pkgData = pkgs.find((item) => item.template_set_id === pkgId);
    if (pkgData) {
      pkgData.template_revisions = versionsData;
    }
  };

  const handleImportConfirm = async () => {
    try {
      pending.value = true;
      conflictDetail.value = await checkAppTemplateBinding(props.bkBizId, props.appId, {
        bindings: selectedPkgs.value.concat(importedPkgs.value),
      });
      if (hasConflict.value) {
        pending.value = false;
        return;
      }
      if (bindingId.value) {
        await updateTemplateConfigPkgs(props.bkBizId, props.appId, bindingId.value, {
          bindings: selectedPkgs.value.concat(importedPkgs.value),
        });
      } else {
        await importTemplateConfigPkgs(props.bkBizId, props.appId, { bindings: selectedPkgs.value });
      }
      emits('imported');
      close();
      Message({
        theme: 'success',
        message: t('配置文件导入成功'),
      });
    } catch (e) {
      console.log(e);
    } finally {
      pending.value = false;
    }
  };

  const close = () => {
    if (route.query.pkg_id) {
      router.replace({ name: 'service-config', params: route.params });
    }
    emits('update:show', false);
  };
</script>
<style lang="scss" scoped>
  .import-template-wrapper {
    display: flex;
    align-items: flex-start;
    height: calc(100vh * 0.7 - 48px);
    h4 {
      margin: 0 0 25px;
      padding: 0 24px;
      line-height: 22px;
      font-size: 14px;
      font-weight: 400;
      color: #313238;
    }
    .template-packages {
      padding: 20px 0;
      width: 50%;
      height: 100%;
    }
    .selected-result-wrapper {
      padding: 12px 0;
      width: 50%;
      height: 100%;
      border-left: 1px solid #dcdee5;
      h4 {
        margin: 0 0 12px;
        padding: 0 24px;
        line-height: 22px;
        font-size: 14px;
        font-weight: 400;
        color: #313238;
      }
      .tips-text {
        margin: 0 0 18px;
        padding: 0 24px;
        line-height: 16px;
        font-size: 12px;
        color: #63656e;
        .num {
          color: #3a84ff;
        }
      }
      .packages-list {
        padding: 0 24px;
        height: calc(100% - 68px);
        overflow: auto;
        .empty-tips {
          font-size: 14px;
          color: #63656e;
          & > p {
            margin: 8px 0 0;
            color: #979ba5;
            font-size: 12px;
          }
        }
      }
    }
  }
  .footer-wrapper {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    .conflict-tip {
      margin: 0 12px 0 0;
      font-size: 12px;
      color: #ff9c01;
    }
    .action-btns {
      .bk-button:not(:last-child) {
        margin-right: 8px;
      }
    }
  }
</style>
<style lang="scss">
  .import-template-config-dialog.bk-modal-wrapper {
    .bk-modal-header,
    .bk-modal-close {
      display: none;
    }
    .bk-modal-content {
      padding: 0;
    }
  }
</style>
