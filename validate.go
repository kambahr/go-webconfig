package webconfig

import (
	"net/http"
	"strings"
)

// ValidateHTTPRequest validates client http reqest according to rules
// defined within the Config structure. It returns true, 0; if request is
// validated, and false, http-error-code; if request is not validated.
// If the forward-paths section has values, the response will be forwarded
// accordingly (if a match is found).
func (c *Config) ValidateHTTPRequest(w http.ResponseWriter, r *http.Request) (bool, int) {

	rPath := strings.ToLower(r.URL.Path)

	// Host name
	rHost := strings.ToLower(strings.Split(r.Host, ":")[0])
	if c.Site.HostName != "" && rHost != c.Site.HostName && rHost != "localhost" {
		return false, http.StatusBadGateway
	}

	// Method allowed
	failed := true
	s := r.Method
	for i := 0; i < len(c.HTTP.AllowedMethods); i++ {
		if s == c.HTTP.AllowedMethods[i] {
			failed = false
			break
		}
	}
	if failed {
		return false, http.StatusMethodNotAllowed
	}

	// This is the order: restrict-paths, exclude-path, forward-paths

	// restrict-paths
	for i := 0; i < len(c.URLPaths.Restrict); i++ {
		sl := strings.ToLower(c.URLPaths.Restrict[i])
		// On-error  the err-text is placed instead of the
		// expected value starting with: ~@error
		if strings.HasPrefix(sl, "~@error") {
			continue
		}
		if sl == rPath {
			return false, http.StatusUnauthorized
		}
	}

	// exclude-path
	for i := 0; i < len(c.URLPaths.Exclude); i++ {
		sl := strings.ToLower(c.URLPaths.Exclude[i])
		if strings.HasPrefix(sl, "~@error") {
			continue
		}
		if sl == rPath {
			return false, http.StatusNotFound
		}
	}

	// forward-paths
	for i := 0; i < len(c.URLPaths.Forward); i++ {
		sl := strings.ToLower(c.URLPaths.Forward[i])
		v := strings.Split(sl, "|")
		left := ""
		right := ""
		if len(v) > 1 {
			left = v[0]
			right = v[1]
		}
		if right == "" || strings.HasPrefix(right, "~@error") {
			continue
		}
		if left == rPath {
			http.Redirect(w, r, right, http.StatusTemporaryRedirect)
			return false, http.StatusTemporaryRedirect
		}
	}

	return true, 0
}
