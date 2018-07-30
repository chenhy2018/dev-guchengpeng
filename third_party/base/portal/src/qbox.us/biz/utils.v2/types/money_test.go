package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoneyString(t *testing.T) {
	assert.Equal(t, Money(1).String(), "0.0001")
	assert.Equal(t, Money(10).String(), "0.001")
	assert.Equal(t, Money(100).String(), "0.01")
	assert.Equal(t, Money(1000).String(), "0.10")
	assert.Equal(t, Money(100100).String(), "10.01")
	assert.Equal(t, Money(-100100).String(), "-10.01")
	assert.Equal(t, Money(-1).String(), "-0.0001")

	m := Money(-1)
	m.String()
	assert.Equal(t, int64(m), -1)
}

func TestMoneyAsFen(t *testing.T) {
	assert.Equal(t, Money(1).AsFen(), "￥ 0.00")
	assert.Equal(t, Money(10).AsFen(), "￥ 0.00")
	assert.Equal(t, Money(100).AsFen(), "￥ 0.01")
	assert.Equal(t, Money(1000).AsFen(), "￥ 0.10")
	assert.Equal(t, Money(1e10).AsFen(), "￥ 1,000,000.00")
	assert.Equal(t, Money(100100).AsFen(), "￥ 10.01")
	assert.Equal(t, Money(-100100).AsFen(), "￥ -10.01")
	assert.Equal(t, Money(-1e10).AsFen(), "￥ -1,000,000.00")
	assert.Equal(t, Money(-1).AsFen(), "￥ -0.00")
	assert.Equal(t, Money(99999).AsFen(), "￥ 9.99")
	assert.Equal(t, Money(-99999).AsFen(), "￥ -9.99")
}

func TestMoneyHumanize(t *testing.T) {
	assert.Equal(t, Money(1).Humanize(), "0.00")
	assert.Equal(t, Money(10).Humanize(), "0.00")
	assert.Equal(t, Money(100).Humanize(), "0.01")
	assert.Equal(t, Money(1000).Humanize(), "0.10")
	assert.Equal(t, Money(1e10).Humanize(), "1,000,000.00")
	assert.Equal(t, Money(100100).Humanize(), "10.01")
	assert.Equal(t, Money(-100100).Humanize(), "-10.01")
	assert.Equal(t, Money(-1e10).Humanize(), "-1,000,000.00")
	assert.Equal(t, Money(-1).Humanize(), "-0.00")
	assert.Equal(t, Money(99999).Humanize(), "9.99")
	assert.Equal(t, Money(-99999).Humanize(), "-9.99")

	assert.Equal(t, Money(1).Humanize(3), "0.00")
	assert.Equal(t, Money(10).Humanize(3), "0.001")
	assert.Equal(t, Money(100).Humanize(3), "0.01")
	assert.Equal(t, Money(1000).Humanize(3), "0.10")
	assert.Equal(t, Money(1e10).Humanize(3), "1,000,000.00")
	assert.Equal(t, Money(100100).Humanize(3), "10.01")
	assert.Equal(t, Money(-100100).Humanize(3), "-10.01")
	assert.Equal(t, Money(-1e10).Humanize(3), "-1,000,000.00")
	assert.Equal(t, Money(-1).Humanize(3), "-0.00")
	assert.Equal(t, Money(99999).Humanize(3), "9.999")
	assert.Equal(t, Money(-99999).Humanize(3), "-9.999")
}

func TestMoneyYuanMoney(t *testing.T) {
	for i := 0; i < 10000; i += 100 {
		assert.Equal(t, Money(i), Money(i).MoneyYuan().Money())
	}
	assert.Equal(t, MoneyYuan(-11.0).Money(), Money(-110000))
	assert.Equal(t, MoneyYuan(0.0001).Money(), Money(1))
	assert.Equal(t, MoneyYuan(0.00014).Money(), Money(1))
	assert.Equal(t, MoneyYuan(0.00016).Money(), Money(2))
	assert.Equal(t, MoneyYuan(0.0002).Money(), Money(2))
	assert.Equal(t, MoneyYuan(-0.0001).Money(), Money(-1))
	assert.Equal(t, MoneyYuan(-0.00014).Money(), Money(-1))
	assert.Equal(t, MoneyYuan(-0.00016).Money(), Money(-2))
	assert.Equal(t, MoneyYuan(-0.0002).Money(), Money(-2))
}
