import { defineComponent, PropType, computed, toRefs } from '@vue/composition-api'
import './status-icon.css'

export type StatusType = 'running' | 'completed' | 'failed' | 'terminating' | 'pending' | 'unknown' | 'notready'

export default defineComponent({
    name: 'status',
    props: {
        status: {
            type: String as PropType<StatusType>,
            default: ''
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
                false: 'red'
            })
        }
    },
    setup (props) {
        const { statusColorMap, status } = toRefs(props)
        const statusClass = computed(() => {
            return `status-icon status-${statusColorMap.value[status.value.toLowerCase()]}`
        })
        return {
            statusClass
        }
    },
    render () {
        return (
            <div class="dashboard-status">
                <span class={this.statusClass}></span>
                {
                    this.$scopedSlots.default
                        ? this.$scopedSlots.default(status)
                        : <span class="status-name bcs-ellipsis">{this.status}</span>
                }
            </div>
        )
    }
})
