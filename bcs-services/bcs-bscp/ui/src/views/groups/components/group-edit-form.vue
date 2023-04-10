<script setup lang="ts">
  import { ref } from 'vue'

  const props = defineProps<{}>()

  const localVal = ref({})

</script>
<template>
  <bk-form>
    <bk-form-item label="分组名称" required>
      <bk-input></bk-input>
    </bk-form-item>
    <bk-form-item label="服务可见范围" required>
      <bk-radio-group>
        <bk-radio>公开</bk-radio>
        <bk-radio>指定服务</bk-radio>
      </bk-radio-group>
      <bk-select placeholder="请选择服务"></bk-select>
    </bk-form-item>
    <bk-form-item label="分组负责" required>
      <bk-radio-group>
        <bk-radio>AND</bk-radio>
        <bk-radio>OR</bk-radio>
      </bk-radio-group>
    </bk-form-item>
  </bk-form>
</template>