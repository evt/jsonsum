package main

import (
	"math/big"
	"strings"
)

func findAndSumNumbers(data any) *big.Int {
	total := big.NewInt(0)

	switch val := data.(type) {
	case map[string]any:
		for k, v := range val {
			total.Add(total, findAndSumNumbers(k))
			total.Add(total, findAndSumNumbers(v))
		}
	case []any:
		for _, v := range val {
			total.Add(total, findAndSumNumbers(v))
		}
	case float64:
		total.SetInt64(int64(val))
	case int:
		total.SetInt64(int64(val))
	case string:
		words := strings.Fields(val)
		for _, f := range words {
			num := new(big.Int)
			// try parsing as a float first
			floatNum := new(big.Float)
			_, ok := floatNum.SetString(f)
			if ok {
				// convert big float to big int without allocating new big int
				floatNum.Int(num)
				total.Add(total, num)
				continue
			}
			// parsing as an int/binary/octal/hex
			_, ok = num.SetString(f, 0)
			if ok {
				total.Add(total, num)
			}
			// neither int/binary/octal/hex nor float, ignore
		}
	}

	return total
}
