<script lang="ts" setup>
  import { computed, ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../../../../../store/global'
  import { useTemplateStore } from '../../../../../../../../store/template'
  import { IConfigEditParams } from '../../../../../../../../../types/config'
  import { IPackagesCitedByApps } from '../../../../../../../../../types/template'
  import { getUnNamedVersionAppsBoundByPackages } from '../../../../../../../../api/template'
  import LinkToApp from '../../../../components/link-to-app.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const { currentTemplateSpace, currentPkg, packageList } = storeToRefs(useTemplateStore())

  const props = defineProps<{
    show: boolean;
    configForm: IConfigEditParams;
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const selectedPkgs = ref<number[]>([])
  const formRef = ref()
  const loading = ref(false)
  const pending = ref(false)
  const citedList = ref<IPackagesCitedByApps[]>([])

  const tips = computed(() => {
    return selectedPkgs.value.includes(0)
      ? '若未指定套餐，此配置项模板将无法被服务引用。后续请使用「添加至」或「添加已有配置项」功能添加至指定套餐'
      : '以下服务配置的未命名版本引用目标套餐的内容也将更新'
  })

  watch(() => props.show, val => {
    if (val) {
      selectedPkgs.value = typeof currentPkg.value === 'number' ? [currentPkg.value] : []
      pending.value = false
      if (selectedPkgs.value.length > 0) {
        getCitedData()
      }
    }
  })

  const allOptions = computed(() => {
    const pkgs = packageList.value.map(item => {
      const { id, spec } = item
      return { id, name: spec.name }
    })

    pkgs.push({ id: 0, name: '未指定套餐' })

    return pkgs
  })

  const handleSelectPkg = (val: number[], modelVal: number[]) => {
    console.log(val, modelVal)
    const currentHasNotSpecified = val.includes(0)
    const preHasNotSpecified = modelVal.includes(0)
    if (currentHasNotSpecified) {
      selectedPkgs.value = preHasNotSpecified ? val.filter(id => id !== 0) : [0]
    } else {
      selectedPkgs.value = val.slice()
      getCitedData()
    }
  }

  const getCitedData = async() => {
    loading.value = true
    const params = {
      start: 0,
      all: true
    }
    const res = await getUnNamedVersionAppsBoundByPackages(spaceId.value, currentTemplateSpace.value, selectedPkgs.value, params)
    console.log(res)
    loading.value = false
  }

  const handleConfirm = async() => {
    const isValid = await formRef.value.validate()
    if (!isValid) return
    pending.value = true
    emits('confirm', selectedPkgs.value)
  }

  const close = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    title="创建至套餐"
    ext-cls="create-to-pkg-dialog"
    confirm-text="创建"
    :width="640"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @confirm="handleConfirm"
    @closed="close">
    <template #header>
      <div class="header-wrapper">
        <div class="title">创建至套餐</div>
        <div class="config-name">{{ props.configForm.name }}</div>
      </div>
    </template>
    <bk-form ref="formRef" form-type="vertical" :model="{ pkgs: selectedPkgs }">
      <bk-form-item label="模板空间描述" property="pkgs" required>
        <bk-select
          multiple
          :popover-options="{ theme: 'light bk-select-popover create-to-pkg-selector-popover' }"
          :model-value="selectedPkgs"
          @change="handleSelectPkg">
          <bk-option
            v-for="pkg in allOptions"
            :key="pkg.id"
            :value="pkg.id"
            :label="pkg.name">
          </bk-option>
        </bk-select>
      </bk-form-item>
    </bk-form>
    <p class="tips">{{ tips }}</p>
    <bk-loading style="min-height: 200px;" :loading="loading">
      <bk-table v-if="!selectedPkgs.includes(0)" :data="citedList">
        <bk-table-column label="模板套餐" prop="template_set_name"></bk-table-column>
        <bk-table-column label="使用此套餐的服务">
          <template #default="{ row }">
            <div class="app-info">
              <div class="name">{{ row.app_name }}</div>
              <LinkToApp :id="row.app_id" />
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .title {
      margin-right: 16px;
      padding-right: 16px;
      line-height: 24px;
      border-right: 1px solid #dcdee5;
    }
    .config-name {
      flex: 1;
      line-height: 24px;
      color: #979ba5;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
  .tips {
    margin: 0 0 16px;
    font-size: 12px;
    color: #63656e;
  }
  .app-info {
    display: flex;
    align-items: center;
    .share-icon {
      font-size: 16px;
    }
  }
</style>
<style lang="scss">
  .create-to-pkg-selector-popover {
    .bk-select-option:last-of-type {
      margin-top: 10px;
      border-top: 1px solid #dcdee5;
    }
  }
</style>
