package xrequests

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/levigross/grequests"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
)

type (
	RequestOptions = grequests.RequestOptions
	Response       = grequests.Response

	SendFunc func(ctx context.Context, url string, params any, responseStruct any, opts ...*RequestOptions) (trace *Trace, response *Response, e error)
)

const (
	// DefaultTimeoutSecond default timeout
	DefaultTimeoutSecond = 15
	// DefaultKeepAliveSecond keep alive
	DefaultKeepAliveSecond = 5
)

// Trace request trace
type Trace struct {
	Url          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	Query        any               `json:"query"`
	Body         any               `json:"body"`
	StatusCode   int               `json:"status_code"`
	ResponseBody string            `json:"response_body"`
}

func buildRequestOptions(ctx context.Context, opts []*RequestOptions) *RequestOptions {
	opt := &RequestOptions{}
	for _, o := range opts {
		if o != nil {
			opt = o
			break
		}
	}
	if opt.DialTimeout == 0 {
		opt.DialTimeout = time.Duration(DefaultTimeoutSecond) * time.Second
	}
	if opt.RequestTimeout == 0 {
		opt.RequestTimeout = time.Duration(DefaultTimeoutSecond) * time.Second
	}
	if opt.DialKeepAlive == 0 {
		opt.DialKeepAlive = time.Duration(DefaultKeepAliveSecond) * time.Second
	}
	return opt
}

func newTrace(ctx context.Context, url string, method string, opt *RequestOptions) *Trace {
	trace := &Trace{
		Url:     url,
		Method:  method,
		Headers: opt.Headers,
	}
	if opt.Data != nil {
		trace.Body = opt.Data
	} else if opt.JSON != nil {
		trace.Body = opt.JSON
	}
	trace.Query = opt.Params
	return trace
}

func wrapResponse(ctx context.Context, trace *Trace, response *Response, err error) (*Trace, *Response, error) {
	responseBody := string(response.Bytes())
	trace.StatusCode = response.StatusCode
	trace.ResponseBody = responseBody
	if err != nil {
		blog.Errorf("[Err] %s; [Trace] %s", err.Error(), common.JsonMarshal(trace))
		return trace, response, err
	}

	if !response.Ok {
		err = errors.Errorf("StatusCode: %d; Body: %s", response.StatusCode, responseBody)
		return trace, response, err
	}
	return trace, response, err
}

// NativeGet is a wrapper of grequests.Get
func NativeGet(ctx context.Context, url string, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	trace := newTrace(ctx, url, http.MethodGet, opt)
	response, e := grequests.Get(url, opt)
	return wrapResponse(ctx, trace, response, e)
}

// NativePost is a wrapper of grequests.Post
func NativePost(ctx context.Context, url string, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	trace := newTrace(ctx, url, http.MethodPost, opt)
	response, e := grequests.Post(url, opt)
	return wrapResponse(ctx, trace, response, e)
}

// NativeDelete is a wrapper of grequests.Delete
func NativeDelete(ctx context.Context, url string, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	trace := newTrace(ctx, url, http.MethodDelete, opt)
	response, e := grequests.Delete(url, opt)
	return wrapResponse(ctx, trace, response, e)
}

// NativePut is a wrapper of grequests.Put
func NativePut(ctx context.Context, url string, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	trace := newTrace(ctx, url, http.MethodPut, opt)
	response, err := grequests.Put(url, opt)
	return wrapResponse(ctx, trace, response, err)
}

// Get will call NativeGET with default options
func Get(ctx context.Context, url string, params any, responseStruct any, opts ...*RequestOptions) (*Trace, *Response, error) {
	queryMap, err := buildQuery(params)
	if err != nil {
		return nil, nil, fmt.Errorf("build query failed, err: %w", err)
	}
	opt := buildRequestOptions(ctx, opts)
	opt.Params = queryMap
	trace, response, e := NativeGet(ctx, url, opt)
	if e != nil {
		return nil, nil, e
	}
	err = jsonUnmarshalResponse(response, responseStruct)
	if err != nil {
		return nil, nil, err
	}
	return trace, response, e
}

// Post will convert the struct into JSON and then call NativePOST
func Post(ctx context.Context, url string, params any, responseStruct any, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	opt.JSON = params
	trace, response, e := NativePost(ctx, url, opt)
	if e != nil {
		return nil, nil, e
	}
	err := jsonUnmarshalResponse(response, responseStruct)
	if err != nil {
		return nil, nil, err
	}
	return trace, response, e
}

// Delete will convert the structure into Json and then call NativeDELETE
func Delete(ctx context.Context, url string, params any, responseStruct any, opts ...*RequestOptions) (*Trace,
	*Response, error) {
	opt := buildRequestOptions(ctx, opts)
	opt.JSON = params
	trace, response, e := NativeDelete(ctx, url, opt)
	if e != nil {
		return nil, nil, e
	}
	err := jsonUnmarshalResponse(response, responseStruct)
	if err != nil {
		return nil, nil, err
	}
	return trace, response, e
}

// Put is a shortcut for calling NativePUT
func Put(ctx context.Context, url string, params any, responseStruct any, opts ...*RequestOptions) (*Trace, *Response, error) {
	opt := buildRequestOptions(ctx, opts)
	opt.JSON = params
	trace, response, err := NativePut(ctx, url, opt)
	if err != nil {
		return nil, nil, err
	}
	err = jsonUnmarshalResponse(response, responseStruct)
	if err != nil {
		return nil, nil, err
	}
	return trace, response, nil
}

func jsonUnmarshalResponse(response *Response, v any) error {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return errors.Errorf("non-pointer field")
	}
	if err := response.JSON(v); err != nil {
		return fmt.Errorf("response'%s' json unmarshal failed, err: %w", response.String(), err)
	}
	return nil
}
