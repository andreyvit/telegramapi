package main

import (
	"fmt"
	"time"
)

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func MakeDate(y int, m time.Month, d int) Date {
	return Date{y, m, d}
}

func (d Date) IsZero() bool {
	return d.Year == 0 && d.Month == 0 && d.Day == 0
}

func (d Date) Equal(u Date) bool {
	return d.Year == u.Year && d.Month == u.Month && d.Day == u.Day
}

func (d Date) String() string {
	return fmt.Sprintf("%02d.%02d.%04d", d.Day, d.Month, d.Year)
}
