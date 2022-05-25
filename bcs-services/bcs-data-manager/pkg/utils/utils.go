package utils

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"time"
)

func formatTimeIgnoreSec(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(types.MinuteTimeFormat, originalTime.Format(types.MinuteTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore second error :%v", err)
		return originalTime
	}
	return formatString
}

func formatTimeIgnoreMin(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(types.HourTimeFormat, originalTime.Format(types.HourTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore minute error :%v", err)
		return originalTime
	}
	return formatString
}

func formatTimeIgnoreHour(originalTime time.Time) time.Time {
	local := time.Local
	formatString, err := time.ParseInLocation(types.DayTimeFormat, originalTime.Format(types.DayTimeFormat), local)
	if err != nil {
		blog.Errorf("format time ignore day error :%v", err)
		return originalTime
	}
	return formatString
}

// FormatTime format time
func FormatTime(originalTime time.Time, dimension string) time.Time {
	switch dimension {
	case types.DimensionDay:
		return formatTimeIgnoreHour(originalTime)
	case types.DimensionHour:
		return formatTimeIgnoreMin(originalTime)
	case types.DimensionMinute:
		return formatTimeIgnoreSec(originalTime)
	default:
		return originalTime
	}
}

// GetBucketTime get bucket time
func GetBucketTime(currentTime time.Time, dimension string) (string, error) {
	switch dimension {
	case types.DimensionDay:
		return currentTime.Format(types.MonthTimeFormat), nil
	case types.DimensionHour:
		return currentTime.Format(types.DayTimeFormat), nil
	case types.DimensionMinute:
		return currentTime.Format(types.HourTimeFormat), nil
	default:
		return "", fmt.Errorf("wrong dimension :%s", dimension)
	}
}

// GetIndex get metric index
func GetIndex(currentTime time.Time, dimension string) int {
	switch dimension {
	case types.DimensionDay:
		return currentTime.Day()
	case types.DimensionHour:
		return currentTime.Hour()
	case types.DimensionMinute:
		return currentTime.Minute()
	default:
		return 0
	}
}
