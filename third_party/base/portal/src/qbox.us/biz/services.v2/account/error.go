package account

type Error struct {
	Error error
}

func (e *Error) IsFailedAuth() bool {
	return e.Error.Error() == "failed_authentication"
}

func (e *Error) IsShortBlocked() bool {
	return e.Error.Error() == "user_short_blocked"
}

func (e *Error) IsExpiredToken() bool {
	return e.Error.Error() == "expired_token"
}

func (e *Error) IsInvalidToken() bool {
	return e.Error.Error() == "invalid_token"
}

func (e *Error) IsPermissionDenied() bool {
	return e.Error.Error() == "permission_denied"
}

func (e *Error) IsRefreshTokenNotExist() bool {
	return e.Error.Error() == "refresh_token_not_exist"
}

func (e *Error) IsUserNotExist() bool {
	// TODO: 需要确认 "Not Found" 是不是等价于 "user_not_exist"
	return e.Error.Error() == "user_not_exist" || e.Error.Error() == "Not Found"
}
