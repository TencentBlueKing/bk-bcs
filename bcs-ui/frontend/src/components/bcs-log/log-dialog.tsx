import { defineComponent } from '@vue/composition-api'
import './style/log-dialog.scss'

export default defineComponent({
    name: 'LogDialog',
    model: {
        prop: 'value',
        event: 'change'
    },
    props: {
        value: {
            type: Boolean,
            default: false
        }
    },
    setup (props, ctx) {
        const handleCloseDialog = () => {
            ctx.emit('change', false)
        }
        return {
            handleCloseDialog
        }
    },
    render () {
        return (
            <div class="log-wrapper" onClick={this.handleCloseDialog} v-show={this.value}>
                { this.$scopedSlots.default && this.$scopedSlots.default({}) }
            </div>
        )
    }
})
