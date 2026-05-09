package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignHMACSHA512_Consistency(t *testing.T) {
	payload := "REQ001:RRN001:MERCHANT001"
	key := "speskilltest"

	result1 := SignHMACSHA512(payload, key)
	result2 := SignHMACSHA512(payload, key)

	assert.Equal(t, result1, result2)
	assert.NotEmpty(t, result1)
}

func TestSignHMACSHA512_DifferentPayloadsProduceDifferentResults(t *testing.T) {
	key := "speskilltest"

	result1 := SignHMACSHA512("payload-A", key)
	result2 := SignHMACSHA512("payload-B", key)

	assert.NotEqual(t, result1, result2)
}

func TestSignHMACSHA512_DifferentKeysProduceDifferentResults(t *testing.T) {
	payload := "REQ001:RRN001:MERCHANT001"

	result1 := SignHMACSHA512(payload, "key-1")
	result2 := SignHMACSHA512(payload, "key-2")

	assert.NotEqual(t, result1, result2)
}

func TestSignHMACSHA512_ActualValue(t *testing.T) {
	payload := "REQ-001:RRN-001:008800223497"
	key := "speskilltest"

	result := SignHMACSHA512(payload, key)

	assert.NotEmpty(t, result)
	assert.NotContains(t, result, " ")
}
