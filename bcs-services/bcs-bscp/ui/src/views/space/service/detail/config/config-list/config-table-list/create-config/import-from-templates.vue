<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { Message } from 'bkui-vue';
  import { ITemplateBoundByAppData } from '../../../../../../../../../types/config'
  import { IAllPkgsGroupBySpaceInBiz } from '../../../../../../../../../types/template'
  import { importTemplateConfigPkgs } from '../../../../../../../../api/config';
  import { getAllPackagesGroupBySpace, getAppPkgBindingRelations } from '../../../../../../../../api/template';
  import PkgTree from './pkg-tree.vue';
  import PkgTemplatesTable from './pkg-templates-table.vue';

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>()

  const emits = defineEmits(['update:show', 'imported'])

  const pkgListLoading = ref(false)
  const pkgList = ref<IAllPkgsGroupBySpaceInBiz[]>([])
  const importedPkgsLoading = ref(false)
  const importedPkgs = ref<ITemplateBoundByAppData[]>([])
  const selectedPkgs = ref<ITemplateBoundByAppData[]>([])
  const expandedPkg = ref(0)
  const pending = ref(false)

  watch(() => props.show, val => {
    if (val) {
      importedPkgs.value = []
      selectedPkgs.value = []
      getPkgList()
      getImportedPkgsData()
    }
  })

  const getPkgList = async () => {
    pkgListLoading.value = true
    const res = await getAllPackagesGroupBySpace(props.bkBizId)
    pkgList.value = res.details
    pkgListLoading.value = false
  }

  const getImportedPkgsData = async () => {
    importedPkgsLoading.value = true
    const res = await getAppPkgBindingRelations(props.bkBizId, props.appId)
    if (res.details.length === 1) {
      importedPkgs.value = res.details[0].spec.bindings
    } else {
      importedPkgs.value = []
    }
    importedPkgsLoading.value = false
  }

  const handlePkgsChange = (pkgs: ITemplateBoundByAppData[]) => {
    selectedPkgs.value = pkgs
  }

  const handleDeletePkg = (id: number) => {
    const index = selectedPkgs.value.findIndex(item => item.template_set_id === id)
    if (index > -1) {
      selectedPkgs.value.splice(index, 1)
    }
  }

  const handleExpandTable = (id: number) => {
    expandedPkg.value = expandedPkg.value === id ? 0 : id
  }

  const handleImportConfirm = async () => {
    try {
      pending.value = true
      await importTemplateConfigPkgs(props.bkBizId, props.appId, { bindings: selectedPkgs.value })
      emits('imported')
      close()
      Message({
        theme: 'success',
        message: '配置文件导入成功'
      })
    } catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const close = () => {
    emits('update:show', false)
  }
</script>
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
        <h4>结果预览</h4>
        <p class="tips-text">已选择导入 <span class="num">{{ importedPkgs.length + selectedPkgs.length }}</span> 个模板套餐，可按需要切换模板版本</p>
        <div class="packages-list">
          <template v-if="importedPkgs.length + selectedPkgs.length > 0">
            <PkgTemplatesTable
              v-for="pkg in importedPkgs"
              :key="pkg.template_set_id"
              :bk-biz-id="props.bkBizId"
              :pkg-list="pkgList"
              :disabled="true"
              :expanded="expandedPkg === pkg.template_set_id"
              :pkg-id="pkg.template_set_id"
              @expand="handleExpandTable" />
            <PkgTemplatesTable
              v-for="pkg in selectedPkgs"
              :key="pkg.template_set_id"
              :bk-biz-id="props.bkBizId"
              :pkg-list="pkgList"
              :expanded="expandedPkg === pkg.template_set_id"
              :pkg-id="pkg.template_set_id"
              @delete="handleDeletePkg"
              @expand="handleExpandTable" />
          </template>
          <bk-exception v-else scene="part" type="empty">
            <div class="empty-tips">
              暂无预览
              <p>请先从左侧选择导入的模板套餐</p>
            </div>
          </bk-exception>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="action-btns">
        <bk-button theme="primary" :disabled="pkgListLoading || selectedPkgs.length === 0" :loading="pending" @click="handleImportConfirm">导入</bk-button>
        <bk-button @click="close">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
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
</style>
<style lang="scss">
  .import-template-config-dialog.bk-dialog-wrapper {
    .bk-modal-header,
    .bk-modal-close {
      display: none;
    }
    .bk-modal-content {
      padding: 0;
    }
    .action-btns {
      .bk-button:not(:last-child) {
        margin-right: 8px;
      }
    }
  }
</style>
