package constant

import "time"

const (
	//CurrentUserAttr user header
	CurrentUserAttr = "current-user"

	// DefaultTokenLength user token default length
	// token is consisted of digital and alphabet(case sensetive)
	// we can refer to http://coolaf.com/tool/rd when testing
	DefaultTokenLength = 32
	// TokenKeyPrefix is the redis key for token
	TokenKeyPrefix = "bcs_auth:token:"
	// TokenLimits for token
	TokenLimits    = 1
	TokenMaxExpire = time.Hour * 24 * 365 // 1 year
)
