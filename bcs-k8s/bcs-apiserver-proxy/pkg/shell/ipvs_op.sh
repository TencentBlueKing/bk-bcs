#!/bin/sh

# 查看ipvs规则
ipvsadm -Ln

# 清除所有ipvs规则
ipvsadm -C

# 添加 virtual service, 添加地址为192.168.10.10:80的虚拟服务，指定调度算法为轮转
ipvsadm -A -t 192.168.10.10:80 -s rr -p 600

# 删除 virtual service
ipvsadm -D -t 192.168.10.10:80

# 添加rs
ipvsadm -a -t 192.168.10.10:80 -r 192.168.10.1:80 -g 	# 添加真实服务器，指定传输模式为DR
ipvsadm -a -t 192.168.10.10:80 -r 192.168.10.2:80 -m  # 添加真实服务器，指定传输模式为NAT

# 删除rs
ipvsadm -d -t 192.168.11.100:80 -r 192.168.11.1:80
