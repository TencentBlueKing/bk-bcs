<template>
  <div class="detail p30">
    <!-- 基础信息 -->
    <div class="detail-title">
      {{ $t('基础信息') }}
    </div>
    <div class="detail-content basic-info">
      <div class="basic-info-item">
        <label>{{ $t('命名空间') }}</label>
        <span>{{ data.metadata.namespace }}</span>
      </div>
      <div class="basic-info-item">
        <label>UID</label>
        <span class="bcs-ellipsis">{{ data.metadata.uid }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('创建时间') }}</label>
        <span>{{ extData.createTime }}</span>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('存在时间') }}</label>
        <span>{{ extData.age }}</span>
      </div>
      <div class="basic-info-item">
        <label>finalizers</label>
        <template v-if="data.metadata.finalizers">
          <bcs-popover placement="top" width="220">
            <span>{{ data.metadata.finalizers.join(',') }}</span>
            <div slot="content" style="white-space: normal;">
              <div v-for="(item, index) in data.metadata.finalizers" :key="index">
                {{ item }}
              </div>
            </div>
          </bcs-popover>
        </template>
        <template v-else>
          <span>--</span>
        </template>
      </div>
      <div class="basic-info-item">
        <label>{{ $t('被挂载') }}</label>
        <span>{{ mountInfo.join(',') || '--' }}</span>
      </div>
    </div>
    <!-- 标签、注解 -->
    <bcs-tab class="mt20" type="card" :label-height="42">
      <bcs-tab-panel name="label" :label="$t('标签')">
        <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="annotations" :label="$t('注解')">
        <bk-table :data="handleTransformObjToArr(data.metadata.annotations)">
          <bk-table-column label="Key" prop="key"></bk-table-column>
          <bk-table-column label="Value" prop="value"></bk-table-column>
        </bk-table>
      </bcs-tab-panel>
      <bcs-tab-panel name="event" :label="$t('事件')">
        <EventQueryTableVue
          hide-cluster-and-namespace
          :kinds="data.kind"
          :namespace="data.metadata.namespace"
          :name="data.metadata.name" />
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import $store from '@/store';
import EventQueryTableVue from '@/views/project-manage/event-query/event-query-table.vue';

export default defineComponent({
  name: 'PvcDetail',
  components: { EventQueryTableVue },
  props: {
    // 当前行数据
    data: {
      type: Object,
      default: () => ({}),
    },
    // 当前行对应的manifestExt数据
    extData: {
      type: Object,
      default: () => ({}),
    },
    clusterId: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const handleTransformObjToArr = (obj) => {
      if (!obj) return [];

      return Object.keys(obj).reduce<any[]>((data, key) => {
        data.push({
          key,
          value: obj[key],
        });
        return data;
      }, []);
    };

    const isLoading = ref(false);
    const mountInfo = ref([]);
    const handleGetMountInfo = async () => {
      isLoading.value = true;
      const { podNames = [] } = await $store.dispatch('dashboard/getPvcMountInfo', {
        $namespace: props.data.metadata.namespace,
        $pvcID: props.data.metadata.name,
        $clusterId: props.clusterId,
      });
      mountInfo.value = podNames;
      isLoading.value = false;
    };

    onMounted(() => {
      handleGetMountInfo();
    });

    return {
      isLoading,
      mountInfo,
      handleTransformObjToArr,
    };
  },
});
</script>
<style lang="postcss" scoped>
.detail {
  font-size: 14px;
  /deep/ .bk-tab-label-item {
      background-color: #FAFBFD;
      border-bottom: 1px solid #dcdee5;
      line-height: 41px !important;
      height: 41px;
      &.active {
          border-bottom: none;
      }
  }
  /deep/ .bk-tab-label-wrapper {
      overflow: unset !important;
  }
  &-title {
      margin-bottom: 10px;
      color: #313238;
  }
  &-content {
      &.basic-info {
          border: 1px solid #dfe0e5;
          border-radius: 2px;
          .basic-info-item {
              display: flex;
              align-items: center;
              height: 32px;
              padding: 0 15px;
              &:nth-of-type(even) {
                  background: #F7F8FA;
              }
              label {
                  line-height: 32px;
                  border-right: 1px solid #dfe0e5;
                  width: 200px;
              }
              span {
                  padding: 0 15px;
                  flex: 1;
                  overflow: hidden;
                  text-overflow: ellipsis;
                  white-space: normal;
                  word-break: break-all;
                  display: -webkit-box;
                  -webkit-line-clamp: 1;
                  -webkit-box-orient: vertical;
              }
          }
      }
  }
}

</style>
