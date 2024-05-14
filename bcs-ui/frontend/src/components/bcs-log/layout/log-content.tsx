/* eslint-disable no-unused-expressions */
import { debounce } from 'lodash';
import { computed, defineComponent, onMounted, PropType, ref, watch } from 'vue';

import AnsiParser from '../common/ansi-parser';
import TransformStringPixel from '../common/transform-string-pixel';

import '../style/log.css';
import { formatTime } from '@/common/util';
import $i18n from '@/i18n/i18n-setup';

export interface ILogData {
  log: string;
  time: number;
}

interface IInnerLogData extends ILogData {
  html: string;
  breakLine: boolean;
}

export default defineComponent({
  name: 'LogContent',
  props: {
    // 日志数据
    log: {
      type: Array as PropType<ILogData[]>,
      default: (): ILogData[] => [],
    },
    // 是否显示时间戳
    showTimeStamp: {
      type: Boolean,
      default: false,
    },
    // 日志内容高度
    height: {
      type: [Number, String],
      default: 600,
    },
    loading: {
      type: Boolean,
      default: true,
    },
  },
  setup(props, ctx) {
    const contentWidth = ref(0);
    const virtualScrollRef = ref<any>(null);
    const hoverIds = ref<any[]>([]);
    const showLastContainer = ref(false);

    const ansiParser = new AnsiParser();
    const transformStringPixel = new TransformStringPixel();

    const scrollIntoIndex = (index?: number) => {
      setTimeout(() => {
        // 滚动到末尾
        virtualScrollRef.value?.scrollPageByIndex(!index ? logData.value.length : index);
      }, 0);
    };

    const style = computed({
      get() {
        return {
          height: typeof props.height === 'string' ? props.height : `${props.height}px`,
        };
      },
      set() {},
    });
    // 处理后端日志
    const logData = computed({
      get() {
        if (!contentWidth.value) return [];

        const data: IInnerLogData[] = [];
        props.log.forEach((item) => {
          const { html, plainText, firstForegroundColor } = ansiParser.transformToHtml(item.log);
          const pixel = transformStringPixel.getStringPixel(plainText);
          // 150： 展示时间戳的宽度
          const width = props.showTimeStamp ? contentWidth.value - 150 : contentWidth.value;
          if (pixel > width) {
            // 当前行大于容器宽度，自动换行处理
            const charPixel = pixel / plainText.length; // 每个字符平均像素（不准确）
            const splitLen = Math.floor(width / charPixel);
            handleSplitMessage(plainText, splitLen).forEach((str, index) => {
              data.push({
                ...item,
                html: `<span style="color: ${firstForegroundColor || '#fff'}">${str}</span>`,
                breakLine: index !== 0,
              });
            });
          } else {
            data.push({
              ...item,
              html,
              breakLine: false, // 是否是被拆分的行（拆分的行不显示时间戳）
            });
          }
        });
        return data;
      },
      set() {},
    });

    watch(contentWidth, () => {
      scrollIntoIndex();
    });

    onMounted(() => {
      setTimeout(() => {
        contentWidth.value = virtualScrollRef.value?.$el.clientWidth;
      }, 0);
    });

    const initStatus = () => {
      virtualScrollRef.value?.initStatus();
    };

    // 分割长日志
    const handleSplitMessage = (str: string, n = 0) => {
      const arr: string[] = [];
      for (let i = 0, len = str.length; i < len / n; i++) {
        arr.push(str.slice(n * i, n * (i + 1)));
      }
      return arr;
    };
    // 纵向滚动
    const handleScrollChange = debounce(() => {
      // 置顶
      if (virtualScrollRef.value?.$refs.scrollMain?.style.top === '0px') {
        ctx.emit('scroll-top');
      }
    }, 200);
    // 悬浮
    const handleMouseEnter = (data: IInnerLogData) => {
      const index = hoverIds.value.findIndex(time => data.time === time);
      index === -1 && hoverIds.value.push(data.time);
    };

    const handleMouseleave = (data: IInnerLogData) => {
      const index = hoverIds.value.findIndex(time => data.time === time);
      index > -1 && hoverIds.value.splice(index, 1);
    };

    const lastText = computed(() => (showLastContainer.value ? $i18n.t('generic.log.button.latest') : $i18n.t('generic.log.button.previous')));

    const handleToggleLast = () => {
      showLastContainer.value = !showLastContainer.value;
      ctx.emit('toggle-last', showLastContainer.value);
    };

    const setHoverIds = (ids: any[]) => {
      ids.forEach((id) => {
        if (!hoverIds.value.includes(id)) {
          hoverIds.value.push(id);
        }
      });
    };

    const removeHoverIds = (ids: any[]) => {
      ids.forEach((id) => {
        const index = hoverIds.value.findIndex(item => item === id);
        if (index > -1) {
          hoverIds.value.splice(index, 1);
        }
      });
    };

    return {
      logData,
      contentWidth,
      style,
      hoverIds,
      virtualScrollRef,
      handleScrollChange,
      showLastContainer,
      lastText,
      handleMouseEnter,
      handleMouseleave,
      formatTime,
      initStatus,
      scrollIntoIndex,
      handleToggleLast,
      setHoverIds,
      removeHoverIds,
    };
  },
  render() {
    return (
            <div class="log-wrapper">
                <div class="log-content-pre"
                    v-show={this.loading}
                    v-bkloading={{
                      isLoading: this.loading,
                      extCls: 'pre-loading',
                      theme: 'default',
                      size: 'small',
                    }}>
                </div>
                <div class="log-content-tips" onClick={this.handleToggleLast} v-show={!this.loading}>
                    <i class={[
                      'bcs-icon mr5',
                      this.showLastContainer ? 'bcs-icon-angle-double-down' : 'bcs-icon-angle-double-up',
                    ]}></i>
                    <span>{this.lastText}</span>
                </div>
                <bcs-virtual-scroll
                    class='log-content'
                    ref="virtualScrollRef"
                    list={this.logData}
                    item-height={16}
                    style={this.style}
                    scopedSlots={
                        {
                          default: ({ data }: { data: IInnerLogData }) => (
                                <div class={['log-item', { active: this.hoverIds.includes(data.time) }]}
                                    onMouseenter={() => {
                                      this.handleMouseEnter(data);
                                    }}
                                    onMouseleave={() => this.handleMouseleave(data)}>
                                    {
                                        this.showTimeStamp && !data.breakLine
                                          ? (
                                                <span class="log-item-time mr5">
                                                    {
                                                        `[${this.formatTime(data.time, 'yyyy-MM-dd hh:mm:ss')}]`
                                                    }
                                                </span>
                                          )
                                          : null
                                    }
                                    <span class="log-item-content"
                                        style={{ 'margin-left': this.showTimeStamp && data.breakLine ? '150px' : '' }}
                                        domProps-InnerHTML={data.html}>
                                    </span>
                                </div>
                          ),
                        }
                    }
                    on-change={this.handleScrollChange}>
                </bcs-virtual-scroll>
                {/* 后置插槽 */}
                {
                    this.$scopedSlots.append?.({})
                }
            </div>
    );
  },
});
