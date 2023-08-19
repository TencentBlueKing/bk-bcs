<script lang="ts" setup>
  import { ref } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue';
  import { useGlobalStore } from '../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../store/template'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import { updateTemplatePackage } from '../../../../../../../api/template'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { packageList, currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
    value: ITemplateConfigItem[];
  }>()

  const emits = defineEmits(['update:show', 'movedOut'])

  const pending = ref(false)

  const handleConfirm = async() => {
    const pkg = packageList.value.find(item => item.id === currentPkg.value)
    if (!pkg) return

    try {
      pending.value = true
      const { name, memo, public: isPublic, bound_apps, template_ids } = pkg.spec
      const ids = template_ids.filter(id => props.value.findIndex(item => item.id === id) === -1)
      const params = {
        name,
        memo,
        template_ids: ids,
        bound_apps,
        public: isPublic
      }
      await updateTemplatePackage(spaceId.value, currentTemplateSpace.value, <number>currentPkg.value, params)
      emits('movedOut')
      close()
      Message({
        theme: 'success',
        message: '配置项移出套餐成功'
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
    ext-cls="move-out-configs-dialog"
    title="批量移出当前套餐"
    confirm-text="确定移出"
    :width="480"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <div class="selected-mark">已选 <span class="num">{{ props.value.length }}</span> 个配置项</div>
    <p class="tips">以下服务配置的未命名版本引用目标套餐的内容也将更新</p>
    <bk-table>
      <bk-table-column label="所在模板套餐"></bk-table-column>
      <bk-table-column label="使用此套餐的服务"></bk-table-column>
    </bk-table>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .selected-mark {
    display: inline-block;
    margin-bottom: 16px;
    padding: 0 12px;
    height: 32px;
    line-height: 32px;
    border-radius: 16px;
    font-size: 12px;
    color: #63656e;
    background: #f0f1f5;
    .num {
      color: #3a84ff;
    }
  }
</style>
