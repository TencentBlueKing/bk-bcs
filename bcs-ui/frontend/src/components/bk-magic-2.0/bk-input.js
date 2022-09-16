export default {
    name: 'bk-input',
    props: {
        value: {
            type: [String, Number],
            default: ''
        },
        parseNumber: {
            type: Boolean,
            default: true
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })

        const _self = this
        try {
            this.$listeners.input = (value, event) => {
                if (_self.$attrs.type === 'number') {
                    _self.$emit('input', _self.parseNumber ? Number(value) : String(value), event)
                } else {
                    _self.$emit('input', value, event)
                }
            }
        } catch (e) {
            console.warn(e)
        }
        
        return h('bcs-input', {
            ref: 'input',
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    },
    methods: {
        focus () {
            this.$refs.input.focus()
        }
    }
}
