package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// CanonicalHash produces a deterministic SHA-256 hash of a typed request payload.
// Keys are sorted recursively for determinism.
func CanonicalHash(payload map[string]any) string {
	normalized := normalizeValue(payload)
	data, err := json.Marshal(normalized)
	if err != nil {
		// Fallback: should never happen with valid map[string]any
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// normalizeValue recursively sorts map keys and converts slices/numbers
// to deterministic JSON representations.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return normalizeMap(val)
	case []any:
		return normalizeSlice(val)
	case float64:
		// JSON unmarshals all numbers as float64
		return ensureNumberType(val)
	case nil, string, bool:
		return v
	default:
		return v
	}
}

// normalizeMap sorts map keys and normalizes values.
func normalizeMap(m map[string]any) map[string]any {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(map[string]any, len(m))
	for _, k := range keys {
		result[k] = normalizeValue(m[k])
	}
	return result
}

// normalizeSlice normalizes each element.
func normalizeSlice(s []any) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = normalizeValue(v)
	}
	return result
}

// ensureNumberType converts float64 to int64 if it's a whole number,
// ensuring consistent JSON representation.
func ensureNumberType(f float64) any {
	if f == float64(int64(f)) && f < 1<<53 && f > -(1<<53) {
		return int64(f)
	}
	return f
}
