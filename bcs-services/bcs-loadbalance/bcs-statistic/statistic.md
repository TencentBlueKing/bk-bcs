# LB metrics数据结构说明

metrics数据结构为json，包含两个主要字段：

* Frontends：用于标注服务端口监听信息
* Backends：用于标注服务监听端口所对应的后端节点

Frontends和Backends是一一对应关系。具备相同的名字，方便索引，以下是演示案例

```json
{
    "Backends": {
        "backend-http_8088": {
            "ErrorCnt": "0",
            "CurQueue": "0",
            "ResponseDeny": "0",
            "Servers": [
                {
                    "ErrorCnt": "0",
                    "CurQueue": "0",
                    "ResponseDeny": "0",
                    "Pxname": "backend-http_8088",
                    "CumSession": "0",
                    "QTime": "0",
                    "WRetry": "0",
                    "BytesIn": "0",
                    "CTime": "0",
                    "Hrsp_5xx": "",
                    "Hrsp_2xx": "",
                    "LBTot": "0",
                    "Name": "backend-127_0_0_1_8088",
                    "ChkDown": "0",
                    "Hrsp_1xx": "",
                    "CurQMax": "0",
                    "Active": "1",
                    "WRedispath": "0",
                    "Hrsp_3xx": "",
                    "Hrsp_4xx": "",
                    "RateMax": "0",
                    "Addr": "127.0.0.1:8088",
                    "Weight": "1",
                    "AgentDur": "",
                    "LastSess": "-1",
                    "SessionLimit": "",
                    "Rate": "0",
                    "ErrResp": "0",
                    "RequestError": "",
                    "Status": "UP",
                    "TTime": "0",
                    "BytesOut": "0",
                    "RTime": "0",
                    "MaxSessRate": "0",
                    "CurSessRate": "0",
                    "AgentStatus": "",
                    "DownTime": "0",
                    "RequestDeny": ""
                }
            ]
        }
    },
    "Frontends": {
        "frontend-http_8088": {
            "ResponseDeny": "0",
            "ConnRate": "0",
            "BytesOut": "0",
            "MaxSessRate": "0",
            "BytesIn": "0",
            "RateMax": "0",
            "CurSessRate": "0",
            "Rate": "0",
            "ConnTot": "0",
            "Name": "frontend-http_8088",
            "SessionLimit": "102400",
            "RequestError": "0",
            "Status": "OPEN",
            "ConnRateMax": "0",
            "RequestDeny": "0",
            "CumSession": "0"
        }
    }
}
```

## Frontends

Name：名称，对应唯一后端名称
CurSessRate：当前每秒会话数
MaxSessRate：当前每秒最大会话数
SessionLimit：当前配置最大的会话数
CumSession：累积的会话数
BytesIn：入流量
BytesOut：出流量
RequestDeny：拒绝请求数
ResponseDeny：拒绝回应数
RequestError：错误请求数
Status：当前端口状态(OPEN/UP/DOWN/NOLB/MAINT)
Rate：上一秒新建会话数
RateMax：最大一秒新建会话数
ConnRate：上一秒新建连接数
ConnRateMax：最大的每秒连接数
ConnTot：累计链接数

## Backends

ErrorCnt：后端所有节点总计错误数
CurQueue：后端所有节点链接对列总数
ResponseDeny：后端所有节点链接拒绝数

### Server

每个backend会有0到多个Server后端，以下是每个后端节点独立统计信息

serv.Name：后端节点名称
serv.Pxname：后端节点backend名称，便于索引
serv.CurQueue：当前对列中缓存的请求
serv.CurQMax：缓存对列中最大缓存请求
serv.CurSessRate：当前Session数
serv.MaxSessRate：最大的Session数
serv.SessionLimit：已配置的最大Session数
serv.CumSession：累计的Session数
serv.BytesIn：入流量
serv.BytesOut：出流量
serv.RequestDeny：该节点请求拒绝数
serv.ResponseDeny：该节点回应拒绝数
serv.RequestError：该节点请求错误数
serv.ErrorCnt：请求该节点错误数
serv.ErrResp：该节点回应错误数
serv.WRetry：该节点重试次数
serv.WRedispath：请求该节点失败，重新派发其他节点次数
serv.Status：该节点端口状态
serv.Weight：该节点权重
serv.Active：该节点是否active
serv.ChkDown：该节点down掉次数
serv.DownTime：该节点down掉总时长，单位秒
serv.LBTot：请求转发至该节点次数
serv.Rate：上一秒该节点Session次数
serv.RateMax：每秒最大新建Session数
serv.Hrsp_1xx：http返回码为1xx的次数，http时有效
serv.Hrsp_2xx：http返回码为2xx的次数，http时有效
serv.Hrsp_3xx：http返回码为3xx的次数，http时有效
serv.Hrsp_4xx：http返回码为4xx的次数，http时有效
serv.Hrsp_5xx：http返回码为5xx的次数，http时有效
serv.LastSess：上一次链接至今秒数，时间越长，意味着空闲
serv.QTime：前1024次缓存对列中请求平均等待时间，单位ms
serv.CTime：前1024次链接该节点的平均等待时间，单位ms
serv.RTime：前1024次 http请求的平均回复时间，TCP时恒为零，单位ms
serv.TTime：前1024次会话建立平均时间，单位ms
serv.AgentStatus：上次健康检查状态
serv.AgentDur：上次健康检查耗费时长，单位ms
serv.Addr：该节点IP：Port
