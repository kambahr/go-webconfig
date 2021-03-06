package webconfig

const (
	CondHTTPSvc_Header      = "header"
	CondHTTPSvc_IPAddress   = "ip-address"
	CondHTTPSvc_QueryString = "query-string"
)

// ConditionalHTTPService serves an HTTP request according to a
// condition based on a value in the header, query string or
// ip address.
type ConditionalHTTPService struct {

	// RuleType can be: header, query-string, or ip-address
	RuleType string `json:"rule-type"`

	// URLPath is the relative URL of the request; i.e. /robot.txt
	URLPath string `json:"url-path"`

	// ServeOnlyToCriteria are the related strings to match. i.e.,
	// bingbot/2.0; +http://www.bing.com/bingbot.htm and
	// Googlebot/2.1; +http://www.google.com/bot.html can be
	// two string elements in the Vaues array that will be used to
	// match in the Header (User-Agent). For example, if the target url
	// is /robot.txt; it will only be served to google and bing bots.
	ServeOnlyToCriteria []string `json:"serve-only-to-criteria"`

	// HTTPStatusCode is http status code that will be retured, if
	// a match is found. The default is 404 (not found).
	HTTPStatusCode int `json:"http-status-code"`
}
type messageBanner struct {
	On               bool `json:"on"`
	SecondsToDisplay int  `json:"seconds-to-display"`

	// When the value of On changes from false to true
	// Tickout is set to SecondsToDisplay and then
	// decremented every second until the banner is closed
	// (on set to false).
	TickCount int `json:"tick-count"`
}

type httpx struct {
	AllowedMethods []string `json:"allowed-methods"`
}

// tlsFiles defines the location of the certificate and
// private-key files. Both files must be
// in PEM format.
type tlsFiles struct {
	CertFilePath string `json:"cert-file-path"`
	KeyFilePath  string `json:"key-file-path"`
}

type urlPaths struct {
	Restrict    []string                 `json:"restrict"`
	Forward     []string                 `json:"forward"`
	Exclude     []string                 `json:"exclude"`
	ServeOnlyTo []ConditionalHTTPService `json:"conditional-http-service"`
}

// admin defines the IP addresses
// of machines that connect to the server.
// The admin pages (or app), whether hosted inside
// the public site or in a separate environment
// should be secured as such that it would be only
// available via the localhost (i.e. on the local
// machine or ssh tunnel) or a list of recognized
// IP addresses.
type admin struct {
	RunOnStartup bool     `json:"run-on-sartup"`
	PortNo       uint     `json:"port-no"`
	AllowedIP    []string `json:"allowed-ip"`
}

// siteStats holds the basic stat that can be
// set in connstat. Every server can create one,
// and handle connections accordingly; the active,
// idel, and new states must be set and undated by
// the web server.
type siteStats struct {
	Active uint `json:"active"`
	Idle   uint `json:"idle"`
	New    uint `json:"new"`
}

type site struct {
	HostName           string   `json:"hostname"`
	AlternateHostNames []string `json:"alternate-host-names"`
	Proto              string   `json:"proto"`
	PortNo             int      `json:"portno"`
}

// Config is defines the fields that are typically required for
// web configuration.  All config values have to have a
// presentation in this struct.
type Config struct {
	refreshRate        uint      // in seconds
	WebRootPath        string    `json:"web-rootp-path"`
	AppDataPath        string    `json:"appdata-path"`
	ConnStat           siteStats `json:"conn-stat"`
	ConfigFilePath     string    `json:"config-file-path"`
	ConfigFileLastHash string    `json:"config-file-last-hash"`
	Admin              admin     `json:"admin"`
	HTTP               httpx     `json:"http"`
	Site               site      `json:"site"`
	URLPaths           urlPaths  `json:"url-paths"`

	// This is to prevent "HTTP Host header attack",
	// however, it may cause problems if the webiste is
	// running behind a load-balancer and/or a security
	// program that monitors http requests.
	ValidateRemoteHost bool

	// These are the offender IP addr. Their connections
	// are drop immediately, without any message returned to them.
	BlockedIP []string `json:"blocked-ip"`

	RedirectHTTPtoHTTPS bool `json:"redirect-http-to-https"`

	MaintenanceWindowOn bool `json:"maintenance-windowon"`

	MessageBanner messageBanner `json:"messagebanner"`

	//-------------------------------------------------
	// Delete these in version > v1.0.3 + 3
	// HostName string `json:"hostname"`
	// Proto    string `json:"proto"`
	// PortNo   int    `json:"portno"`
	// [2021-09-24] Removed; moved to the Site type.
	//-------------------------------------------------

	TLS  tlsFiles          `json:"tls"`
	Data map[string]string `json:"data"`
}

