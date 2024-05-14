/* eslint-disable no-unused-expressions */
import { defineComponent, onBeforeUnmount, reactive, ref, toRefs, watch } from 'vue';

import LogContent, { ILogData } from './layout/log-content';
import LogHeader from './layout/log-header';

import './style/log.css';
import $store from '@/store';

interface IState {
  log: ILogData[]; // 日志内容
  showTimeStamp: boolean; // 是否显示时间戳
  container: string; // 当前容器
  loading: boolean;
  contentLoading: boolean;
  previous: string;
  step: number;
  showLastContainer: boolean;
}

export default defineComponent({
  name: 'BcsLog',
  components: {
    LogHeader,
    LogContent,
  },
  props: {
    height: {
      type: [Number, String],
      default: '100%',
    },
    defaultContainer: {
      type: String,
      default: '',
    },
    podId: {
      type: String,
      default: '',
    },
    projectId: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
    namespaceId: {
      type: String,
      default: '',
    },
    containerList: {
      type: Array,
      default: () => [],
    },
    globalLoading: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    const contentRef = ref<any>(null);
    const state = reactive<IState>({
      log: [], // 日志内容
      showTimeStamp: false, // 是否显示时间戳
      container: props.defaultContainer, // 当前容器
      loading: false,
      contentLoading: false,
      previous: '',
      step: 4,
      showLastContainer: false,
    });

    const getParams = () => ({
      $projectId: props.projectId,
      $clusterId: props.clusterId,
      $namespaceId: props.namespaceId,
      $podId: props.podId,
      container_name: state.container,
      previous: state.showLastContainer,
    });
    const handleGetLog = async () => {
      if (!state.container || state.contentLoading) return;

      state.contentLoading = true;
      const data = await $store.dispatch('log/podLogs', getParams());
      state.log = data.logs;
      state.previous = data.previous;
      state.contentLoading = false;
      contentRef.value?.scrollIntoIndex();
    };

    const { defaultContainer } = toRefs(props);

    watch(defaultContainer, async (container) => {
      state.container = container;
      handleGetLog();
    }, { immediate: true });

    const initVirtualStatus = () => {
      contentRef.value?.initStatus();
    };

    const handleLoadMoreLog = async () => {
      // 滚动到顶部时加载更多日志
      if (!state.previous || state.loading) return;

      state.loading = true;
      const data = await $store.dispatch('log/previousLogList', state.previous);
      state.log.splice(0, 0, ...data.logs);
      setTimeout(() => {
        contentRef.value?.scrollIntoIndex(data.logs.length - state.step);
        state.previous = data.previous;
        state.loading = false;
      }, 0);
    };

    const handleRefresh = () => {
      // 刷新日志
      handleGetLog();
    };

    const handleDownload = () => {
      const { $clusterId, $namespaceId, $podId, container_name } = getParams();
      $store.dispatch('log/downloadLogs', {
        $clusterId,
        $namespaceId,
        $podId,
        $containerName: container_name,
        $previous: state.showLastContainer,
      });
    };

    const handleTimeStampChange = (show: boolean) => {
      // 是否显示时间戳
      state.showTimeStamp = show;
    };
    let logSSR: EventSource | null = null;
    // 实时日志功能
    const handleRealTimeLog = async (realTime: boolean) => {
      if (realTime) {
        state.contentLoading = true;
        logSSR = await $store.dispatch('log/realTimeLogStream', {
          $clusterId: props.clusterId,
          $namespaceId: props.namespaceId,
          $podId: props.podId,
          $containerName: state.container,
          $startedAt: encodeURIComponent(String(state.log[state.log.length - 1]?.time || '')),
        });
        state.contentLoading = false;
        logSSR?.addEventListener('open', () => {
          console.log('open ssr');
        });
        logSSR?.addEventListener('message', (event: MessageEvent) => {
          try {
            const data: ILogData[] = JSON.parse(event.data);
            setTimeout(() => {
              state.log.push(...data);
            }, 0);
            contentRef.value?.scrollIntoIndex();
            const ids = data.map(item => item.time) ;
            contentRef.value?.setHoverIds(ids);
            // 关闭hover效果
            setTimeout(() => {
              contentRef.value?.removeHoverIds(ids);
            }, 1000);
          } catch (err) {
            console.log(err);
          }
        });
      } else {
        logSSR?.close();
      }
    };

    const handleContainerChange = (newValue: string) => {
      // 当前容器变更
      state.container = newValue;
      handleGetLog();
    };

    const handleToggleLast = async (v: boolean) => {
      state.showLastContainer = v;
      handleGetLog();
    };

    onBeforeUnmount(() => {
      logSSR?.close();
    });

    return {
      ...toRefs(state),
      contentRef,
      handleLoadMoreLog,
      handleRefresh,
      handleDownload,
      handleTimeStampChange,
      initVirtualStatus,
      handleContainerChange,
      handleToggleLast,
      handleRealTimeLog,
      getParams,
    };
  },
  render() {
    return (
            <article class="log-container" v-bkloading={{ isLoading: this.globalLoading, opacity: 0.1 }}>
                <log-header
                    title={this.podId}
                    containerList={this.containerList}
                    defaultTimeStamp={this.showTimeStamp}
                    defaultContainer={this.container}
                    disabled={this.loading}
                    on-time-stamp-change={this.handleTimeStampChange}
                    on-refresh={this.handleRefresh}
                    on-download={this.handleDownload}
                    on-container-change={this.handleContainerChange}
                    on-real-time={this.handleRealTimeLog}
                />
                <log-content
                    ref="contentRef"
                    height={this.height}
                    showTimeStamp={this.showTimeStamp}
                    log={this.log}
                    loading={this.loading}
                    v-bkloading={{ isLoading: this.contentLoading, opacity: 0.1 }}
                    on-scroll-top={this.handleLoadMoreLog}
                    on-toggle-last={this.handleToggleLast}
                />
            </article>
    );
  },
});
