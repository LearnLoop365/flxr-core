package sqldb

import (
	"strconv"
	"strings"
)

func ConvertStaticPlaceholders(sql string, prefix byte) string {
	if prefix == '?' || prefix == 0 {
		return sql
	}
	var builder strings.Builder
	builder.Grow(len(sql) + 8) // small padding; rough pre-optimization
	cnt := 1
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' {
			builder.WriteByte(prefix)
			builder.WriteString(strconv.Itoa(cnt))
			cnt++
		} else {
			builder.WriteByte(sql[i])
		}
	}
	return builder.String()
}