const (
	cfgTemplateAll string = `
# ------------------------------------------------------------------
# About this config
#   --delimiter between key and value is space (one or many).
#   --comments must begin with #.
#   --one key/value per line; to continue to another line, 
#     place a backslash (\) at the end of the statement.
#   --The headers and keys are case- insensitive, but the
#     following is the recommanded format:
#        SomeHeaderName
#           key_1  value_1
#     All entries can appear with or without a header.      
#    
#   Some of the config values are directly related to implementation
#   of feature within a Go website. Please, see the following template
#   for implementation of these feature.
#   https://github.com/kambahr/go-webstandard
# ----------------------------------------------------------------------

# This is the hostname that will be accessed from the outside
# i.e. mydomain.com. It may still be localhost if the website is
# on a load balancer.
Site
	hostname         localhost
    # alternate-hostnames can be a local or public hostname.
	# e.g. myhost can be a local hostname, and mydomain.com 
	# can be a public host name.
	alternate-hostnames 
	portno           8085
	proto            http

# location of certificate and private files;
# both in the PEM format and must be full path.
# The paths can be an local paths; but 
# /appdata/certs/<domain name>/ is recommended.
TLS 
   #cert /usr/local/mywebapp dir/appdata/certs/mydomain/certx.pem
   #key /usr/local/webapp dir/appdata/certs/mydomain/keyx.pem

# This will affect the entire site; used for times that the whole
# site needs to be worked on. Your app will have to response 
# to requests (and display a maint-page) accordingly.
maintenance-window     off
 
Admin
   # List of IP address that will be allowed to access 
   # the admin website; separated by comma; otherwise, the admin
   # section of the website will only be served to the local machine.
   allowed-ip-addr	<ip add 1>, <ip add 2>
   run-on-startup	yes          
   portno			30000

# This message shows up on every request (page).
# Users can dismiss the banner; their option is save in 
# a cookie (named banner) so that they don't keep seeing it after reading.
# The message html template file is in /appdata/banner-msg.html. 
# Modify the "Your message goes here." with your own html/text.
MessageBanner
   # Expected values: on/off.
   # Web server should react to this value (on/off) to place and remove
   # the banner from the end-user request.
   display-mode      off

   # The banner message can be displayed for the below indicated value;
   # and then be automatically disabled (display-mode => off), when this 
   # period elapses:
   #    display-mode will be set to off from on.
   # A of value > 0 will trigger the auto-timeout. So, if you would like
   # to clear the banner message manually keep this value set to zero.
   seconds-to-display  0

# URL paths can be restricted, excluded and forwarded explicitly; the end-user 
# will receive the appropriate error message.
# These option to make a portion of your site unavailable for maintenance
# or other reasons. Each path must begin with a slash (relative path).
# The following should be the order or evaluation: 
# restrict-paths, exclude-paths, forward-paths, conditional-http-service.
URLPaths
   # restrict-paths <url paths separated by comma>
   # e.g.
   # restrict-paths   /gallery,/accounting, /customer-review, /myblog
   restrict-paths

   # with the exclude option, files will be intact in the same location, but not
   # served; the end-user will receive 404 error; or the behaviour can be 
   # customized.
   # exclude-paths /galleryx, /someotherpath
   exclude-paths

   # forward-paths <url-from|url-to-forward paths separated by comma>.
   # Note that forwarding to a fully qualified url must not be allowed.
   # e.g.   
   # forward-paths  /along-name-of-a-blog-page|/latest-blog, /another-along-name-of-a-blog-page|/best-of-blogs, \ 
   #           /and-more-and-more-pages|/yet-the-best-blog
   forward-paths

   # Conditional HTTP Service serves an HTTP request according to a
   # condition based on a value in the header, query string or
   # ip address. If a matching string is found the request will be
   # e.g. the following only allows the bing and google bots see the robot.txt file.
   # conditional-http-service [{"rule-type":"header","url-path":"/robot.txt","serve-only-to-criteria":["+http://www.bing.com/bingbot.htm","+http://www.google.com/bot.html"],"http-status-code":404}]
   conditional-http-service

# HEAD and GET are generally allowed by default.
HTTP
   allowed-methods     GET, OPTIONS, CONNECT, HEAD

# This section holds user-data. The following is the format.
# Key...... no spaces
# Value.... can include spaces.
# The data value can be any text (hex, JSON, text, xml,..).
# The delimiter is a new-line.
#
# Examples:
#     my-postgresql-conn-str  User ID=root;Password=pwd;Host=localhost;Port=5432;Database=mydb;Pooling=false;
#     my-json-value           {"mylist":["v1","v2"]}
#     my-hex-value            68656c6c6f206f75742074686572652e206775697461722069732074686520736f6e67
Data
   
`
	cnfTemplateBlockedIP string = `
# ip addresses in this file will be blocked from connecting the website.
# the following is the format:
# <ip address><minimum of one space><description>
# Example:
# 10.12.3.4 <a short description of the reason>
`
)
