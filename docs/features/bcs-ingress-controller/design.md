# bcs-ingress-controller design

## features

* 支持HTTPS，HTTP，TCP，UDP协议
* 支持clb多种详细参数设置
* 支持单个ingress控制多个clb实例
* 支持转发到NodePort模式和转发到直通Pod模式
* 支持单端口多Service流量转发，WRR负载均衡方法下权重配比
* 直通Pod模式下，支持Service内部通过Label选择Pod，WRR负载均衡方法下权重配比
* 支持StatefulSet和GameStatefulSet端口段映射（借助clb端口段特性，实现规则合并）
* 良好的观测性（Metrics和Event机制）
* 云接口的限流与重试
* 合并短期事件，防止抖动

## architecture

![bcs-ingress-controller](./img/bcs-ingress-controller-architecture.png)

### worker cache

![bcs-ingress-controller-worker-cache](./img/bcs-ingress-controller-worker-cache.png)
