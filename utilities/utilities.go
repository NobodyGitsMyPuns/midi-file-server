package utilities

import "fmt"

// WrapError wraps a custom error around an original error, preserving the custom error type for testing
func WrapError(err error, customErr error) error {
	if err != nil {
		return fmt.Errorf("%w: %v", customErr, err)
	}
	return nil
}
