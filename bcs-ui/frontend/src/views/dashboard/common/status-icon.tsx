import { defineComponent, computed, toRefs } from '@vue/composition-api';
import LoadingIcon from '@/components/loading-icon.vue';
import './status-icon.css';

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
        terminating: 'blue',
        true: 'green',
        false: 'red',
        unknown: 'gray',
      }),
    },
  },
  setup(props) {
    const { statusColorMap, statusTextMap, status } = toRefs(props);
    const color = computed(() => statusColorMap.value[status.value.toLowerCase()]
      || statusColorMap.value[status.value]);
    const statusClass = computed(() => `status-icon status-${color.value}`);
    const statusText = computed(() => statusTextMap.value[status.value] || status.value || '--');

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
                  : <span class="status-name bcs-ellipsis">{this.statusText}</span>
            }
        </div>
      );
  },
});
