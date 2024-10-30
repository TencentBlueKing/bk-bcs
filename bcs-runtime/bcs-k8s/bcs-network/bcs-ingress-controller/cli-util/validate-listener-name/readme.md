cli工具， 功能包括

1. 检查集群内所有Listener资源的状态是否正常
2. 检查集群内所有Listener资源关联的云上监听器名称是否符合ListenerValidateMode对应的规范，ListenerValidateMod包括：
    + CLOSE: 不校验云上监听器名称
    + NORMAL: 云上监听器名称为 [lbID-protocol-port] | [lbID-port]
    + STRICT: 云上监听器名为[BCSClusterID-lbID-protocol-port] | [BCSClusterID-lbID-port]
3. 根据ListenerValidateMode， 更新集群内所有监听器关联的云上监听器名称