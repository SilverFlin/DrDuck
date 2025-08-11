package cache

import (
	"os"
	"reflect"
)

// ConfigAdapter converts between different config types
type ConfigAdapter struct{}

// ConvertConfig converts any config structure to cache.CacheConfig
func (ca *ConfigAdapter) ConvertConfig(cfg interface{}) CacheConfig {
	// Check if it's already a cache.CacheConfig
	if cacheConfig, ok := cfg.(CacheConfig); ok {
		return cacheConfig
	}

	// Use reflection to extract fields from config struct
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return DefaultCacheConfig()
	}

	config := DefaultCacheConfig()

	// Extract fields by name
	if field := v.FieldByName("MaxAge"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().(int); ok {
			config.MaxAge = val
		}
	}
	if field := v.FieldByName("MaxEntries"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().(int); ok {
			config.MaxEntries = val
		}
	}
	if field := v.FieldByName("ExcludeFiles"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().([]string); ok {
			config.ExcludeFiles = val
		}
	}
	if field := v.FieldByName("IncludeFiles"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().([]string); ok {
			config.IncludeFiles = val
		}
	}
	if field := v.FieldByName("ADRDirs"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().([]string); ok {
			config.ADRDirs = val
		}
	}
	if field := v.FieldByName("CleanupAfter"); field.IsValid() && field.CanInterface() {
		if val, ok := field.Interface().(int); ok {
			config.CleanupAfter = val
		}
	}

	return config
}

// NewManagerFromMainConfig creates a cache manager from main config
func NewManagerFromMainConfig(cfg interface{}) *Manager {
	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		// Fallback to current directory
		workingDir = "."
	}

	adapter := &ConfigAdapter{}
	cacheConfig := adapter.ConvertConfig(cfg)
	
	return NewManager(workingDir, cacheConfig)
}