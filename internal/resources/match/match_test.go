package match

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCapitalized(t *testing.T) {
	assert.True(t, Capitalized(`Yes`))
	assert.False(t, Capitalized(`no`))
}

func TestMatch(t *testing.T) {
	assert.True(t, Match(`MyStructType`, `MyStruct*`))
	assert.True(t, Match(`MyStructType`, `*StructType`))
	assert.True(t, Match(`MyStructType`, `MyStructTyp?`))
	assert.True(t, Match(`MyStructType`, `?yStructType`))
	assert.False(t, Match(`MyStructType`, `Struct`))
}
