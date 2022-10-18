# BCS容器服务前端代码

### 运行

建议删除之前 `node_modules` 和 `package-lock.json` 重新安装

1. 安装依赖（node版本要大于 **14**）

```shell
npm i
```

2. 在当前目录下新建`.bk.development.env`文件，其中主要配置有三个，其余配置根据情况填：

  - BK_PROXY_DEVOPS_BCS_API_URL （后端服务地址）
  - BK_BCS_API_HOST （下沉服务地址）
  - BK_LOCAL_HOST （本地HOST地址）

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
BK_IAM_APP_URL = ''

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

### FAQ

- 如何新增全局变量
  1. 在 `index.html` 下配置变量，环境变量必须以`BK_`开头，eg: `process.env.BK_xxx`
  2. 在 `.bk.production.env` 配置后端渲染变量模板，eg: BK_xxx = '{{ BK_xxx }}'
  3. 在 `.bk.development.env` 配置本地调试的变量，eg: BK_xxx = 'test'