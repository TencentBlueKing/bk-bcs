import { defineComponent, PropType, computed } from '@vue/composition-api'
import './status-icon.css'

export type StatusType = 'running' | 'completed' | 'failed' | 'terminating' | 'pending' | 'unknown' | 'notready'

export default defineComponent({
    name: 'status',
    props: {
        status: {
            type: String as PropType<StatusType>,
            default: ''
        }
    },
    setup (props) {
        // 每种状态对应的颜色, 默认黄色
        const statusMap = {
            running: 'green',
            completed: 'green',
            failed: 'red',
            terminating: 'blue',
            true: 'green',
            false: 'red'
        }
        const statusClass = computed(() => {
            return `status-icon status-${statusMap[props.status.toLowerCase()]}`
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
