## View

项目业务代码，命名建议驼峰形式，一般一个菜单对应一个目录

- app (导航、通知等跟整个UI都相关的界面)
- project (项目创建、编辑)
- variable (变量管理)
- cluster (集群管理)
- node 节点管理
- helm
- tools 组件库
- hpa
- storage
- network
- dashborad 资源视图

### 注意事项

- 禁止使用mixins
- 