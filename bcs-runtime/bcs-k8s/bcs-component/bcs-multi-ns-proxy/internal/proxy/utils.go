package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

const (
	// DefaultLegacyAPIPrefix is where the legacy APIs will be located.
	DefaultLegacyAPIPrefix = "/api"
	// APIGroupPrefix is where non-legacy API group will be located.
	APIGroupPrefix = "/apis"
)

func getNamespaceFromRequest(req *http.Request) (string, error) {
	apiPrefixes := sets.NewString(strings.Trim(APIGroupPrefix, "/"))
	legacyAPIPrefixes := sets.String{}
	apiPrefixes.Insert(strings.Trim(DefaultLegacyAPIPrefix, "/"))
	legacyAPIPrefixes.Insert(strings.Trim(DefaultLegacyAPIPrefix, "/"))

	requestInfoFactory := &apirequest.RequestInfoFactory{
		APIPrefixes:          apiPrefixes,
		GrouplessAPIPrefixes: legacyAPIPrefixes,
	}

	requestInfo, err := requestInfoFactory.NewRequestInfo(req)
	if err != nil {
		return "", fmt.Errorf("create info from request %s %s failed, err %s",
			req.RemoteAddr, req.URL.String(), err.Error())
	}
	return requestInfo.Namespace, nil
}
