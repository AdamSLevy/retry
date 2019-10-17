package retry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStop(t *testing.T) {
	err := fmt.Errorf("test")
	assert.EqualError(t, ErrorStop(err), err.Error())
}
