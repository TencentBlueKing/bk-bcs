import { computed, defineComponent, toRefs } from 'vue';

import './status-icon.css';
import LoadingIcon from '@/components/loading-icon.vue';
import $i18n from '@/i18n/i18n-setup';

export default defineComponent({
  name: 'StatusIcon',
  components: { LoadingIcon },
  props: {
    pending: {
      type: Boolean,
      default: false,
    },
    status: {
      type: String,
      default: '',
    },
    statusTextMap: {
      type: Object,
      default: () => ({}),
    },
    // 每种状态对应的颜色, 默认黄色
    statusColorMap: {
      type: Object,
      default: () => ({
        running: 'green',
        completed: 'green',
        failed: 'red',
        FAILURE: 'red',
        terminating: 'blue',
        true: 'green',
        false: 'red',
        unknown: 'gray',
      }),
    },
    type: {
      type: String,
      default: 'persistence', // persistence 或 result
    },
    hideText: {
      type: Boolean,
      default: false,
    },
    message: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { statusColorMap, statusTextMap, status, type, hideText } = toRefs(props);
    const color = computed(() => statusColorMap.value[status.value.toLowerCase()]
      || statusColorMap.value[status.value]);
    const statusClass = computed(() => (type.value === 'persistence'
      ? `status-icon status-${color.value}`
      : `status-icon-result status-${color.value}-result`));
    const statusText = computed(() => {
      if (hideText.value) return '';
      return statusTextMap.value[status.value] || status.value || $i18n.t('generic.status.unknown1');
    });

    return {
      statusClass,
      statusText,
    };
  },
  render() {
    return this.pending
      ? <LoadingIcon
        {...{
          scopedSlots: {
            default: () => (this.$scopedSlots.default
              ? this.$scopedSlots.default(this.status)
              : <span class="status-name bcs-ellipsis">{this.statusText}</span>),
          },
        }} />
      : (
        <div class="dashboard-status">
            <span class={this.statusClass}></span>
            {
                this.$scopedSlots.default
                  ? this.$scopedSlots.default(this.status)
                  : <span
                      class={['status-name bcs-ellipsis', this.message ? 'bcs-border-tips !flex-none' : '']}
                      v-bk-tooltips={{ content: this.message, disabled: !this.message }}>
                      {this.statusText}
                    </span>
            }
        </div>
      );
  },
});
