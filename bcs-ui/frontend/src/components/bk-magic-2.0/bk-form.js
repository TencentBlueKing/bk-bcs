export default {
    name: 'bk-form',
    props: {

    },
    data () {
        return {
            formItems: []
        }
    },
    methods: {
        validate () {
            return this.$refs.bcsForm.validate()
        },
        clearError () {
            return this.$refs.bcsForm.clearError()
        }
    },
    render (h) {
        const slots = Object.keys(this.$slots)
            .reduce((arr, key) => arr.concat(this.$slots[key]), [])
            .map(vnode => {
                vnode.context = this._self
                return vnode
            })
            
        return h('bcs-form', {
            ref: 'bcsForm',
            on: this.$listeners, // 透传事件
            props: Object.assign({}, this.$props, this.$attrs), // 透传props
            scopedSlots: this.$scopedSlots, // 透传scopedSlots
            attrs: this.$attrs // 透传属性，非props
        }, slots)
    }
}
