# Web Configuration for Go websites

## Manage website settings with webconfig
go-webconfig is a free-syle text-based config helper with built-in type and events.

### Usage
Create an instance and start using the Config type.

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
- Conditional HTTP Service based on ip address, header, and query string.
The following example allows only bing and google bots to see /robot.txt:

conditional-http-service [{"rule-type":"ip-address","url-path":"/robot.txt","serve-only-to-criteria":["+http://www.bing.com/bingbot.htm","+http://www.google.com/bot.html"],"http-status-code":404}]]

See *ConditionalHTTPService* and *conditional-http-service* in defs.go

- Built-in Request Validation; usage example:
``` go
isRequestValid, httpErrCode := Config.ValidateHTTPRequest(w, r)

if httpErrCode == http.StatusTemporaryRedirect || httpErrCode == http.StatusBadGateway {
    return
} else {
    // deal with the request according to the http error code
}
```
