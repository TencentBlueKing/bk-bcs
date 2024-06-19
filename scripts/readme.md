## Dependabot alerts导出到表格
go生成Dependabot alerts导出到表格方法，运行文件：scripts/main.go   
github官方文档地址：https://docs.github.com/en/rest/dependabot/alerts?apiVersion=2022-11-28  

启动需要添加两个环境变量：  
GHToken：github personal access tokens  
OrgName：github fork 项目的组织名称，如：github.com/TencentBlueKing/bk-bcs，组织名称则为TencentBlueKing 

GHToken生成：  
github个人中心 -> Setting -> Developer Settings -> Personal access tokens ->  Tokens (classic) -> 
将生成的密钥复制下来给GHToken用

启动命令：  
GHToken="" OrgName="" go run ./main.go
