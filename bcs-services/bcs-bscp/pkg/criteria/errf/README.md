# 错误规范
## 错误使用原则
1. 需要返回给普通用户看的错误，统一使用errf.Errorf方法返回错误，也统一为该方法对应的错误msg进行国际化，便于普通用户理解
2. 系统内部常见错误，可直接定义成ErrorF错误对象，有自己特定的错误码和错误msg，便于复用
3. 需要封装携带额外上下文信息，但不需要给普通用户看的错误，统一使用errors.Wrapf方法；不需额外携带上下文，则可直接返回err

## 使用原则详解
### 1. 需要返回给普通用户看的错误，统一使用errf.Errorf方法返回错误，也统一为该方法对应的错误msg进行国际化，便于普通用户理解
#### bcs-services/bcs-bscp/pkg/criteria/errf包下面的Errorf方法
```go
// Errorf 返回自定义封装的bscp错误，包括错误码、错误信息
// bcs-services/bcs-bscp/pkg/rest/response.go中的错误中间件方法GRPCErr会统一进行错误码转换处理
// 需要返回给普通用户看的错误，统一使用该方法返回错误，国际化也以此方法作为提取依据，便于普通用户理解
// 该方法会统一打印错误根因，便于研发排查问题
// 优先使用最底层的bscp错误，越底层的错误越能看出问题根因
func Errorf(err error, code int32, format string, args ...interface{}) error {
	if err == nil {
		err = fmt.Errorf(format, args...)
	}
	// 如果已经是bscp错误，直接返回底层的bscp错误
	if _, ok := err.(BSCPErrI); ok {
		return err
	}
	logs.ErrorDepthf(1, "bscp inner err cause: %v", err)
	return &ErrorF{Code: code, Message: fmt.Sprintf(format, args...)}
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
		return nil, errf.Errorf(nil, errf.InvalidArgument, "template variable name must start with %s",
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

- 日志
```bash
E1204 10:48:33.473870   83335 template_variable.go:44] bscp inner err cause: template variable name must start with bk_bscp_
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
		return nil, errf.ErrDBOpsFailedF(err)
	}
	if err == nil {
		return nil, errf.Errorf(nil, errf.AlreadyExists, "template variable's same name %s already exists",
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

- 日志（同名错误）
```bash
E1204 10:53:25.052671   83328 template_variable.go:42] bscp inner err cause: template variable's same name BK_BSCP_AGE already exists 
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
E1204 10:59:03.758864   83328 template_variable.go:39] bscp inner err cause: dial tcp 127.0.0.1:3306: connect: connection refused
```


**dal层**
- 调用位置示例
```go
// Validate the file format is supported or not.
func (t VariableType) Validate() error {
	switch t {
	case StringVar:
	case NumberVar:
	default:
		return errf.Errorf(nil, errf.InvalidArgument, "unsupported variable type: %s", t)
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
- 日志
```bash
E1204 11:21:04.252386   83328 template_variable.go:210] bscp inner err cause: unsupported variable type: yaml 
```

**公共库或其他代码层报错**
- 调用位置示例
```go
// ValidateVariableName validate bscp variable's length and format.
func ValidateVariableName(name string) error {
	if len(name) < 9 {
		return errf.Errorf(nil, errf.InvalidArgument, "invalid name, "+
			"length should >= 9 and must start with prefix bk_bscp_ (ignore case)")
	}

	if len(name) > 128 {
		return errf.Errorf(nil, errf.InvalidArgument, "invalid name, length should <= 128")
	}

	if !qualifiedVariableNameRegexp.MatchString(name) {
		return errf.Errorf(nil, errf.InvalidArgument,
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
- 日志
```bash
E1204 11:31:31.625247   83328 name.go:80] bscp inner err cause: invalid name: BK_BSCP_AGE{}, only allows to include english、numbers、underscore (_), and must start with prefix bk_bscp_ (ignore case)
```

### 2. 系统内部常见错误，可直接定义成ErrorF错误对象，有自己特定的错误码和错误msg，便于复用
- 常见错误声明
```go
var (
	// ErrDBOpsFailedF is for db operation failed with extra err context
	ErrDBOpsFailedF = func(err error) error {
		return Errorf(err, Internal, "db operation failed")
	}
	// ErrInvalidArgF is for invalid argument with extra err context
	ErrInvalidArgF = func(err error) error {
		return Errorf(err, InvalidArgument, "invalid argument")
	}
	// ErrWithIDF is for id should not be set
	ErrWithIDF = func() error {
		return Errorf(nil, InvalidArgument, "id should not be set")
	}
	// ErrNoSpecF is for spec not set
	ErrNoSpecF = func() error {
		return Errorf(nil, InvalidArgument, "spec not set")
	}
	// ErrNoAttachmentF is for attachment not set
	ErrNoAttachmentF = func() error {
		return Errorf(nil, InvalidArgument, "attachment not set")
	}
	// ErrNoRevisionF is for revision not set
	ErrNoRevisionF = func() error {
		return Errorf(nil, InvalidArgument, "revision not set")
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
		return nil, errf.ErrDBOpsFailedF(err)
	}
	if err == nil {
		return nil, errf.Errorf(nil, errf.AlreadyExists, "template variable's same name %s already exists",
			req.Spec.Name)
	}

	...
}
```

- 调用位置示例（参数错误）
```go
// ValidateCreate validate template variable is valid or not when create it.
func (t *TemplateVariable) ValidateCreate() error {
	if t.ID > 0 {
		return errf.ErrWithIDF()
	}

	if t.Spec == nil {
		return errf.ErrNoSpecF()
	}

	if err := t.Spec.ValidateCreate(); err != nil {
		return err
	}

	if t.Attachment == nil {
		return errf.ErrNoAttachmentF()
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	if t.Revision == nil {
		return errf.ErrNoRevisionF()
	}

	if err := t.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}
```

### 3. 需要封装携带额外上下文信息，但不需要给普通用户看的错误，可使用errors.Wrapf或fmt.Errorf("extra context, %w", err)方法；不需额外携带上下文，则可直接返回err
- 调用示例（更多可全局搜索）
```go
// Download download file from bkrepo
func (c *bkrepoClient) Download(kt *kit.Kit, sign string) (io.ReadCloser, int64, error) {
	node, err := repo.GenNodePath(&repo.NodeOption{Project: c.project, BizID: kt.BizID, Sign: sign})
	if err != nil {
		return nil, 0, err
	}
	
    ...
	
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, 0, errors.Errorf("download status %d != 200", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}
```