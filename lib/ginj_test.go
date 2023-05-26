package lib

import "testing"

func TestCheckHandler(t *testing.T) {
	CheckHandler(func() {})
}
