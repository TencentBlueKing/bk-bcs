package configuration

import "fmt"

type AggregationBcsStorageInfo struct {
	bcsStoragePodUrlBase string
	bcsStorageToken      string
}

func (asi *AggregationBcsStorageInfo) SetBcsStorageInfo(acm *AggregationConfigMapInfo) {
	asi.bcsStoragePodUrlBase = fmt.Sprintf("%s/%s", acm.GetBcsStorageAddress(), acm.GetBcsStoragePodUri())
	asi.bcsStorageToken = acm.bcsStorageToken
}

func (asi *AggregationBcsStorageInfo) GetBcsStorageToken() string {
	return asi.bcsStorageToken
}

func (asi *AggregationBcsStorageInfo) GetBcsStoragePodUrlBase() string {
	return asi.bcsStoragePodUrlBase
}
