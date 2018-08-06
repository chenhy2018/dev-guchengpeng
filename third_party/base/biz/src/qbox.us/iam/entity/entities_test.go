package entity_test

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
	"qbox.us/iam/enums"
)

func TestEffect(t *testing.T) {
	expectFuncs := map[enums.Effect]func(enums.Effect) bool{
		enums.EffectAllow:   enums.Effect.IsAllow,
		enums.EffectDeny:    enums.Effect.IsDeny,
		enums.EffectUnknown: enums.Effect.IsUnknown,
	}
	for effect, _ := range expectFuncs {
		e := enums.MakeEffect(effect.String())
		if assert.Equal(t, effect, e, "effect: %+v", effect) {
			for effect2, fn := range expectFuncs {
				assert.Equal(t, effect == effect2, fn(e))
			}
		}
	}
}
