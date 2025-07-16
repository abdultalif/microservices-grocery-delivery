package error

import "errors"

	var (
		ErrRoleHasUsers = errors.New("role has users")
	)
