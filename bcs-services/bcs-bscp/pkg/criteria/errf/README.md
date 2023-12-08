# 错误规范
## 错误使用原则
1. 需要返回给普通用户看的错误，统一使用errf.Errorf方法返回错误，也统一为该方法对应的错误msg进行国际化，便于普通用户理解
2. 系统内部常见错误，可直接定义成ErrorF错误对象，有自己特定的错误码和错误msg，便于复用
3. 需要封装携带额外上下文信息，但不需要给普通用户看的错误，可使用errors.Wrapf或fmt.Errorf("extra context, %w", err)等方法；不需额外携带上下文，则可直接返回err

## 使用原则详解
### 1. 需要返回给普通用户看的错误，统一使用errf.Errorf方法返回错误，也统一为该方法对应的错误msg进行国际化，便于普通用户理解
#### bcs-services/bcs-bscp/pkg/criteria/errf包下面的Errorf方法
```go
// ErrorF defines an error with error code and message.
type ErrorF struct {
	// Kit is bscp kit
	Kit *kit.Kit
	// Code is bscp errCode
	Code int32 `json:"code"`
	// Message is error detail
	Message string `json:"message"`
}

// Errorf 返回自定义封装的bscp错误，包括错误码、错误信息
// bcs-services/bcs-bscp/pkg/rest/response.go中的错误中间件方法GRPCErr会统一进行错误码转换处理
// 需要返回给普通用户看的错误，统一使用该方法返回错误，且对错误信息进行国际化处理，便于普通用户理解
func Errorf(kit *kit.Kit, code int32, format string, args ...interface{}) *ErrorF {
	return &ErrorF{
		Kit:  kit,
		Code: code,
		// 错误信息国际化
		Message: localizer.Get(kit.Lang).Translate(format, args...),
	}
}

// Error implement the golang's basic error interface
func (e *ErrorF) Error() string {
	if e == nil || e.Code == OK {
		return "nil"
	}

	// return with a json format string error, so that the upper service
	// can use Wrap to decode it.
	return fmt.Sprintf(`{"code": %d, "message": "%s"}`, e.Code, e.Message)
}

// WithCause 打印根因错误，有底层错误需要暴露时调用该方法，便于研发排查问题
func (e *ErrorF) WithCause(cause error) *ErrorF {
	if cause == nil {
		return e
	}

	// 如果底层根因错误已经是bscp错误，直接使用该根因错误
	if c, ok := cause.(*ErrorF); ok {
		return c
	}
    // 打印其他错误根因日志
    logs.ErrorDepthf(1, "bscp inner err cause: %v, rid: %s", cause, e.Kit.Rid)
	return e
}

// GRPCStatus implements interface{ GRPCStatus() *Status } , so that it can be recognized by grpc
func (e *ErrorF) GRPCStatus() *status.Status {
	return status.New(codes.Code(e.Code), e.Message)
}
```

#### 使用示例
**config-server层** 
- 调用位置示例
```go
// CreateTemplateVariable create a template variable
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbcs.CreateTemplateVariableReq) (
	*pbcs.CreateTemplateVariableResp, error) {
	...

	if !strings.HasPrefix(strings.ToLower(req.Name), constant.TemplateVariablePrefix) {
        return nil, errf.Errorf(grpcKit, errf.InvalidArgument, "template variable name must start with %s",
            constant.TemplateVariablePrefix)
}

	...
}
```

- response
```json
{
    "error": {
        "code": "INVALID_ARGUMENT",
        "message": "template variable name must start with bk_bscp_",
        "data": null,
        "details": []
    }
}
```

**data-service层**
- 调用位置示例
```go
// CreateTemplateVariable create template variable.
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbds.CreateTemplateVariableReq) (*pbds.CreateResp,
	error) {
	...

	_, err := s.dao.TemplateVariable().GetByUniqueKey(kt, req.Attachment.BizId, req.Spec.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errf.ErrDBOpsFailedF(kt).WithCause(err)
	}
	if err == nil {
		return nil, errf.Errorf(kt, errf.AlreadyExists, "template variable's same name %s already exists",
			req.Spec.Name)
	}

	...
}
```

- response（同名错误）
```json
{
    "error": {
        "code": "ALREADY_EXISTS",
        "message": "template variable's same name BK_BSCP_AGE already exists",
        "data": null,
        "details": []
    }
}
```

- response（DB操作错误）
```json
{
  "error": {
    "code": "INTERNAL",
    "message": "db operation failed",
    "data": null,
    "details": []
  }
}
```

- 日志（DB操作错误）
```bash
E1204 10:59:03.758864   83328 template_variable.go:39] bscp inner err cause: dial tcp 127.0.0.1:3306: connect: connection refused, rid: b9629955c9664069af8ce0ec5245eb22
```


