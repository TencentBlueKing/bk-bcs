export default {
    data () {
        return { PROJECT_CONFIG: window.BCS_CONFIG || { doc: {}, str: {} } }
    },
    computed: {
        $INTERNAL  () {
            return !['ce', 'ee'].includes(window.REGION)
        }
    }
}
