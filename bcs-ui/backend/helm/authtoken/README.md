# 模块说明
> 本模块提供统一登录之外的一种认证功能，用于解决纯后台应用无用户access_token 的情况调用helm模块的功能

## 设计背景
- 需求来源：结合 gitlab CI/CD 将运营系统部署到k8s集群
- 适用场景：无 access_token 校验的场景

## 实现难点
- 直接将单个用户绑定到一个token的方式，风险非常大，一旦 token 泄漏了，用户的身份很可能出现被盗用的风险
- 一个 token 用于多个场景，定期维护 token 非常困难。比如 token 用于 ABC 三个场景，某一天用户感觉A场景可能泄漏了token，希望更新token值，这时候用户就需要将ABC三个场景的token都更换。

## 方案
- 一个用户可以拥有多个 token，每个token都独立拥有用户的一部分权限。
- token 的创建和维护，基于已有的统一登录和paas-auth鉴权体系

### 通过该方案实现的认证/鉴权模型 - 场景一
- 场景A需要用户身份执行创建 Helm 应用的权限，用户可以以自己身份申请一个用于场景A的 token, 做创建 Helm应用 的操作。
- 场景B只希望监控 Helm应用 的状态，用户可以再以自己的身份申请一个用于场景B的 token，做查询 Helm应用 状态的事情。
- 场景A需要更新 token，不影响场景B

> note: 这种场景不局限于一种大的功能块，也包括一个权限范围，比如A系统需要有在集群1下创建Helm应用的能力，B系统需要有在集群2下创建Helm应用的能力，这时候就可以单独分配 token


## 实现
- 参考 [扩展drf支持一个用户多个token](https://consideratecode.com/2016/10/06/multiple-authentication-tokens-per-user-with-django-rest-framework/)


## TODO
- 基于 paas-auth 请求时鉴权(因为没有access_token，需要paas-auth对paas-backend放开查询权限)
- 基于 paas-auth 的通用 provider(nice to have)

