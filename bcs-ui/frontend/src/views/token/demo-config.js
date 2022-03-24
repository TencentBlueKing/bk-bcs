export default {
    apiVersion: 'v1',
    kind: 'Config',
    clusters: [
        {
            cluster: {
                server: '${bcs_api_host}/clusters/${cluster_id}/'
            },
            name: '${cluster_id}'
        }
    ],
    contexts: [
        {
            context: {
                cluster: '${cluster_id}',
                user: '${username}'
            },
            name: 'BCS'
        }
    ],
    'current-context': 'BCS',
    users: [
        {
            name: '${username}',
            user: {
                token: '${token}'
            }
        }
    ]
}
