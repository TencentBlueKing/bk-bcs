<template>
  <div class="bcs-sideslider-content" v-bkloading="{ isLoading }">
    <DetailItem :label="$t('名称')">{{ chart ? chart.name : '--' }}</DetailItem>
    <DetailItem :label="$t('简介')">{{ chart ? chart.latestDescription : '--' }}</DetailItem>
    <DetailItem :label="$t('版本')">
      <bcs-select :clearable="false" v-model="curVersion">
        <bcs-option
          v-for="item in versions"
          :key="item.version"
          :name="item.version"
          :id="item.version">
        </bcs-option>
      </bcs-select>
    </DetailItem>
    <bcs-tab class="mt-[16px]">
      <bcs-tab-panel :label="$t('资源文件')" name="1">
        <ChartFileTree class="h-full" :contents="detail.contents"></ChartFileTree>
      </bcs-tab-panel>
      <bcs-tab-panel :label="$t('详细说明')" name="2">
        <BcsMd :code="mdCode" v-if="mdCode"></BcsMd>
        <bcs-exception type="empty" scene="part" v-else></bcs-exception>
      </bcs-tab-panel>
    </bcs-tab>
  </div>
</template>
<script lang="ts">
import {  computed, defineComponent, ref, toRefs, watch } from 'vue';
import DetailItem from '@/components/layout/DetailItem.vue';
import ChartFileTree from './chart-file-tree.vue';
import BcsMd from '@/components/bcs-md/index.vue';
import useHelm from './use-helm';

export default defineComponent({
  name: 'ChartDetail',
  components: { DetailItem, ChartFileTree, BcsMd },
  props: {
    repoName: {
      type: String,
      default: '',
      required: true,
    },
    chart: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props) {
    const { repoName, chart } = toRefs(props);
    const { handleGetRepoChartVersionDetail, handleGetRepoChartVersions } = useHelm();
    const detail = ref<Record<string, any>>({});
    const curVersion = ref('');
    const versions = ref<any[]>([]);
    const isLoading = ref(false);
    const treeRef = ref<any>(null);

    const mdCode = computed(() => {
      const contents = detail.value?.contents || {};
      const key = Object.keys(contents).find(key => contents[key]?.name === 'README.md');
      return key ? contents[key]?.content : '';
    });
    watch([repoName, chart], async () => {
      const { name, latestVersion } = chart.value;
      if (name && latestVersion) {
        isLoading.value = true;
        const [versionsData, detailData] = await Promise.all([
          handleGetRepoChartVersions(repoName.value, name),
          handleGetRepoChartVersionDetail(repoName.value, name, latestVersion),
        ]);
        curVersion.value = latestVersion;
        versions.value = versionsData?.data;
        detail.value = detailData;
        isLoading.value = false;
      }
    }, { immediate: true, deep: true });
    watch(curVersion, async () => {
      isLoading.value = true;
      detail.value = await handleGetRepoChartVersionDetail(repoName.value, chart.value.name, curVersion.value);
      isLoading.value = false;
    });

    return {
      mdCode,
      isLoading,
      curVersion,
      detail,
      versions,
      treeRef,
    };
  },
});
</script>
<style lang="postcss" scoped>
>>> .bk-tab-content {
  height: calc(100vh - 310px);
}
</style>
