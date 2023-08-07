<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../../../../store/global'
  import { deleteTemplatePackage } from '../../../../../../api/template'
  import { ITemplatePackageItem } from '../../../../../../../types/template'
  import { IAppItem } from '../../../../../../../types/app'

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    show: boolean,
    templateSpaceId: number;
    pkg: ITemplatePackageItem
  }>()

  const emits = defineEmits(['update:show', 'deleted'])

  const isShow = ref(false)
  const appsLoading = ref(false)
  const appList = ref<IAppItem[]>([])
  const pending = ref(false)

  watch(() => props.show, val => {
    isShow.value = val

  })

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    :title="`确认删除【${props.pkg ? props.pkg.spec.name : ''}】?`"
    ext-cls="delete-template-package-dialog"
    header-align="center"
    dialog-type="show"
    :width="400"
    :is-show="isShow"
    :esc-close="false"
    :quick-close="false"
    @closed="handleClose">
    <p class="tips">以下服务的未命名版本中引用此套餐的内容也将删除</p>
    <div class="service-table">
      <bk-table :data="appList">
        <bk-table-column label="引用此套餐的服务"></bk-table-column>
      </bk-table>
    </div>
    <div class="action-btns">
      <bk-button theme="primary" class="delete-btn">删除套餐</bk-button>
      <bk-button>取消</bk-button>
    </div>
  </bk-dialog>
</template>
<style lang="postcss" scoped>
  .tips {
    margin: 0 0 16px;
    font-size: 14px;
    line-height: 22px;
    color: #63656e;
    text-align: center;
  }
  .action-btns {
    margin-top: 32px;
    text-align: center;
    .delete-btn {
      margin-right: 8px;
    }
  }
</style>
