package numutil

import "strconv"

func FormatFloat(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}
