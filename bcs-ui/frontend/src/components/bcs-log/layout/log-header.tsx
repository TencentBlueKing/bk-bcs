import { defineComponent, PropType, ref, toRefs, watch, computed } from 'vue';
import { TranslateResult } from 'vue-i18n';
import $i18n from '@/i18n/i18n-setup';

interface IOption {
  id: string | number;
  name: string;
}

interface IOperate {
  event: string;
  tips: TranslateResult;
  icon: string;
}

export default defineComponent({
  name: 'LogHeader',
  props: {
    // 标题
    title: {
      type: String,
      default: '',
    },
    // 当前容器
    defaultContainer: {
      type: String,
      default: '',
    },
    // 容器列表
    containerList: {
      type: Array as PropType<IOption[]>,
      default: (): IOption[] => [],
    },
    // 是否显示时间戳
    defaultTimeStamp: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const showTimeStamp = ref(props.defaultTimeStamp);
    const container = ref(props.defaultContainer);
    const realTimeLog = ref(false);

    const { defaultContainer } = toRefs(props);
    watch(defaultContainer, (v) => {
      container.value = v;
    });

    const operateList = computed<IOperate[]>(() => [
      {
        event: 'time-stamp-change',
        tips: showTimeStamp.value ? $i18n.t('隐藏时间') : $i18n.t('显示时间'),
        icon: 'bcs-icon bcs-icon-clock',
      },
      {
        event: 'refresh',
        tips: $i18n.t('刷新'),
        icon: 'bcs-icon bcs-icon-refresh',
      },
      {
        event: 'real-time',
        tips: realTimeLog.value ? $i18n.t('关闭实时日志') : $i18n.t('开启实时日志'),
        icon: realTimeLog.value ? 'bcs-icon bcs-icon-pause' : 'bcs-icon bcs-icon-play2',
      },
      {
        event: 'download',
        tips: $i18n.t('下载'),
        icon: 'bcs-icon bcs-icon-download',
      },
    ]);

    const handleOperate = (item: IOperate) => {
      if (item.event === 'time-stamp-change') {
        showTimeStamp.value = !showTimeStamp.value;
        ctx.emit(item.event, showTimeStamp.value);
      } else if (item.event === 'real-time') {
        realTimeLog.value = !realTimeLog.value;
        ctx.emit(item.event, realTimeLog.value);
      } else {
        ctx.emit(item.event);
      }
    };

    const handleContainerChange = (newValue: string, oldValue: string) => {
      ctx.emit('container-change', newValue, oldValue);
    };

    return {
      showTimeStamp,
      container,
      operateList,
      handleContainerChange,
      handleOperate,
    };
  },
  render() {
    return (
            <header class="log-header">
                <div class="log-header-left">
                    <div class="title">{ this.title }</div>
                </div>
                <div class="log-header-right">
                    <bcs-select
                        class="select"
                        clearable={false}
                        disabled={this.disabled}
                        v-model={this.container}
                        onChange={this.handleContainerChange}>
                        {
                            this.containerList.map(option => (
                                <bcs-option id={option.name} name={option.name}></bcs-option>
                            ))
                        }
                    </bcs-select>
                    {
                        this.operateList.map(item => (
                            <span class={['icon ml20', this.disabled ? 'disabled' : '']}
                                v-bk-tooltips={{ content: item.tips }}
                                onClick={() => {
                                  !this.disabled && this.handleOperate(item);
                                }}>
                                <i class={item.icon}></i>
                            </span>
                        ))
                    }
                </div>
            </header>
    );
  },
});
