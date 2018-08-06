package types

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Money int64

func (m Money) MoneyYuan() MoneyYuan {
	return MoneyYuan(float64(m) / 1e4)
}

func (m Money) IsZero() bool {
	return int64(m) == 0
}

func (m Money) String() string {
	symbol := ""
	if m < 0 {
		symbol = "-"
		m = -m
	}
	if m%100 == 0 {
		return fmt.Sprintf("%s%d.%02d", symbol, m/10000, (m%10000)/100)
	} else if m%10 == 0 {
		return fmt.Sprintf("%s%d.%03d", symbol, m/10000, (m%10000)/10)
	}
	return fmt.Sprintf("%s%d.%04d", symbol, m/10000, m%10000)
}

func (m Money) AsFen() string {
	return "￥ " + m.Humanize()
}

func (m Money) Humanize(accuracy ...int) string {
	acc := 2
	if len(accuracy) > 0 && accuracy[0] <= 4 {
		acc = accuracy[0]
	}

	var res string
	if m < 0 {
		res += "-"
		m = -m
	}
	res += thousandDelimeter(int64(m) / 10000)
	base := int64(math.Pow10(4 - acc))
	res += fmt.Sprintf(fmt.Sprintf(".%%0%dd", acc), int64(m%10000)/base)
	moneyStrArr := strings.Split(res, ".")
	if len(moneyStrArr) != 1 {
		decimal := moneyStrArr[1]
		decimal = strings.TrimRight(decimal, "0")
		if len(decimal) < 2 {
			decimal += strings.Repeat("0", 2-len(decimal))
		}
		res = fmt.Sprintf("%s.%s", moneyStrArr[0], decimal)
	}
	return res
}

func thousandDelimeter(i int64) string {
	if i >= 1e3 {
		return fmt.Sprintf("%s,%03d", thousandDelimeter(i/1e3), i%1e3)
	} else {
		return fmt.Sprintf("%d", i)
	}
}

//------------------------------------------

type MoneyYuan float64

func (y MoneyYuan) Money() Money {
	// http://golang.org/ref/spec#Conversions
	// When converting a floating-point number to an integer, the fraction is discarded (truncation towards zero).
	if y < 0 {
		return Money(y*10000 - 0.5)
	} else {
		return Money(y*10000 + 0.5)
	}
}

func ParseMoneyYuan(s string) (money Money, err error) {
	m, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return
	}
	money = MoneyYuan(m).Money()
	return
}
