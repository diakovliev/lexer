package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamed(t *testing.T) {
	b := makeTestDisposeBuilder()
	named := b.Named("test")
	assert.Equal(t, "test", named.nodeName)

	err := named.deref().Update(context.Background(), nil)
	assert.ErrorIs(t, err, ErrChainNext)

	s0 := named.append("s0", newFakeState)
	assert.Equal(t, "test.s0", s0.nodeName)

	s1 := s0.append("s1", newFakeState)
	assert.Equal(t, "test.s0.s1", s1.nodeName)

	assert.Panics(t, func() {
		s1.Named("bad")
	})
}
