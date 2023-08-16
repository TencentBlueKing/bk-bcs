<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import { useTemplateStore } from '../../../../../../../store/template'

  const { packageList, currentPkg } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
    value: ITemplateConfigItem[];
  }>()

  const emits = defineEmits(['update:show'])

  const formRef = ref()
  const selectedPkgs = ref<number[]>([])
  const pending = ref(false)

  const allPackages = computed(() => {
    return packageList.value.filter(pkg => pkg.id !== currentPkg.value)
  })

  const isMultiple = computed(() => {
    return props.value.length > 1
  })

  watch(() => props.show, val => {
    if (val) {
      selectedPkgs.value =[]
    }
  })

  const handleConfirm = async () => {
    const isValid = await formRef.value.validate()
    if (!isValid) return
  }

  const close = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    ext-cls="add-configs-to-pkg-dialog"
    confirm-text="添加"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <template #header>
      <div class="header-wrapper">
        <div class="title">{{ isMultiple ? '批量添加至' : '添加至套餐' }}</div>
        <div v-if="props.value.length === 1" class="config-name">{{ props.value[0].spec.name }}</div>
      </div>
    </template>
    <div v-if="isMultiple" class="selected-mark">已选 <span class="num">{{ props.value.length }}</span> 个配置项</div>
    <bk-form ref="formRef" form-type="vertical" :model="{ pkgs: selectedPkgs }">
      <bk-form-item :label="isMultiple ? '添加至模板套餐' : '模板套餐'" property="pkgs" required>
        <bk-select v-model="selectedPkgs" multiple>
          <bk-option
            v-for="pkg in allPackages"
            :key="pkg.id"
            :value="pkg.id"
            :label="pkg.spec.name">
          </bk-option>
        </bk-select>
      </bk-form-item>
    </bk-form>
    <p class="tips">以下服务配置的未命名版本引用目标套餐的内容也将更新</p>
    <bk-table>
      <bk-table-column label="模板套餐"></bk-table-column>
      <bk-table-column label="使用此套餐的服务"></bk-table-column>
    </bk-table>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .title {
      line-height: 24px;
    }
    .config-name {
      flex: 1;
      margin-left: 16px;
      padding-left: 16px;
      line-height: 24px;
      color: #979ba5;
      border-left: 1px solid #dcdee5;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
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
  .tips {
    margin: 0 0 16px;
    font-size: 12px;
    color: #63656e;
  }
</style>
