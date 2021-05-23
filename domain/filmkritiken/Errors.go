package filmkritiken

import (
	"errors"
	"fmt"
)

type (
	RepositoryError struct {
		err error
	}

	NotFoundError struct {
		err error
	}

	InvalidInputError struct {
		err error
	}
)

func NewRepositoryError(err error) *RepositoryError {
	return &RepositoryError{err}
}

func (re *RepositoryError) Error() string {
	return fmt.Sprintf("error in database: %v", re.err)
}

func (re *RepositoryError) Unwrap() error {
	return re.err
}

func NewNotFoundErrorFromString(err string) *NotFoundError {
	return &NotFoundError{errors.New(err)}
}

func (nfe *NotFoundError) Error() string {
	return nfe.err.Error()
}

func (nfe *NotFoundError) Unwrap() error {
	return nfe.err
}

func NewInvalidInputErrorFromString(err string) *InvalidInputError {
	return &InvalidInputError{errors.New(err)}
}

func (nfe *InvalidInputError) Error() string {
	return nfe.err.Error()
}

func (nfe *InvalidInputError) Unwrap() error {
	return nfe.err
}
