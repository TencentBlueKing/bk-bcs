/**
 * @file 申请权限的 mixin
 */

// 功能、角色对应 map
const POLICY_ROLE_MAP = {
    // 创建 -> 创建者角色
    create: 'creator',
    // 删除 -> 拥有者角色
    delete: 'bcs_manager',
    // 列表 -> 拥有者角色
    list: 'bcs_manager',
    // 查看 -> 拥有者角色
    view: 'bcs_manager',
    // 编辑 -> 拥有者角色
    edit: 'bcs_manager',
    // 使用 -> 拥有者角色
    use: 'bcs_manager'
}

export default {
    methods: {
        createApplyPermUrl ({ policy, projectCode, idx }) {
            const url = `${DEVOPS_BCS_API_URL}/api/perm/apply/subsystem/?client_id=bcs-web-backend&service_code=bcs`
                + `&project_code=${projectCode}&role_${POLICY_ROLE_MAP[policy]}=${idx}`
            return url
        }
    }
}
