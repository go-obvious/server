package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-obvious/server/request"
)

// helper method
func StringPtr(s string) *string {
	return &s
}

func TestBuildLinkHeaders(t *testing.T) {
	testCases := []struct {
		name             string
		serverURL        string
		path             string
		cursor           request.Cursor
		expectedLink     string
		expectedErrorMsg string
	}{
		{
			name:      "Valid Cursor - Both Prev and Next",
			serverURL: "http://localhost:8080",
			path:      "/api/users",
			cursor: request.Cursor{
				Prev: StringPtr("abcdefg"),
				Next: StringPtr("hijklmn"),
			},
			expectedLink: `<http://localhost:8080/api/users?cursor=abcdefg&>; rel="prev", <http://localhost:8080/api/users?cursor=hijklmn&>; rel="next"`,
		},
		{
			name:      "Valid Cursor - Prev Only",
			serverURL: "http://localhost:8080",
			path:      "/api/users",
			cursor: request.Cursor{
				Prev: StringPtr("abcdefg"),
			},
			expectedLink: `<http://localhost:8080/api/users?cursor=abcdefg&>; rel="prev"`,
		},
		{
			name:      "Valid Cursor - Next Only",
			serverURL: "http://localhost:8080",
			path:      "/api/users",
			cursor: request.Cursor{
				Next: StringPtr("hijklmn"),
			},
			expectedLink: `<http://localhost:8080/api/users?cursor=hijklmn&>; rel="next"`,
		},
		{
			name:             "Empty Cursor",
			serverURL:        "http://localhost:8080",
			path:             "/api/users",
			cursor:           request.Cursor{},
			expectedLink:     "",
			expectedErrorMsg: "No cursor provided",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com/foo", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()

			err = request.BuildLinkHeaders(req, rr, tc.serverURL, tc.path, tc.cursor)

			if err != nil {
				if tc.expectedErrorMsg == "" {
					t.Fatalf("Unexpected error: %v", err)
				}
				if err.Error() != tc.expectedErrorMsg {
					t.Errorf("Unexpected error message. Expected: %s, Got: %s", tc.expectedErrorMsg, err.Error())
				}
			}

			linkHeaders := rr.Header().Get("Link")

			if linkHeaders != tc.expectedLink {
				t.Errorf("Unexpected Link headers. Expected: %s, Got: %s", tc.expectedLink, linkHeaders)
			}
		})
	}
}