**dal层**
- 调用位置示例
```go
// Validate the file format is supported or not.
func (t VariableType) Validate(kit *kit.Kit) error {
	switch t {
	case StringVar:
	case NumberVar:
	default:
		return errf.Errorf(kit, errf.InvalidArgument, "unsupported variable type: %s", t)
	}

	return nil
}
```

- response
```json
{
    "error": {
        "code": "INVALID_ARGUMENT",
        "message": "unsupported variable type: yaml",
        "data": null,
        "details": []
    }
}
```

**公共库或其他代码层报错**
- 调用位置示例
```go
// ValidateVariableName validate bscp variable's length and format.
func ValidateVariableName(kit *kit.Kit, name string) error {
	if len(name) < 9 {
		return errf.Errorf(kit, errf.InvalidArgument, "invalid name, "+
			"length should >= 9 and must start with prefix bk_bscp_ (ignore case)")
	}

	if len(name) > 128 {
		return errf.Errorf(kit, errf.InvalidArgument, "invalid name, length should <= 128")
	}

	if !qualifiedVariableNameRegexp.MatchString(name) {
		return errf.Errorf(kit, errf.InvalidArgument,
			"invalid name: %s, only allows to include english、numbers、underscore (_)"+
				", and must start with prefix bk_bscp_ (ignore case)", name)
	}

	return nil
}
```

- response
```json
{
    "error": {
        "code": "INVALID_ARGUMENT",
        "message": "invalid name: BK_BSCP_AGE{}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)",
        "data": null,
        "details": []
    }
}
```

### 2. 系统内部常见错误，可直接定义成ErrorF错误对象，有自己特定的错误码和错误msg，便于复用
- 常见错误声明
```go
var (
	// ErrDBOpsFailedF is for db operation failed
	ErrDBOpsFailedF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, Internal, "db operation failed")
	}
	// ErrInvalidArgF is for invalid argument
	ErrInvalidArgF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "invalid argument")
	}
	// ErrWithIDF is for id should not be set
	ErrWithIDF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "id should not be set")
	}
	// ErrNoSpecF is for spec not set
	ErrNoSpecF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "spec not set")
	}
	// ErrNoAttachmentF is for attachment not set
	ErrNoAttachmentF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "attachment not set")
	}
	// ErrNoRevisionF is for revision not set
	ErrNoRevisionF = func(kit *kit.Kit) *ErrorF {
		return Errorf(kit, InvalidArgument, "revision not set")
	}
)
```

- 调用位置示例（DB操作错误，同上面的dataservice的示例之一）
```go
// CreateTemplateVariable create template variable.
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbds.CreateTemplateVariableReq) (*pbds.CreateResp,
	error) {
	...

	_, err := s.dao.TemplateVariable().GetByUniqueKey(kt, req.Attachment.BizId, req.Spec.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errf.ErrDBOpsFailedF(kt).WithCause(err)
	}
	if err == nil {
		return nil, errf.Errorf(kt, errf.AlreadyExists, "template variable's same name %s already exists",
			req.Spec.Name)
	}

	...
}
```

- 调用位置示例（参数错误）
```go
// ValidateCreate validate ReleasedAppTemplateVariable is valid or not when created.
func (t *ReleasedAppTemplateVariable) ValidateCreate(kit *kit.Kit) error {
	if t.Spec != nil {
		if err := t.Spec.ValidateCreate(kit); err != nil {
			return err
		}
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	if t.Revision == nil {
		return errors.New("revision not set")
	}

	if err := t.Revision.Validate(); err != nil {
		return err
	}

	return nil
}
```

### 3. 需要封装携带额外上下文信息，但不需要给普通用户看的错误，可使用errors.Wrapf或fmt.Errorf("extra context, %w", err)等方法；不需额外携带上下文，则可直接返回err
- 调用示例（更多可全局搜索）
```go
// Parse api-gateway request header to context kit and validate.
func (p *jwtParser) Parse(ctx context.Context, header http.Header) (*kit.Kit, error) {
	jwtToken := header.Get(constant.BKGWJWTTokenKey)
	if len(jwtToken) == 0 {
		return nil, errors.Errorf("jwt token header %s is required", constant.BKGWJWTTokenKey)
	}

	token, err := p.parseToken(jwtToken)
	if err != nil {
		return nil, errors.Wrapf(err, "parse jwt token %s", jwtToken)
	}

	if err := token.validate(); err != nil {
		return nil, errors.Wrapf(err, "validate jwt token %s", jwtToken)
	}

	...

	if err := kt.Validate(); err != nil {
		return nil, errors.Wrapf(err, "validate kit")
	}

	return kt, nil
}
```