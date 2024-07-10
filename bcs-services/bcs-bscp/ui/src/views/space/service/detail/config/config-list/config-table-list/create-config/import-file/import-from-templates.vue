<template>
  <div class="wrap">
    <div class="label">{{ $t('选择模板') }}</div>
    <bk-select
      :model-value="allImportPkgs"
      custom-content
      multiple-mode="tag"
      multiple
      :remote-method="handleSearchPkg"
      :search-placeholder="$t('请输入空间/套餐名称')"
      :popover-options="{ theme: 'light bk-select-popover pkg-selector-popover' }"
      @tag-remove="handleDeletePkg"
      @clear="selectedPkgs = []">
      <template #tag>
        <bk-tag v-for="pkg in importedPkgs" :key="pkg.template_set_id">
          {{ pkg.template_set_name }}
        </bk-tag>
        <bk-tag
          v-for="pkg in selectedPkgs"
          :key="pkg.template_set_id"
          closable
          @close="handleDeletePkg(pkg.template_set_id)">
          {{ pkg.template_set_name }}
        </bk-tag>
      </template>
      <PkgTree
        :bk-biz-id="props.bkBizId"
        :pkg-list="pkgList"
        :imported="importedPkgs"
        :value="selectedPkgs"
        :search-str="searchPkgStr"
        @change="handlePkgsChange" />
      <template #extension>
        <div class="link-btn" @click="handleLinkToTemplates">
          <Share class="icon" />
          <span>{{ $t('管理模板') }}</span>
        </div>
      </template>
    </bk-select>
  </div>
  <div class="table-list" v-if="allImportPkgs.length > 0">
    <div class="tips">
      {{ $t('已选择导入') }}
      <span class="count">{{ allImportPkgs.length }}</span>
      {{ $t('个模板套餐，可按需要切换模板版本') }}
    </div>
    <div v-if="!pkgListLoading" class="packages-list">
      <PkgTemplatesTable
        v-for="pkg in selectedPkgs"
        :key="pkg.template_set_id"
        :bk-biz-id="props.bkBizId"
        :pkg-list="pkgList"
        :pkg-id="pkg.template_set_id"
        :selected-versions="pkg.template_revisions"
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
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, onMounted, computed } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { Share } from 'bkui-vue/lib/icon';
  import { ITemplateBoundByAppData } from '../../../../../../../../../../types/config';
  import { IAllPkgsGroupBySpaceInBiz } from '../../../../../../../../../../types/template';
  import { importTemplateConfigPkgs, updateTemplateConfigPkgs } from '../../../../../../../../../api/config';
  import { getAllPackagesGroupBySpace, getAppPkgBindingRelations } from '../../../../../../../../../api/template';
  import PkgTree from './pkg-tree.vue';
  import PkgTemplatesTable from './pkg-templates-table.vue';

  const route = useRoute();
  const router = useRouter();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const pkgListLoading = ref(false);
  const pkgList = ref<IAllPkgsGroupBySpaceInBiz[]>([]);
  const bindingId = ref(0); // 模板和服务的绑定关系id，不为0表示绑定关系已经存在，编辑时需要调用编辑接口
  const importedPkgsLoading = ref(false);
  const importedPkgs = ref<ITemplateBoundByAppData[]>([]);
  const selectedPkgs = ref<ITemplateBoundByAppData[]>([]);
  const searchPkgStr = ref('');
  const allImportPkgs = computed(() => [...importedPkgs.value, ...selectedPkgs.value]);

  const isImportBtnDisabled = computed(() => {
    return pkgListLoading.value || importedPkgs.value.length + selectedPkgs.value.length === 0;
  });

  onMounted(async () => {
    bindingId.value = 0;
    importedPkgs.value = [];
    selectedPkgs.value = [];
    await getPkgList();
    getImportedPkgsData();
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
  });

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
      pkgList.value.some((templateSpace) => {
        importedPkgs.value.forEach((importedPkg) => {
          templateSpace.template_sets.some((pkg) => {
            if (pkg.template_set_id === importedPkg.template_set_id) {
              importedPkg.template_set_name = `${templateSpace.template_space_name} - ${pkg.template_set_name}`;
            }
            return undefined;
          });
        });
        return undefined;
      });
    } else {
      bindingId.value = 0;
      importedPkgs.value = [];
    }
    importedPkgsLoading.value = false;
  };

  const handleSearchPkg = (val: string) => {
    searchPkgStr.value = val;
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
    if (bindingId.value) {
      await updateTemplateConfigPkgs(props.bkBizId, props.appId, bindingId.value, {
        bindings: selectedPkgs.value.concat(importedPkgs.value),
      });
    } else {
      await importTemplateConfigPkgs(props.bkBizId, props.appId, { bindings: selectedPkgs.value });
    }
    close();
  };

  const close = () => {
    if (route.query.pkg_id) {
      router.replace({ name: 'service-config', params: route.params });
    }
  };

  const handleLinkToTemplates = () => {
    setTimeout(() => {
      router.push({ name: 'templates-list' });
    }, 300);
  };

  defineExpose({
    isImportBtnDisabled: isImportBtnDisabled.value,
    handleImportConfirm,
  });
</script>
<style lang="scss" scoped>
  .table-list {
    margin-top: 24px;
    border-top: 1px solid #dcdee5;
    overflow: auto;
    max-height: 490px;
    .tips {
      margin: 16px 0;
      font-size: 12px;
      color: #63656e;
      .count {
        color: #3a84ff;
      }
    }
  }

  .bk-select {
    width: 818px;
  }
</style>
<style lang="scss">
  .pkg-selector-popover {
    width: 420px !important;
    .bk-select-extension {
      justify-content: center;
      .link-btn {
        display: flex;
        align-items: center;
        gap: 6px;
        color: #63656e;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
</style>
