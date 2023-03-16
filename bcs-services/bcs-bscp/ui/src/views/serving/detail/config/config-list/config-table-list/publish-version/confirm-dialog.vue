<script setup lang="ts">
  import { defineProps, defineEmits, ref, watch } from 'vue'
  import InfoBox from "bkui-vue/lib/info-box";
  import { publishVersion } from '../../../../../../../api/config'
  import CategoryGroupSelect from '../../../../group/components/category-group-select.vue';

  interface IFormData {
    groups: Array<number>;
    all: boolean;
    memo: string;
  }

  const props = defineProps<{
    show: boolean,
    bkBizId: string,
    appId: number,
    releaseId: number|null,
    groups: Array<number>
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const localVal = ref<IFormData>({
    groups: [],
    all: false,
    memo: ''
  })
  const pending = ref(false)
  const formRef = ref()
  const rules = {
    groups: [
      {
        validator: (value: string) => value.length > 0,
        message: '分组不能为空'
      }
    ],
    memo: [
      {
        validator: (value: string) => value.length <= 100,
        message: '最大长度100个字符'
      }
    ]
  }

  watch(() => props.groups, (val) => {
    localVal.value.groups = [...val]
  }, { immediate: true })

  // 选择分组
  const handleGroupChange = (group: number[]) => {
    localVal.value.groups = group
  }

  const handleClose = () => {
    emits('update:show', false)
    localVal.value = {
      groups: [],
      all: false,
      memo: ''
    }
  }

  const handleConfirm = async() => {
    try {
      pending.value = true
      await formRef.value.validate()
      await publishVersion(props.bkBizId, props.appId, <number>props.releaseId, localVal.value)
      emits('confirm')
      handleClose()
    } catch (e) {
      console.error(e)
      // InfoBox({
      // // @ts-ignore
      //   infoType: "danger",
      //   title: '版本上线失败',
      //   subTitle: e.response.data.error.message,
      //   confirmText: '重试',
      //   onConfirm () {
      //     handleConfirm()
      //   }
      // })
    } finally {
      pending.value = false
    }
  }

</script>
<template>
  <bk-dialog
    title="上线版本"
    ext-cls="release-version-dialog"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    :is-loading="pending"
    @closed="handleClose"
    @confirm="handleConfirm">
      <bk-form class="form-wrapper" form-type="vertical" ref="formRef" :rules="rules" :model="localVal">
          <bk-form-item label="上线分组" property="groups">
            <CategoryGroupSelect :app-id="props.appId" :multiple="true" :value="localVal.groups" @change="handleGroupChange" />
          </bk-form-item>
          <bk-form-item label="上线说明" property="memo">
            <bk-input v-model="localVal.memo" type="textarea" :maxlength="100"></bk-input>
          </bk-form-item>
      </bk-form>
      <template #footer>
        <div class="dialog-footer">
          <bk-button theme="primary" :loading="pending" @click="handleConfirm">确定上线</bk-button>
          <bk-button :disabled="pending" @click="handleClose">取消</bk-button>
        </div>
      </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .form-wrapper {
    padding-bottom: 24px;
  }
  .dialog-footer {
    .bk-button {
      margin-left: 8px;
    }
  }
</style>
<style lang="scss">
  .release-version-dialog.bk-dialog-wrapper .bk-dialog-header {
    padding-bottom: 20px;
  }
</style>