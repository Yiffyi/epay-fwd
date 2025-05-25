package epay

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// CalculateSign calculates the sign for the given request and key
func CalculateSign(i EpaySignedRequest, key string) (string, error) {
	// Extract all fields from the request
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", errors.New("request is not a struct")
	}

	// Collect all fields as key-value pairs
	params := make(map[string]string)
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		tag := field.Tag.Get("json")
		if tag == "" {
			continue
		}

		// Extract the field name from the tag
		parts := strings.Split(tag, ",")
		name := parts[0]

		// Skip sign and sign_type fields
		if name == "sign" || name == "sign_type" {
			continue
		}

		// Get the field value as string
		fieldVal := val.Field(i)
		var strVal string

		switch fieldVal.Kind() {
		case reflect.String:
			strVal = fieldVal.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			strVal = fmt.Sprintf("%d", fieldVal.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			strVal = fmt.Sprintf("%d", fieldVal.Uint())
		case reflect.Float32, reflect.Float64:
			strVal = fmt.Sprintf("%g", fieldVal.Float())
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", fieldVal.Bool())
		default:
			// Skip complex types
			continue
		}

		// Skip empty values
		if strVal == "" {
			continue
		}

		params[name] = strVal
	}

	// Sort the keys
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build the string to sign
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}

	// Join the parts with &
	str := strings.Join(parts, "&")

	// Append the key
	str = str + key

	// Calculate MD5
	hash := md5.Sum([]byte(str))
	sign := hex.EncodeToString(hash[:])

	// Return lowercase sign
	return strings.ToLower(sign), nil
}

type EpaySignValidator struct {
	Key string // Merchant key for signing
}

// NewEpaySignValidator creates a new validator with the given merchant key
func NewEpaySignValidator(key string) *EpaySignValidator {
	return &EpaySignValidator{
		Key: key,
	}
}

// Validate checks if the sign in the request is valid
func (v *EpaySignValidator) Validate(i EpaySignedRequest) error {
	if v.Key == "" {
		return errors.New("merchant key is not set")
	}

	// Get the sign from the request
	providedSign := i.GetSign()
	if providedSign == "" {
		return errors.New("sign is empty")
	}

	// Calculate the expected sign
	expectedSign, err := CalculateSign(i, v.Key)
	if err != nil {
		return fmt.Errorf("failed to calculate sign: %w", err)
	}

	// Compare the signs
	if providedSign != expectedSign {
		return fmt.Errorf("invalid sign: expected %s, got %s", expectedSign, providedSign)
	}

	return nil
}
