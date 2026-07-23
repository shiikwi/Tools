package util

import (
	"fmt"
	"strings"
)

func Bytes2HexStr(data []byte) string {
	values := make([]string, len(data))

	for i, value := range data {
		values[i] = fmt.Sprintf("%02X", value)
	}

	return strings.Join(values, " ")
}
