package dynamic

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	msgqueue "github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueuev4"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

func PushCreateResourcesToQueue(data operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}

	// queueFlag true
	err := publishDynamicResourceToQueue(data, nsFeatTags, msgqueue.EventTypeUpdate)
	if err != nil {
		blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putNamespaceResources", err)
	}
}

func PushDeleteResourcesToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteNamespaceResources", err)
			}
		}
	}(mList, nsFeatTags)
}

func PushDeleteBatchResourceToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteBatchNamespaceResource", err)
			}
		}
	}(mList, nsListFeatTags)
}

func PushCreateClusterToQueue(data operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	err := publishDynamicResourceToQueue(data, csFeatTags, msgqueue.EventTypeUpdate)
	if err != nil {
		blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "putClusterResources", err)
	}

}

func PushDeleteClusterToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterResources", err)
			}
		}
	}(mList, csFeatTags)
}

func PushDeleteBatchClusterToQueue(mList []operator.M) {
	if !apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		return
	}
	// queueFlag true
	go func(mList []operator.M, featTags []string) {
		for _, data := range mList {
			err := publishDynamicResourceToQueue(data, featTags, msgqueue.EventTypeDelete)
			if err != nil {
				blog.Errorf("func[%s] call publishDynamicResourceToQueue failed: err[%v]", "deleteClusterNamespaceResource", err)
			}
		}
	}(mList, csListFeatTags)
}
