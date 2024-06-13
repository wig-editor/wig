package mcwig

import (
	"testing"

	"github.com/firstrow/mcwig/testutils"
	"github.com/stretchr/testify/assert"
)

func TestProjectFindRoot(t *testing.T) {
	e := NewEditor(testutils.Viewport, nil)
	e.OpenFile("/home/andrew/code/mcwig/buffer_test.txt")

	r, err := e.Projects.FindRoot(e.Buffers[0])

	assert.NoError(t, err)
	assert.Equal(t, "/home/andrew/code/mcwig", r)
}
