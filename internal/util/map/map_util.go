package maputil

import (
	"reflect"
)

// MapUtil provides utility functions for map operations
type MapUtil struct{}

// NewMapUtil creates a new MapUtil
func NewMapUtil() *MapUtil {
	return &MapUtil{}
}

// Get gets a value from a map
func (m *MapUtil) Get(data map[string]interface{}, key string) (interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return value, true
}

// Set sets a value in a map
func (m *MapUtil) Set(data map[string]interface{}, key string, value interface{}) {
	data[key] = value
}

// Delete deletes a value from a map
func (m *MapUtil) Delete(data map[string]interface{}, key string) {
	delete(data, key)
}

// Has checks if a key exists in a map
func (m *MapUtil) Has(data map[string]interface{}, key string) bool {
	_, exists := data[key]
	return exists
}

// Keys gets all keys from a map
func (m *MapUtil) Keys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// Values gets all values from a map
func (m *MapUtil) Values(data map[string]interface{}) []interface{} {
	values := make([]interface{}, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// Merge merges multiple maps into one
func (m *MapUtil) Merge(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, data := range maps {
		for k, v := range data {
			result[k] = v
		}
	}
	return result
}

// DeepMerge merges multiple maps recursively
func (m *MapUtil) DeepMerge(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, data := range maps {
		for k, v := range data {
			if existing, exists := result[k]; exists {
				if m.isMap(existing) && m.isMap(v) {
					result[k] = m.DeepMerge(existing.(map[string]interface{}), v.(map[string]interface{}))
				} else {
					result[k] = v
				}
			} else {
				result[k] = v
			}
		}
	}
	return result
}

// Filter filters a map based on a predicate
func (m *MapUtil) Filter(data map[string]interface{}, predicate func(key string, value interface{}) bool) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// Map transforms a map using a mapper function
func (m *MapUtil) Map(data map[string]interface{}, mapper func(key string, value interface{}) (string, interface{})) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		newKey, newValue := mapper(k, v)
		result[newKey] = newValue
	}
	return result
}

// Reduce reduces a map to a single value
func (m *MapUtil) Reduce(data map[string]interface{}, initial interface{}, reducer func(accumulator interface{}, key string, value interface{}) interface{}) interface{} {
	result := initial
	for k, v := range data {
		result = reducer(result, k, v)
	}
	return result
}

// ForEach iterates over a map
func (m *MapUtil) ForEach(data map[string]interface{}, iterator func(key string, value interface{})) {
	for k, v := range data {
		iterator(k, v)
	}
}

// Clone creates a shallow copy of a map
func (m *MapUtil) Clone(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		result[k] = v
	}
	return result
}

// DeepClone creates a deep copy of a map
func (m *MapUtil) DeepClone(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(data))
	for k, v := range data {
		if m.isMap(v) {
			result[k] = m.DeepClone(v.(map[string]interface{}))
		} else {
			result[k] = v
		}
	}
	return result
}

// isMap checks if a value is a map
func (m *MapUtil) isMap(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

// IsEmpty checks if a map is empty
func (m *MapUtil) IsEmpty(data map[string]interface{}) bool {
	return len(data) == 0
}

// Size gets the size of a map
func (m *MapUtil) Size(data map[string]interface{}) int {
	return len(data)
}

// Clear clears a map
func (m *MapUtil) Clear(data map[string]interface{}) {
	for k := range data {
		delete(data, k)
	}
}

// GetOrDefault gets a value from a map or returns a default value
func (m *MapUtil) GetOrDefault(data map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, exists := m.Get(data, key); exists {
		return value
	}
	return defaultValue
}

// PutIfAbsent puts a value in a map if the key is not present
func (m *MapUtil) PutIfAbsent(data map[string]interface{}, key string, value interface{}) interface{} {
	if existing, exists := m.Get(data, key); exists {
		return existing
	}
	data[key] = value
	return value
}

// Remove removes a value from a map and returns it
func (m *MapUtil) Remove(data map[string]interface{}, key string) (interface{}, bool) {
	if value, exists := m.Get(data, key); exists {
		delete(data, key)
		return value, true
	}
	return nil, false
}

// Replace replaces a value in a map if the key exists
func (m *MapUtil) Replace(data map[string]interface{}, key string, value interface{}) (interface{}, bool) {
	if oldValue, exists := m.Get(data, key); exists {
		data[key] = value
		return oldValue, true
	}
	return nil, false
}

// ReplaceIfPresent replaces a value in a map if the key exists and the old value matches
func (m *MapUtil) ReplaceIfPresent(data map[string]interface{}, key string, oldValue, newValue interface{}) bool {
	if current, exists := m.Get(data, key); exists && reflect.DeepEqual(current, oldValue) {
		data[key] = newValue
		return true
	}
	return false
}
