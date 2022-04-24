import { Route } from 'vue-router'
import { userPerms } from '@/api/base'
export default {
    data () {
        return {
            $BcsAuthMap: {}
        }
    },
    beforeRouteEnter (to: Route, from: Route, next) {
        next(vm => {
            const actionIds = to.meta?.authority?.actionIds || []
            if (actionIds.length) {
                setTimeout(async () => {
                    const params = vm.$BcsAuthParams || {}
                    const data = await userPerms({
                        action_ids: actionIds,
                        ...params
                    })
                    vm.$BcsAuthMap = data?.perms || {}
                }, 0)
            }
        })
    }
}
