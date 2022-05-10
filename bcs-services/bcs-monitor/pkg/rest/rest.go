package rest

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	UnauthorizedError = errors.New("用户未登入")
)

// Result 返回的标准结构
type Result struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestId string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// Context
type Context struct {
	*gin.Context
	RequestId string
}

// HandlerFunc
type HandlerFunc func(*Context) (interface{}, error)

func AbortWithBadRequestError(c *Context, err error) {
	result := Result{Code: 1400, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusBadRequest, result)
}

func AbortWithUnauthorizedError(c *Context, err error) {
	result := Result{Code: 1401, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusUnauthorized, result)
}

func AbortWithWithForbiddenError(c *Context, err error) {
	result := Result{Code: 1403, Message: err.Error(), RequestId: c.RequestId}
	c.AbortWithStatusJSON(http.StatusForbidden, result)
}

func RequestIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.Replace(uid, "-", "", -1)
	return requestId
}

func InitRestContext(c *gin.Context) *Context {
	restContext := &Context{
		Context:   c,
		RequestId: requestid.Get(c),
	}
	c.Set("rest_context", restContext)
	return restContext
}

// GetAuthContext 查询鉴权信息
func GetRestContext(c *gin.Context) (*Context, error) {
	ctxObj, ok := c.Get("rest_context")
	if !ok {
		return nil, UnauthorizedError
	}

	restContext, ok := ctxObj.(*Context)
	if !ok {
		return nil, UnauthorizedError
	}

	return restContext, nil
}

func RestHandlerFunc(handler HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := GetRestContext(c)
		if err != nil {
			AbortWithUnauthorizedError(InitRestContext(c), err)
			return
		}
		result, err := handler(restContext)
		if err != nil {
			AbortWithBadRequestError(restContext, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}
