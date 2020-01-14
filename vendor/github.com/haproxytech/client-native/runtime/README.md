# HAProxy runtime client

## usage

```go
package main

import (
	"log"

	"github.com/haproxytech/client-native/runtime"
)

func main() {
	client := runtime.SingleRuntime{}
	err := client.Init("/var/run/haproxy-runtime-api.sock")

	response, err := client.GetStats()
	if err != nil {
		log.Println(err)
	}
	log.Println(response)

	result, err := client.GetInfo()
	if err != nil {
		log.Println(err)
	}
	log.Println(result)

	result, err = client.ExecuteRaw("show env")
	if err != nil {
		log.Println(err)
	}
	log.Println(result)
}
```