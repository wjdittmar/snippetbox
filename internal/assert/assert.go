package assert

import (

	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	// when printing out the error message go will skip over this function
	// and report the file name and calling function of the previous caller
	t.Helper()

	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
