package epay

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
)

type ParamCarrier struct {
	Pid       int
	NotifyUrl string
	Param     string
	// EpayReturnUrl string
}

func DecodeParamCarrier(s string) (*ParamCarrier, error) {
	// Decode the base64-encoded data
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding failed: %w", err)
	}

	// Create a new gob decoder that reads from the decoded data
	dec := gob.NewDecoder(bytes.NewReader(decoded))

	// Decode the data into the result
	var result ParamCarrier
	err = dec.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("gob decoding failed: %w", err)
	}

	return &result, nil
}

func (c *ParamCarrier) Encode() (string, error) {
	// Create a buffer to hold the encoded data
	var buf bytes.Buffer

	// Create a new gob encoder that writes to the buffer
	enc := gob.NewEncoder(&buf)

	// Encode the data
	err := enc.Encode(*c)
	if err != nil {
		return "", fmt.Errorf("gob encoding failed: %w", err)
	}

	// Return the encoded data as a byte slice
	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}
