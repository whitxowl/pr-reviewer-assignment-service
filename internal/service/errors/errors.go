package serviceErr

import "errors"

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")

	ErrPRExists        = errors.New("pull request already exists")
	ErrPRNotFound      = errors.New("pull request not found")
	ErrPRAlreadyMerged = errors.New("pull request is already merged")

	ErrAuthorNotCorrect = errors.New("author is not found or has no team")
)
