package storageErr

import "errors"

var (
	ErrTeamExists = errors.New("team already exists")

	ErrUserNotFound = errors.New("user not found")

	ErrPRNotFound = errors.New("pull request not found")
	ErrPRExists   = errors.New("pull request already exists")
	ErrPRMerged   = errors.New("pull request merged")
)
