package tools

import (
	"strconv"
	"time"
)

// ParseDuration 解析时间范围字符串（如 5m, 1h, 1d）为 time.Duration
// 支持的时间单位：s (秒), m (分), h (小时), d (天)
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// 尝试直接解析为 time.Duration（支持 ns, us, ms, s, m, h）
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// 支持 'd' (天) 单位
	if len(s) > 0 && s[len(s)-1] == 'd' {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return 0, nil
}

// ParseSinceSeconds 解析时间范围字符串为秒数（用于 K8s 日志查询）
func ParseSinceSeconds(s string) int64 {
	duration, err := ParseDuration(s)
	if err != nil || duration == 0 {
		return 0
	}
	return int64(duration.Seconds())
}
