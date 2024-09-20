<template>
  <div>
    <bk-date-picker
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
      @change="change"
      @click="open = !open">
      <template #confirm>
        <div>
          <bk-button theme="primary" @click="handleChange" style="margin-right: 8px; height: 26px"> 确定 </bk-button>
        </div>
      </template>
    </bk-date-picker>
  </div>
</template>

<script setup lang="ts">
  import { onBeforeMount, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { datetimeFormat } from '../../../../utils';

  const emits = defineEmits(['changeTime']);

  const route = useRoute();
  const router = useRouter();

  const shortcutsRange = ref([
    {
      text: '近7天',
      value() {
        const end = new Date();
        const start = new Date();
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 7);
        return [start, end];
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
    },
    {
      text: '近30天',
      value() {
        const end = new Date();
        const start = new Date();
        start.setTime(start.getTime() - 3600 * 1000 * 24 * 30);
        return [start, end];
      },
    },
  ]);
  const open = ref(false);
  const datePickerRef = ref(null);
  const defaultValue = ref<string[]>([]);

  onBeforeMount(() => {
    const hasQueryTime = ['start_time', 'end_time'].every((key) => key in route.query);
    const defaultTimeRange = [
      datetimeFormat(String(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000))),
      datetimeFormat(String(new Date())),
    ];

    if (hasQueryTime) {
      let startTime = String(route.query.start_time);
      let endTime = String(route.query.end_time);
      const isValidTime =
        !isNaN(new Date(startTime).getTime()) &&
        !isNaN(new Date(endTime).getTime()) &&
        new Date(startTime).getTime() < Date.now() &&
        new Date(endTime).getTime() < Date.now() &&
        startTime.length === 19 &&
        endTime.length === 19;

      if (isValidTime) {
        if (startTime > endTime) [startTime, endTime] = [endTime, startTime];
        defaultValue.value = [startTime, endTime];
      } else {
        defaultValue.value = defaultTimeRange;
      }
    } else {
      defaultValue.value = defaultTimeRange;
    }
    setUrlParams();
    emits('changeTime', defaultValue.value);
  });

  const change = (date: []) => {
    defaultValue.value = date;
    setUrlParams();
  };

  const setUrlParams = () => {
    router.replace({
      query: {
        ...route.query,
        start_time: defaultValue.value[0],
        end_time: defaultValue.value[1],
      },
    });
  };

  const disabledDate = (date: Date) => date && date.valueOf() > Date.now() - 86400;

  const handleChange = () => {
    emits('changeTime', defaultValue.value);
    open.value = false;
  };
</script>

<style scoped></style>
