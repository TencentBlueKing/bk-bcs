# BCS容器服务前端代码

### 依赖

- vscode
- node > 14 （ps:推荐用20）
- volar、tailwindcss、eslint插件
- 浏览器插件 Gimli Tailwind

### 运行

建议删除之前 `node_modules` 和 `package-lock.json` 重新安装

1. 安装依赖（node版本要大于 **14**）

```shell
npm i
```

2. 在当前目录下新建`.bk.development.env`文件，其中主要配置有三个，其余配置根据情况填：

  - BK_PROXY_DEVOPS_BCS_API_URL （后端服务地址）
  - BK_BCS_API_HOST （下沉服务地址）
  - BK_LOCAL_HOST （本地HOST地址）项目启动之前看看是否已配置本地host地址，教程（https://blog.csdn.net/cc1949/article/details/78411865）

```text
# .env.development 开发模式生效

# API地址
# CVM容量接口
BK_BKSRE_HOST = ''
# 后端服务地址，仅供代理时使用（正式环境时读取 BK_DEVOPS_BCS_API_URL 变量）
BK_PROXY_DEVOPS_BCS_API_URL = 'xxx.com'
# golang 下沉服务地址
BK_BCS_API_HOST = 'xxx.com'
# 本地HOST地址
BK_LOCAL_HOST = 'xxx.com'

# python API地址（由于之前接口写死了绝对路径，这个变量需要设置为空，通过BK_PROXY_DEVOPS_BCS_API_URL变量走代理）
# BK_DEVOPS_BCS_API_URL = ''

# 登录地址 (关联: 退出登录)
BK_LOGIN_FULL = ''

# 蓝盾地址 (关联: 跳转蓝盾项目管理)
BK_DEVOPS_HOST = ''

# 静态资源地址
BK_STATIC_URL = ''

# 当前后端环境
BK_RUN_ENV = 'dev'

# 镜像地址
BK_DEVOPS_ARTIFACTORY_HOST = ''

# 当前版本
BK_REGION = 'ieod'

# 路由前缀
BK_SITE_URL = '/'

# 权限中心地址
BK_IAM_HOST = ''

# PaaS平台地址
BK_PAAS_HOST = ''

# 监控地址
BK_BKMONITOR_HOST = ''

# 允许加载的域名
BK_PREFERRED_DOMAINS = ''

```

3. 开发

```shell
npm run dev
```

4. 构建

```shell
npm run build
```

### 重构计划

#### CSS

CSS文件目前禁止添加新样式，准备通过[tailwindui](https://tailwindui.com/documentation)来统一维护

#### BCS前端公共组件（待整理）

放在`Components`目录之前，请确定这个组件至少被 2 个以上不同功能点的界面用到，不要把界面上抽取的组件放在这里，且每个组件需要测试用例。写界面之前确定当前目录下是否有可以复用的，不要重复造轮子

- key-value组件（节点列表、日志采集、CA、节点模板、命名空间设置变量、集群设置变量、metric管理）
- 污点组件
- 展示Key-Value组件 Tag
- 代码编辑器（带diff模式）
- MD组件
- 集群选择器
- **布局组件**（上下布局、行布局、列布局），写界面之前想去布局组件找
- HOOK（通用USE，eg：分页、table联动等、app相关全局信息）

#### 路由

- 尽量不要使用嵌套路由
- 路由尽量跟后端API风格保持一致，尽可能清晰
- 界面input类型参数组件需要支持路由参数

#### View

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

注意事项

- 禁止使用mixins
- CSS全部使用 `tailwindcss` 来写
- 按钮组件bk-button在有icon的时候需要注意:
  1. <bk-button icon="xx">{{xxx}}</bk-button> :不换行icon和文字间无间隙
  2. <bk-button icon="xx">
      {{xxx}}</bk-button>  ：换行后，icon和文字间有间隙，和magicbox展示的带icon的button的效果是一样的
### TODO

- 代码编辑器逐步替换为 `monaco-editor/new-editor`，以前`monaco-editor/editor.vue`、`ace-editor`和`k8s-create/yaml-mode`不要再使用
- 删除模块Vuex，只保留全局vuex
- 删除images下无用文件、把SVG转换成icon库维护
- 一个API模块最好对应一个 use 操作

### FAQ

- 如何新增全局变量
  1. 在 `index.html` 下配置变量，环境变量必须以`BK_`开头，eg: `process.env.BK_xxx`
  2. 在 `.bk.production.env` 配置后端渲染变量模板，eg: BK_xxx = '{{ BK_xxx }}'
  3. 在 `.bk.development.env` 配置本地调试的变量，eg: BK_xxx = 'test'

- Vue 单文件 Eslint 检测不正常
  1. 检查是否安装最新 `Volar` 插件
  2. 检查 Vscode 是否配置 eslint 自动保存配置
  3. 删除 nodemodules 重新安装包，然后重启

- 路由规范