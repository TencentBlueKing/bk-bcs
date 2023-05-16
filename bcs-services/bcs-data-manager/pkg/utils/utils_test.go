package utils

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFormatTimeIgnoreMin(t *testing.T) {
	original := "2022-03-08 22:00:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.HourTimeFormat, original, local)
	result := formatTimeIgnoreMin(originalTime)
	assert.Equal(t, "2022-03-08 22:00:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreSec(t *testing.T) {
	original := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.MinuteTimeFormat, original, local)
	result := formatTimeIgnoreSec(originalTime)
	assert.Equal(t, "2022-03-08 22:11:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreHour(t *testing.T) {
	original := "2022-03-08"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.DayTimeFormat, original, local)
	result := formatTimeIgnoreHour(originalTime)
	assert.Equal(t, "2022-03-08 00:00:00 +0800 CST", result.String())
}

func TestGetIndex(t *testing.T) {
	origin := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.MinuteTimeFormat, origin, local)
	day := GetIndex(originalTime, "day")
	hour := GetIndex(originalTime, "hour")
	minute := GetIndex(originalTime, "minute")
	assert.Equal(t, 8, day)
	assert.Equal(t, 22, hour)
	assert.Equal(t, 11, minute)
}
