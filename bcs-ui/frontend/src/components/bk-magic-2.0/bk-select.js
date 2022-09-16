// 有点问题，在bk-select下使用bk-option时构建会一直pending，请使用bcs-select + bcs-option
export default {
    name: 'bk-select',
    props: {
        value: {
            type: [String, Number, Array],
            default: ''
        },
        selected: {
            type: [Number, Array, String],
            required: true
        },
        settingKey: {
            type: String,
            default: 'id'
        },
        multiSelect: {
            type: Boolean,
            default: false
        },
        allowClear: {
            type: Boolean,
            default: false
        },
        isLoading: {
            type: Boolean,
            default: false
        },
        list: {
            type: Array,
            required: true
        }
    },

    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        // 兼容1.0版本属性:
        // selected -> v-model
        // setting-key -> id-key
        // is-loading -> loading
        // allow-clear -> clearable
        // multi-select -> multiple
        // item-selected -> change
        const that = this
        this.$props.value = this.$props.selected
        this.$props.idKey = this.$props.settingKey
        this.$props.multiple = this.$props.multiSelect
        this.$props.clearable = this.$props.allowClear
        this.$props.loading = this.$props.isLoading

        // change事件兼容
        this.$listeners.change = function (oldVal, newVal) {
            // const matchData = that.list.find(item => {
            //     return item[that.$props.idKey] === newVal
            // })
            that.$emit('change', oldVal, newVal)
            // that.$emit('item-selected', newVal, matchData)
        }

        // 针对1.0版本直接启用虚拟滚动模式来兼容
        if (this.$props.list && this.$props.list.length) {
            this.$props.enableVirtualScroll = true
            this.$props.idKey = this.$props.settingKey
        }

        return h('bcs-select', {
            on: this.$listeners, // 透传事件
            props: this.$props, // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
