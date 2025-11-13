package cli

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	wCli "github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
)

// HTTPInput implements cli.Input interface for HTTP requests
type HTTPInput struct {
	ctx      context.Context
	request  *http.Request
	values   map[string]interface{}
	stdout   io.Writer
	stderr   io.Writer
}

// NewHTTPInput creates a new HTTPInput from an HTTP request
func NewHTTPInput(r *http.Request, stdout, stderr io.Writer) wCli.Input {
	values := make(map[string]interface{})

	// Parse query parameters
	for key, vals := range r.URL.Query() {
		if len(vals) > 0 {
			if len(vals) == 1 {
				values[key] = vals[0]
			} else {
				values[key] = vals
			}
		}
	}

	// Parse form data if Content-Type is application/x-www-form-urlencoded or multipart/form-data
	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		if err := r.ParseForm(); err == nil {
			for key, vals := range r.PostForm {
				if len(vals) > 0 {
					if len(vals) == 1 {
						values[key] = vals[0]
					} else {
						values[key] = vals
					}
				}
			}
		}
	}

	return &HTTPInput{
		ctx:     r.Context(),
		request: r,
		values:  values,
		stdout:  stdout,
		stderr:  stderr,
	}
}

// SetValue sets a value for a given key
func (h *HTTPInput) SetValue(key string, value interface{}) {
	h.values[key] = value
}

// Context returns the request context
func (h *HTTPInput) Context() context.Context {
	return h.ctx
}

// IsSet checks whether a flag has been set
func (h *HTTPInput) IsSet(key string) bool {
	_, ok := h.values[key]
	return ok
}

// Bool returns boolean value of given flag
func (h *HTTPInput) Bool(key string) bool {
	val, ok := h.values[key]
	if !ok {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)
		return b
	default:
		return false
	}
}

// String returns string value of given flag
func (h *HTTPInput) String(key string) string {
	val, ok := h.values[key]
	if !ok {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []string:
		if len(v) > 0 {
			return v[0]
		}
		return ""
	default:
		return ""
	}
}

// StringSlice returns string slice of given flag
func (h *HTTPInput) StringSlice(key string) []string {
	val, ok := h.values[key]
	if !ok {
		return []string{}
	}
	switch v := val.(type) {
	case []string:
		return v
	case string:
		// Handle comma-separated values
		if v == "" {
			return []string{}
		}
		return strings.Split(v, ",")
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}

// Int returns int value of given flag
func (h *HTTPInput) Int(key string) int {
	val, ok := h.values[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

// IntSlice returns int slice of given flag
func (h *HTTPInput) IntSlice(key string) []int {
	val, ok := h.values[key]
	if !ok {
		return []int{}
	}
	switch v := val.(type) {
	case []int:
		return v
	case []interface{}:
		result := make([]int, 0, len(v))
		for _, item := range v {
			switch i := item.(type) {
			case int:
				result = append(result, i)
			case float64:
				result = append(result, int(i))
			case string:
				if num, err := strconv.Atoi(i); err == nil {
					result = append(result, num)
				}
			}
		}
		return result
	case string:
		parts := strings.Split(v, ",")
		result := make([]int, 0, len(parts))
		for _, part := range parts {
			if num, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				result = append(result, num)
			}
		}
		return result
	default:
		return []int{}
	}
}

// Int64 returns int64 value of given flag
func (h *HTTPInput) Int64(key string) int64 {
	val, ok := h.values[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		i, _ := strconv.ParseInt(v, 10, 64)
		return i
	default:
		return 0
	}
}

// Int64Slice returns int64 slice of given flag
func (h *HTTPInput) Int64Slice(key string) []int64 {
	val, ok := h.values[key]
	if !ok {
		return []int64{}
	}
	switch v := val.(type) {
	case []int64:
		return v
	case []int:
		result := make([]int64, len(v))
		for i, val := range v {
			result[i] = int64(val)
		}
		return result
	case []interface{}:
		result := make([]int64, 0, len(v))
		for _, item := range v {
			switch i := item.(type) {
			case int64:
				result = append(result, i)
			case int:
				result = append(result, int64(i))
			case float64:
				result = append(result, int64(i))
			case string:
				if num, err := strconv.ParseInt(i, 10, 64); err == nil {
					result = append(result, num)
				}
			}
		}
		return result
	case string:
		parts := strings.Split(v, ",")
		result := make([]int64, 0, len(parts))
		for _, part := range parts {
			if num, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil {
				result = append(result, num)
			}
		}
		return result
	default:
		return []int64{}
	}
}

// GlobalBool returns global boolean value
func (h *HTTPInput) GlobalBool(key string) bool {
	return h.Bool(key)
}

// GlobalString returns global string value
func (h *HTTPInput) GlobalString(key string) string {
	return h.String(key)
}

// GlobalStringSlice returns global string slice
func (h *HTTPInput) GlobalStringSlice(key string) []string {
	return h.StringSlice(key)
}

// GlobalInt returns global int value
func (h *HTTPInput) GlobalInt(key string) int {
	return h.Int(key)
}

// GlobalIntSlice returns global int slice
func (h *HTTPInput) GlobalIntSlice(key string) []int {
	return h.IntSlice(key)
}

// GlobalInt64 returns global int64 value
func (h *HTTPInput) GlobalInt64(key string) int64 {
	return h.Int64(key)
}

// GlobalInt64Slice returns global int64 slice
func (h *HTTPInput) GlobalInt64Slice(key string) []int64 {
	return h.Int64Slice(key)
}

// Duration returns time duration of given flag
func (h *HTTPInput) Duration(key string) time.Duration {
	val, ok := h.values[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case time.Duration:
		return v
	case string:
		d, _ := time.ParseDuration(v)
		return d
	case int:
		return time.Duration(v) * time.Second
	case int64:
		return time.Duration(v) * time.Second
	default:
		return 0
	}
}

// Stdout returns io.Writer for stdout
func (h *HTTPInput) Stdout() io.Writer {
	if h.stdout != nil {
		return h.stdout
	}
	return io.Discard
}

// Stderr returns io.Writer for stderr
func (h *HTTPInput) Stderr() io.Writer {
	if h.stderr != nil {
		return h.stderr
	}
	return io.Discard
}

