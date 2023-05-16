# GameDeployment

> GameDeployment 是针对游戏 gameserver 实现的管理无状态应用的增强版 deployment，基于 k8s 原生 replicaset 改造，并参考了 openkruise 的部分实现，支持原地重启、镜像热更新、滚动更新、灰度发布等多种更新策略。

## 参考资料

1. [GameDeployment 介绍](https://github.com/Tencent/bk-bcs/tree/master/docs/features/bcs-gamedeployment-operator)
2. [GameDeployment 使用示例](https://github.com/Tencent/bk-bcs/tree/master/docs/features/bcs-gamedeployment-operator/example)
3. [金丝雀发布使用示例](https://github.com/Tencent/bk-bcs/tree/master/docs/features/bcs-gamedeployment-operator/features/canary)
4. [preDeleteHook 使用示例](https://github.com/Tencent/bk-bcs/tree/master/docs/features/bcs-gamedeployment-operator/features/preDeleteHook)
