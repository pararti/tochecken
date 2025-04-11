package tools

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func FormatNumberUSD(numStr string) string {
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		fmt.Printf("не удалось преобразовать строку %s в число: %v", numStr, err)
		return "N/A"
	}

	num = math.Abs(num)
	switch {
	case num >= 1_000_000_000: // миллиард
		return fmt.Sprintf("%.2fB$", num/1_000_000_000)
	case num >= 1_000_000: // миллион
		return fmt.Sprintf("%.2fM$", num/1_000_000)
	case num >= 1_000:
		return fmt.Sprintf("%.2fK$", num/1_000)
	default:
		return fmt.Sprintf("%.2f$", num)
	}
}

func FormatNumberUSDFloat(num float64) string {
	num = math.Abs(num)
	switch {
	case num >= 1_000_000_000: // миллиард
		return fmt.Sprintf("%.2fB$", num/1_000_000_000)
	case num >= 1_000_000: // миллион
		return fmt.Sprintf("%.2fM$", num/1_000_000)
	case num >= 1_000:
		return fmt.Sprintf("%.2fK$", num/1_000)
	default:
		return fmt.Sprintf("%.2f$", num)
	}
}

func PointsToSlashPoints(s string) string {
	return strings.ReplaceAll(s, ".", "\\.")
}
