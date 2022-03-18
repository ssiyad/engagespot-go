package engagespot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHmac(t *testing.T) {
	client := NewEngagespotClient("A", "B")
	assert.Equal(t, client.GenHmac("hello@example.com"), "8c10fc039230663b3b1c074f16db7c7dbb3dd9da64b68965aba85d89acd3a8da")
}
