package ogenerrors

import "github.com/go-faster/errors"

var (
	// ErrSecurityRequirementIsNotSatisfied is returned when security requirement is not satisfied.
	ErrSecurityRequirementIsNotSatisfied = errors.New("security requirement is not satisfied")
	// ErrSkipClientSecurity is guard error to exclude security scheme from client request.
	ErrSkipClientSecurity = errors.New("skip client security")
	// ErrSkipServerSecurity is guard error to exclude security scheme from server request.
	ErrSkipServerSecurity = errors.New("skip server security")
)
