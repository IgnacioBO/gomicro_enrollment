package enrollment

import (
	"errors"
	"fmt"
)

var ErrUserIDRequired = errors.New("user_id is required")
var ErrCourseIDRequired = errors.New("course_id is required")

var ErrStatusRequired = errors.New("status is required")

var ErrStatusTooLong = errors.New("status cant have more than 2 char")

type ErrEnrollNotFound struct {
	EnrollmentID string
}

func (e ErrEnrollNotFound) Error() string {
	return fmt.Sprintf("enrollment with id: %s not found", e.EnrollmentID)
}

type ErrInvalidStatus struct {
	Status string
}

func (e ErrInvalidStatus) Error() string {
	return fmt.Sprintf("invalid: %s status", e.Status)
}
