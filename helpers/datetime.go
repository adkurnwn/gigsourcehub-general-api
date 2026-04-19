package helpers

import (
	"time"
)

func ParseDatetime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func GetCurrentTime() time.Time {
	return time.Now()
}
