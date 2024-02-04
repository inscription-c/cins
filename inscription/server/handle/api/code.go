package api

type Code = int

// common
const (
	CodeSuccess        Code = 0
	CodeError500       Code = 500
	CodeParamsInvalid  Code = 10000
	CodeMethodNotExist Code = 10001
	CodeDbError        Code = 10002
	CodeCacheError     Code = 10003
)
