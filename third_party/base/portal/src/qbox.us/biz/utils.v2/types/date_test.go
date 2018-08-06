package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakeDateFromTime(t *testing.T) {
	t1 := time.Date(2013, time.February, 14, 0, 0, 0, 0, time.Local)
	assert.Equal(t, MakeDateFromTime(t1), "2013-02-14")
}

func TestDateTime(t *testing.T) {
	d := Date("2013-02-14")
	t1, err := d.Time()
	assert.Nil(t, err)
	assert.Equal(t, t1.Unix(), time.Date(2013, time.February, 14, 0, 0, 0, 0, time.Local).Unix())
}

func TestMonthTime(t *testing.T) {
	res := Month("2013-02").Time()
	assert.Equal(t, res.Unix(), time.Date(2013, time.February, 1, 0, 0, 0, 0, time.Local).Unix())
}

func TestMakeMonthFromTime(t *testing.T) {
	t1 := time.Date(2013, time.February, 14, 0, 0, 0, 0, time.Local)
	assert.Equal(t, MakeMonthFromTime(t1), "2013-02")
}

func TestMonthNextMonth(t *testing.T) {
	a := []string{"2012-11", "2012-12", "2013-01", "2013-02", "2013-03"}
	for i := 0; i < 4; i++ {
		assert.Equal(t, Month(a[i]).NextMonth(), Month(a[i+1]))
	}
}
func TestMonthPrevMonth(t *testing.T) {
	a := []string{"2012-11", "2012-12", "2013-01", "2013-02", "2013-03"}
	for i := 1; i < 5; i++ {
		assert.Equal(t, Month(a[i]).PrevMonth(), Month(a[i-1]))
	}
}

func TestMonthStartDate(t *testing.T) {
	a := map[string]string{
		"2012-12": "2012-12-01",
		"2013-01": "2013-01-01",
		"2013-02": "2013-02-01",
		"2013-03": "2013-03-01",
	}
	for k, v := range a {
		assert.Equal(t, Month(k).StartDate(), Date(v))
	}
}
func TestMonthEndDate(t *testing.T) {
	a := map[string]string{
		"2012-02": "2012-02-29",
		"2012-12": "2012-12-31",
		"2013-01": "2013-01-31",
		"2013-02": "2013-02-28",
		"2013-03": "2013-03-31",
	}
	for k, v := range a {
		assert.Equal(t, Month(k).EndDate(), Date(v))
	}
}
