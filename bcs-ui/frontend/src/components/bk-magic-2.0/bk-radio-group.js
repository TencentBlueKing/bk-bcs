export default {
    name: 'bk-radio-group',
    props: {
        value: {
            type: [String, Number, Boolean],
            default: ''
        }
    },
    watch: {
        value () {
            this.$nextTick(() => {
                this.updateValue(this.value)
            })
        }
    },
    methods: {
        change (data) {
            this.$refs.bcsRadioGroup.change(data)
            this.updateValue(data.value)
        },
        updateValue (value) {
            const radios = []
            for (const node of this.$refs.bcsRadioGroup.radios) {
                radios.push(...this.$refs.bcsRadioGroup.findComponentsDownward(node, 'bcs-radio'), ...this.$refs.bcsRadioGroup.findComponentsDownward(node, 'bcs-radio-button'))
            }
            radios.forEach(child => {
                child.current = value
            })
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
        return h('bcs-radio-group', {
            ref: 'bcsRadioGroup',
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
