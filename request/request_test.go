package request_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-obvious/server/request"
)

func TestParam(t *testing.T) {
	// Create a chi router with a route that has URL parameters
	r := chi.NewRouter()
	var capturedParam string
	r.Get("/users/{userID}/posts/{postID}", func(w http.ResponseWriter, req *http.Request) {
		capturedParam = request.Param(req, "userID")
		w.WriteHeader(http.StatusOK)
	})

	// Test valid parameter extraction
	req, err := http.NewRequest("GET", "/users/123/posts/456", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, "123", capturedParam)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestQS(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		paramName      string
		expectedValue  string
	}{
		{
			name:          "Valid query parameter",
			url:           "/test?name=john&age=30",
			paramName:     "name",
			expectedValue: "john",
		},
		{
			name:          "Multiple values - returns first",
			url:           "/test?tag=golang&tag=testing",
			paramName:     "tag",
			expectedValue: "golang",
		},
		{
			name:          "Non-existent parameter",
			url:           "/test?name=john",
			paramName:     "missing",
			expectedValue: "",
		},
		{
			name:          "Empty parameter value",
			url:           "/test?name=&age=30",
			paramName:     "name",
			expectedValue: "",
		},
		{
			name:          "URL encoded parameter",
			url:           "/test?message=hello%20world",
			paramName:     "message",
			expectedValue: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			result := request.QS(req, tt.paramName)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestQSAll(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		paramName      string
		expectedValues []string
	}{
		{
			name:           "Multiple values",
			url:            "/test?tag=golang&tag=testing&tag=web",
			paramName:      "tag",
			expectedValues: []string{"golang", "testing", "web"},
		},
		{
			name:           "Single value",
			url:            "/test?name=john",
			paramName:      "name",
			expectedValues: []string{"john"},
		},
		{
			name:           "Non-existent parameter",
			url:            "/test?name=john",
			paramName:      "missing",
			expectedValues: nil,
		},
		{
			name:           "Empty parameter",
			url:            "/test?name=",
			paramName:      "name",
			expectedValues: []string{""},
		},
		{
			name:           "Mix of empty and non-empty",
			url:            "/test?tag=&tag=golang&tag=",
			paramName:      "tag",
			expectedValues: []string{"", "golang", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			result := request.QSAll(req, tt.paramName)
			assert.Equal(t, tt.expectedValues, result)
		})
	}
}

func TestQSDefault(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		paramName      string
		defaultValue   string
		expectedValue  string
	}{
		{
			name:          "Parameter exists",
			url:           "/test?limit=50",
			paramName:     "limit",
			defaultValue:  "10",
			expectedValue: "50",
		},
		{
			name:          "Parameter missing - use default",
			url:           "/test?other=value",
			paramName:     "limit",
			defaultValue:  "10",
			expectedValue: "10",
		},
		{
			name:          "Parameter empty - use default",
			url:           "/test?limit=",
			paramName:     "limit",
			defaultValue:  "10",
			expectedValue: "10",
		},
		{
			name:          "Parameter with whitespace",
			url:           "/test?limit=%20%20%20",
			paramName:     "limit",
			defaultValue:  "10",
			expectedValue: "   ",
		},
		{
			name:          "Default value is empty",
			url:           "/test?other=value",
			paramName:     "limit",
			defaultValue:  "",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)

			result := request.QSDefault(req, tt.paramName, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestGetBody(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	tests := []struct {
		name        string
		body        string
		expectError bool
		errorMsg    string
		expected    TestStruct
	}{
		{
			name:        "Valid JSON",
			body:        `{"name":"john","age":30,"email":"john@example.com"}`,
			expectError: false,
			expected:    TestStruct{Name: "john", Age: 30, Email: "john@example.com"},
		},
		{
			name:        "Empty body",
			body:        "",
			expectError: true,
			errorMsg:    "request body must not be empty",
		},
		{
			name:        "Invalid JSON syntax",
			body:        `{"name":"john","age":30,}`,
			expectError: true,
			errorMsg:    "request body contains badly-formed JSON",
		},
		{
			name:        "Invalid type",
			body:        `{"name":"john","age":"thirty","email":"john@example.com"}`,
			expectError: true,
			errorMsg:    "request body contains an invalid value for the \"age\" field",
		},
		{
			name:        "Unknown field",
			body:        `{"name":"john","age":30,"unknown_field":"value"}`,
			expectError: false, // Go's json.Unmarshal ignores unknown fields by default
			expected:    TestStruct{Name: "john", Age: 30, Email: ""},
		},
		{
			name:        "Truncated JSON",
			body:        `{"name":"john","age":30`,
			expectError: true,
			errorMsg:    "request body contains badly-formed JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			var result TestStruct

			err = request.GetBody(rr, req, &result)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetBody_MaxBodySize(t *testing.T) {
	// Test that MaxBytesReader is being used (integration test)
	// The specific behavior may vary based on how the reader is consumed
	t.Run("MaxBytesReader integration", func(t *testing.T) {
		// Create a body around the limit
		largeValue := strings.Repeat("a", request.MaxBodySize/2)
		largeBody := `{"large_field":"` + largeValue + `"}`
		
		req, err := http.NewRequest("POST", "/test", strings.NewReader(largeBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		var result map[string]interface{}

		// This should succeed since we're under the limit
		err = request.GetBody(rr, req, &result)
		require.NoError(t, err)
		assert.Equal(t, largeValue, result["large_field"])
	})
}

func TestGetBody_PointerTypes(t *testing.T) {
	// Test with different pointer types to ensure compatibility
	tests := []struct {
		name string
		body string
		dest interface{}
	}{
		{
			name: "Map pointer",
			body: `{"key":"value"}`,
			dest: &map[string]interface{}{},
		},
		{
			name: "Slice pointer",
			body: `["item1","item2"]`,
			dest: &[]string{},
		},
		{
			name: "String pointer",
			body: `"hello world"`,
			dest: new(string),
		},
		{
			name: "Number pointer",
			body: `42`,
			dest: new(int),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			err = request.GetBody(rr, req, tt.dest)
			require.NoError(t, err)
		})
	}
}

func TestHandleJSONDecodeError_AllErrorTypes(t *testing.T) {
	// Test different JSON decode error scenarios through GetBody
	tests := []struct {
		name     string
		body     string
		errorMsg string
	}{
		{
			name:     "Syntax error",
			body:     `{"name": "john",}`,
			errorMsg: "request body contains badly-formed JSON (at position",
		},
		{
			name:     "Unexpected EOF",
			body:     `{"name": "john"`,
			errorMsg: "request body contains badly-formed JSON",
		},
		{
			name:     "Type mismatch",
			body:     `{"age": "not-a-number"}`,
			errorMsg: "request body contains an invalid value for the \"age\" field",
		},
		{
			name:     "EOF error",
			body:     ``,
			errorMsg: "request body must not be empty",
		},
	}

	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/test", strings.NewReader(tt.body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			var result TestStruct

			err = request.GetBody(rr, req, &result)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestGetBody_EdgeCases(t *testing.T) {
	t.Run("Nil destination", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/test", strings.NewReader(`{"name":"john"}`))
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		// This should not panic but will likely error
		err = request.GetBody(rr, req, nil)
		// We expect some kind of error with nil destination
		require.Error(t, err)
	})

	t.Run("Multiple reads from same body", func(t *testing.T) {
		body := `{"name":"john"}`
		req, err := http.NewRequest("POST", "/test", strings.NewReader(body))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		var result1, result2 map[string]interface{}

		// First read should succeed
		err = request.GetBody(rr, req, &result1)
		require.NoError(t, err)

		// Second read should fail (body already consumed)
		err = request.GetBody(rr, req, &result2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request body must not be empty")
	})
}