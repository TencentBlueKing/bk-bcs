<template>
  <div>
    <bk-date-picker
      v-click-outside="() => (open = false)"
      ref="datePickerRef"
      style="width: 300"
      :disabled-date="disabledDate"
      :model-value="defaultValue"
      :shortcuts="shortcutsRange"
      :editable="false"
      :clearable="false"
      :open="open"
      type="datetimerange"
      append-to-body
      clear
      @change="handleChange"
      @click="open = !open">
      <template #confirm>
        <div>
          <bk-button theme="primary" @click="open = false" style="margin-right: 8px; height: 26px"> 确定 </bk-button>
        </div>
      </template>
    </bk-date-picker>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';

  const shortcutsRange = ref([
    // {
    //   text: '今天',
    //   value() {
    //     const end = new Date();
    //     const start = new Date(end.getFullYear(), end.getMonth(), end.getDate());
    //     return [start, end];
    //   },
    //   onClick: (picker) => {
    //     console.log(picker);
    //   },
    // },
    {
      text: '近7天',
      value() {
        const end = new Date();
        const start = new Date();
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
        return [start, end];
      },
      onClick: (picker) => {
        console.log(picker);
      },
    },
    {
      text: '近15天',
      value() {
        const end = new Date();
        const start = new Date();
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 15);
        return [start, end];
      },
      onClick: (picker) => {
        console.log(picker);
      },
    },
    {
      text: '近30天',
      value() {
        const end = new Date();
        const start = new Date();
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
        return [start, end];
      },
      onClick: (picker) => {
        console.log(picker);
      },
    },
  ]);
  const open = ref(false);
  const datePickerRef = ref(null);
  const defaultValue = ref([new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), new Date()]);

  const handleChange = (date: []) => {
    defaultValue.value = date;
  };

  const disabledDate = (date: Date) => date && date.valueOf() > Date.now() - 86400;
</script>

<style scoped></style>
