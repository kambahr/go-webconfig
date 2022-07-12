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
	if c.ValidateRemoteHost {
		rHost := strings.ToLower(strings.Split(r.Host, ":")[0])

		// Do not exclude loadhost or loopback ip addr; as the remote host can be
		// altered as anything ["HTTP Host header attack"]:
		// For example,
		//    the following could go through with no problems, if not dealt with:
		//    curl -X GET -H "Host:127.0.0.1" "https://your-good-domain-name.com/"
		if c.Site.HostName != "" && rHost != c.Site.HostName /*&& rHost != "localhost" && rHost != "127.0.0.1"*/ {

			// Also check the alternate host names
			ok := false
			for i := 0; i < len(c.Site.AlternateHostNames); i++ {
				if c.Site.AlternateHostNames[i] == rHost {
					ok = true
					break
				}
			}

			if !ok {
				return false, http.StatusBadGateway
			}
		}
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

	// This is the order of: restrict-paths, exclude-path, forward-paths, conditional-http-service

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

	// conditional-http-service
	ip := strings.Split(strings.Replace(strings.Replace(r.RemoteAddr, "[", "", -1), "]", "", -1), ":")[0]
	qs := r.URL.RawQuery

	for i := 0; i < len(c.URLPaths.ServeOnlyTo); i++ {
		if rPath == c.URLPaths.ServeOnlyTo[i].URLPath {
			// check headers
			if c.URLPaths.ServeOnlyTo[i].RuleType == CondHTTPSvc_Header {
				for _, v := range r.Header {
					for j := 0; j < len(c.URLPaths.ServeOnlyTo[i].ServeOnlyToCriteria); j++ {
						m := c.URLPaths.ServeOnlyTo[i].ServeOnlyToCriteria[j]
						for k := 0; k < len(v); k++ {
							if strings.Contains(v[k], m) {
								// The caller can view the page - as its request header
								// has a value that matches the ServerOnlyTo critiera
								return true, 0
							}
						}
					}
				}
			}

			// check IP address and query string
			for j := 0; j < len(c.URLPaths.ServeOnlyTo[i].ServeOnlyToCriteria); j++ {
				if c.URLPaths.ServeOnlyTo[i].ServeOnlyToCriteria[j] == ip || strings.Contains(qs, c.URLPaths.ServeOnlyTo[i].ServeOnlyToCriteria[j]) {
					// The caller can view the page - as its request header
					// has a value that matches the ServerOnlyTo critiera
					return true, 0
				}
			}

			// If we get here, it means that:
			//   1. there is a rule for the current r.Path
			//   2. there is no match found to allow the client to recieve the content
			// so, return error
			errCode := c.URLPaths.ServeOnlyTo[i].HTTPStatusCode

			if errCode < 1 {
				errCode = 404 // default error code
			}

			return false, errCode
		}
	}

	return true, 0
}
