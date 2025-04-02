package utils

import (
	"fmt"
	"math/big"
)

// FromWei converts a wei amount to a human-readable decimal string based on the token's decimals
func FromWei(amount *big.Int, decimals uint8) string {
	if amount == nil {
		return "0"
	}
	
	// Create a copy of the amount
	wei := new(big.Int).Set(amount)
	
	// Convert to a decimal string based on decimals
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	
	// Integer part
	intPart := new(big.Int).Div(wei, divisor)
	
	// Fractional part
	fracPart := new(big.Int).Mod(wei, divisor)
	
	// Format fractional part with leading zeros
	fracStr := fmt.Sprintf("%0*s", decimals, fracPart.String())
	
	// Trim trailing zeros
	for len(fracStr) > 0 && fracStr[len(fracStr)-1] == '0' {
		fracStr = fracStr[:len(fracStr)-1]
	}
	
	if len(fracStr) > 0 {
		return fmt.Sprintf("%s.%s", intPart.String(), fracStr)
	}
	
	return intPart.String()
}

// ToWei converts a human-readable decimal string to wei based on the token's decimals
func ToWei(amount string, decimals uint8) (*big.Int, error) {
	// Parse the decimal amount
	parts := splitDecimal(amount)
	intPart, fracPart := parts[0], ""
	if len(parts) > 1 {
		fracPart = parts[1]
	}
	
	// Parse integer part
	intValue, ok := new(big.Int).SetString(intPart, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part: %s", intPart)
	}
	
	// Multiply by 10^decimals
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	intValue.Mul(intValue, multiplier)
	
	// Add fractional part if it exists
	if fracPart != "" {
		// Pad or truncate fractional part to match decimals
		if len(fracPart) > int(decimals) {
			fracPart = fracPart[:decimals]
		} else {
			fracPart = fracPart + "0000000000000000000"[:int(decimals)-len(fracPart)]
		}
		
		// Parse fractional part
		fracValue, ok := new(big.Int).SetString(fracPart, 10)
		if !ok {
			return nil, fmt.Errorf("invalid fractional part: %s", fracPart)
		}
		
		// Add to result
		intValue.Add(intValue, fracValue)
	}
	
	return intValue, nil
}