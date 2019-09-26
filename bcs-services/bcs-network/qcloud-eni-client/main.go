package main

import (
	"flag"
)

var cvmIPs string
var action string
var ipNum int

func main() {
	flag.StringVar(&action, "action", "", "action for qcloud cvm")
	flag.StringVar(&cvmIPs, "cvm", "", "ips for cvm to apply eni")
	flag.IntVar(&ipNum, "ipnum", 0, "ip num for each cvm")
}
