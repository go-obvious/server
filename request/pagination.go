package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultLimit   = 1000
	DefaultCursor  = ""
	ParamCursor    = "cursor"
	ParamLimit     = "limit"
	ParamSort      = "sort"
	ParamSortField = "sortField"
)

type PaginationOptions struct {
	Limit      int
	Cursor     string
	Descending bool
	SortField  *string
}

// Forward/Backward cursor
type Cursor struct {
	Prev *string `json:"prev"`
	Next *string `json:"next"`
}

// GetPagingOpts extracts pagination options from the HTTP request.
func GetPagingOpts(r *http.Request) PaginationOptions {
	sortField := ""
	opts := PaginationOptions{
		Limit:      DefaultLimit,
		Cursor:     DefaultCursor,
		Descending: false,
		SortField:  &sortField,
	}
	if r == nil {
		return opts
	}
	opts.Cursor = getQueryParam(r, ParamCursor, DefaultCursor)
	opts.Limit = getQueryParamAsInt(r, ParamLimit, DefaultLimit)
	opts.Descending = strings.ToLower(getQueryParam(r, ParamSort, "")) == "desc"
	sortField = getQueryParam(r, ParamSortField, "")
	return opts
}

func getQueryParam(r *http.Request, param, defaultValue string) string {
	if v := QS(r, param); v != "" {
		return v
	}
	return defaultValue
}

func getQueryParamAsInt(r *http.Request, param string, defaultValue int) int {
	if v := QS(r, param); v != "" {
		if val, err := strconv.Atoi(v); err == nil && val > 0 {
			return val
		}
	}
	return defaultValue
}

// BuildLinkHeaders adds pagination Link headers to the HTTP response.
func BuildLinkHeaders(r *http.Request, w http.ResponseWriter, serverURLWithProtocol, path string, cursor Cursor) error {
	serverURL, err := url.Parse(serverURLWithProtocol)
	if err != nil {
		return err
	}
	queryParams := r.URL.Query()
	queryParams.Del(ParamCursor)
	queryString := queryParams.Encode()

	if cursor.Prev != nil {
		addLinkHeader(w, buildLinkHeader(serverURL, path, *cursor.Prev, queryString, "prev"))
	}
	if cursor.Next != nil {
		addLinkHeader(w, buildLinkHeader(serverURL, path, *cursor.Next, queryString, "next"))
	}
	return nil
}

func buildLinkHeader(serverURL *url.URL, path, cursor, queryString, rel string) string {
	linkURL := &url.URL{
		Scheme:   serverURL.Scheme,
		Host:     serverURL.Host,
		Path:     path,
		RawQuery: fmt.Sprintf("cursor=%s&%s", url.PathEscape(cursor), queryString),
	}
	return fmt.Sprintf("<%s>; rel=\"%s\"", linkURL.String(), rel)
}

// addLinkHeader appends a Link header to the HTTP response.
func addLinkHeader(w http.ResponseWriter, linkHeader string) {
	existingHeaders := w.Header().Get("Link")
	if existingHeaders == "" {
		w.Header().Set("Link", linkHeader)
		return
	}
	existingHeaders = strings.Trim(existingHeaders, " ,")
	w.Header().Set("Link", existingHeaders+", "+linkHeader)
}
