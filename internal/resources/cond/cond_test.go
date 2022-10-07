package cond

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirst_String(t *testing.T) {
	s := `abc`
	assert.Equal(t, s, First(s).(string))
	assert.Equal(t, s, First(s, ``, ``).(string))
	assert.Equal(t, s, First(``, s, ``).(string))
	assert.Equal(t, s, First(``, ``, s).(string))
	assert.Equal(t, ``, First(``).(string))
	assert.Equal(t, ``, First(``, ``).(string))
	assert.Equal(t, ``, First(``, ``).(string))
}

func TestFirst_Num(t *testing.T) {
	d := 42
	assert.Equal(t, d, First(d).(int))
	assert.Equal(t, d, First(0, d).(int))
	assert.Equal(t, d, First(d, 1, 2).(int))
	assert.Equal(t, d, First(0, d, 2).(int))
	assert.Equal(t, d, First(0, 0, d).(int))
	assert.Equal(t, 0, First(0, 0).(int))
}

func TestFirst_Bool(t *testing.T) {
	assert.True(t, First(true).(bool))
	assert.True(t, First(false, true, false).(bool))
	assert.True(t, First(false, false, true).(bool))
	assert.False(t, First(false, false, false).(bool))
}

func TestFirst_Nil(t *testing.T) {
	assert.Nil(t, First(nil))
	assert.Nil(t, First(nil, nil))
	assert.Nil(t, First(nil, nil, nil))
}

func TestFirst_Struct(t *testing.T) {
	type mytype struct {
	}
	typ := mytype{}
	v, ok := First(nil, typ).(mytype)
	assert.True(t, ok)
	assert.Equal(t, typ, v)
	v, ok = First(typ, nil).(mytype)
	assert.True(t, ok)
	assert.Equal(t, typ, v)
	v, ok = First(nil, typ, nil).(mytype)
	assert.True(t, ok)
	assert.Equal(t, typ, v)
}
