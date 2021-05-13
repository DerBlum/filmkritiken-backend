package filmkritiken

import "fmt"

type (
	RepositoryError struct {
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
