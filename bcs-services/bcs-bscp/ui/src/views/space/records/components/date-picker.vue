<template>
  <div>
    <bk-date-picker
      ref="datePickerRef"
      style="width: 300px"
      type="datetimerange"
      append-to-body
      :disabled-date="disabledDate"
      :model-value="defaultValue"
      :shortcuts="shortcutsRange"
      :editable="false"
      :clearable="false"
      :open="open"
      @change="change"
      @click="open = !open">
      <template #confirm>
        <div>
          <bk-button class="primary-button" theme="primary" @click="handleChange"> 确定 </bk-button>
        </div>
      </template>
    </bk-date-picker>
  </div>
</template>

<script setup lang="ts">
  import { onBeforeMount, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import dayjs from 'dayjs';

  const emits = defineEmits(['changeTime']);

  const route = useRoute();
  const router = useRouter();

  const shortcutsRange = ref([
    {
      text: '近7天',
      value() {
        const end = dayjs().toDate();
        const start = dayjs().subtract(7, 'day').toDate();
        return [start, end];
      },
    },
    {
      text: '近15天',
      value() {
        const end = dayjs().toDate();
        const start = dayjs().subtract(15, 'day').toDate();
        return [start, end];
      },
    },
    {
      text: '近30天',
      value() {
        const end = dayjs().toDate();
        const start = dayjs().subtract(30, 'day').toDate();
        return [start, end];
      },
    },
  ]);
  const open = ref(false);
  const datePickerRef = ref(null);
  const defaultValue = ref<string[]>([]);

  onBeforeMount(() => {
    const hasQueryTime = ['start_time', 'end_time'].every(
      (key) => key in route.query && route.query[key]?.length === 19,
    );
    const defaultTimeRange = [
      dayjs().subtract(7, 'day').format('YYYY-MM-DD HH:mm:ss'),
      dayjs().format('YYYY-MM-DD HH:mm:ss'),
    ];
    if (hasQueryTime) {
      let startTime = dayjs(String(route.query.start_time));
      let endTime = dayjs(String(route.query.end_time));
      // 验证时间格式且在当前时间以前
      const isValidTime = startTime.isValid() && endTime.isValid() && startTime.isBefore() && endTime.isBefore();
      if (isValidTime) {
        if (startTime.isAfter(endTime)) [startTime, endTime] = [endTime, startTime];
        defaultValue.value = [startTime.format('YYYY-MM-DD HH:mm:ss'), endTime.format('YYYY-MM-DD HH:mm:ss')];
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

<style lang="scss" scoped>
  .primary-button {
    margin-right: 4px;
    height: 26px;
  }
</style>
