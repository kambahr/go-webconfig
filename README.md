# Web Configuration for Go websites

## Manage website settings with webconfig
go-webconfig is a free-syle text-based config helper with built-in type and events.

### Usage
Create an instance and starting using the Config type.

``` go
	Config *webconfig.Config
	Config = webconfig.NewWebConfig(<websites' full root path>)
```

### Features
- Free-style text based: view/read naturally.
- Use comments throughout the config file.
- Integrates with a Go website project; initializes with only a web root-path
  *  One type (Config) holds all webconfig data.
- It keeps up with changes frequently; no need to restart the webserver to get a refreshed config data.
- Common web settings + security, and URL management options.
- Keeps a separate file for blocked IP addresses. 
- Built-in timeout event to reset the Message Banner display value to off.

### Code Usage Examples

#### Managing HTTP Requests

``` go
	rPath := strings.ToLower(r.URL.Path)

	// Method allowed
	failed := true
	s := r.Method
	for i := 0; i < len(e.Config.HTTP.AllowedMethods); i++ {
		if s == e.Config.HTTP.AllowedMethods[i] {
			failed = false
			break
		}
	}
	if failed {
		e.displayError(w, 405)
		return false
	}

	// This is the order: restrict-paths, exclude-path, forward-paths

	// restrict-paths
	for i := 0; i < len(e.Config.URLPaths.Restrict); i++ {
		sl := strings.ToLower(e.Config.URLPaths.Restrict[i])
		// On-error  the err-text is placed instead of the
		// expected value starting with: ~@error
		if strings.HasPrefix(sl, "~@error") {
			continue
		}
		if sl == rPath {
			e.displayError(w, http.StatusUnauthorized)
			return false
		}
	}

	// exclude-path
	for i := 0; i < len(e.Config.URLPaths.Exclude); i++ {
		sl := strings.ToLower(e.Config.URLPaths.Exclude[i])
		if strings.HasPrefix(sl, "~@error") {
			continue
		}
		if sl == rPath {
			e.displayError(w, http.StatusNotFound)
			return false
		}
	}

	// forward-paths
	for i := 0; i < len(e.Config.URLPaths.Forward); i++ {
		sl := strings.ToLower(e.Config.URLPaths.Forward[i])
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
			return false
		}
	}

```

#### Managing Offenders

``` go
// connState enables you to monitor callers before their
// requests get to the handlers. You can use this for security
// performance enhancements, or monitoring (i.e. number of
// active connections).
func (e *Environment) connState(conn net.Conn, connState http.ConnState) {

	// Check the blocked ips
	ip := strings.Split(strings.Replace(strings.Replace(conn.RemoteAddr().String(), "[", "", -1), "]", "", -1), ":")[0]

	// blank means its ::1 (ipv6 loopback ip)
	if ip != "" {
		for i := 0; i < len(e.Config.BlockedIP); i++ {
			if e.Config.BlockedIP[i] == ip {
			        // Drop the connection; it ends here and it 
				// will not get to the http handlers.
				conn.Close()
				return
			}
		}
	}
}

```

