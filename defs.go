package webconfig

type messageBanner struct {
	On               bool `json:"on"`
	SecondsToDisplay int  `json:"seconds-to-display"`

	// When the value of On chnages from false to true
	// Tickout is set to SecondsToDisplay and then
	// decremented every second until the banner is closed
	// (on set to false).
	TickCount int `json:"tick-count"`
}

type httpx struct {
	AllowedMethods []string `json:"allowed-methods"`
}

// tlsFiles defines the location of the certificate and
// private-key fiels. Both files must be
// in PEM format.
type tlsFiles struct {
	CertFilePath string `json:"cert-file-path"`
	KeyFilePath  string `json:"key-file-path"`
}
type urlPaths struct {
	Restrict []string `json:"restrict"`
	Forward  []string `json:"forward"`
	Exclude  []string `json:"exclude"`
}

// admin defines the ip addresses
// of machins that connect to the server.
// The admin pages (or app), whether hosted inside
// the public site or in a separete environment
// should be secured as such that it would be only
// available via the localhost (i.e. on the local
// machine or ssh tunnel) or a list of recognized
// ip addresses.
type admin struct {
	RunOnStartup bool     `json:"run-on-sartup"`
	PortNo       uint     `json:"port-no"`
	AllowedIP    []string `json:"allowed-ip"`
}

// siteStats holds the basic stat that can be
// set in connstat. Every server can create one,
// and handle connections accordingly; the active,
// idel, and new states must be set and udated by
// the web server.
type siteStats struct {
	Active uint `json:"active"`
	Idle   uint `json:"idle"`
	New    uint `json:"new"`
}

type site struct {
	HostName string `json:"hostname"`
	Proto    string `json:"proto"`
	PortNo   int    `json:"portno"`
}

// Config is defines the fields that are typicaly required for
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

	// These are the offender ip addr. Their connections
	// are drop immedietely, without any message returned to them.
	BlockedIP []string `json:"blocked-ip"`

	RedirectHTTPtoHTTPS bool `json:"redirect-http-to-https"`

	MaintenanceWindowOn bool `json:"maintenance-window0on"`

	MessageBanner messageBanner `json:"messagebanner"`

	//-------------------------------------------------
	// Delete these in version > v1.0.3 + 3
	HostName string `json:"hostname"`
	Proto    string `json:"proto"`
	PortNo   int    `json:"portno"`
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
#   --one key/valuee per line; to continue to another line, 
#     place a backslash (\) at the end of the statement.
#   --The headers and keys are case- insensitive, but the
#     following is the recommanded format:
#        SomeHeaderName
#           key_1  value_1
#     All entries can appear with or without a header.      
#    
#   Some of the config values are directely related to implementation
#   of feature within a Go website. Please, see the following template
#   for implementation of these feature.
#   https://github.com/kambahr/go-webstandard
#----------------------------------------------------------------------

# This is the hostname that will be accessed from the outside
# i.e. mydomain.com. It may still be localhost if the website is
# on a loadbalancer.
Site
	hostname         localhost
	portno           8085
	proto            http

# location of certificate and private files;
# both in the PEM format and must be full path.
# The paths can be an local paths; but 
# /appdata/certs/<domain anme>/ is recommended.
TLS 
   #cert /usr/local/mywebapp dir/appdata/certs/mydomain/certx.pem
   #key /usr/local/webapp dir/appdata/certs/mydomain/keyx.pem

# This will affect the entire site; used for times that the whole
# site needs to be worked on. Your app will have to response 
# to requests (and display a maint-page) accordingly.
maintenance-window     off
 
Admin
   # List of ip address that will be allowed to access 
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
   # and then be automatically disalbed (display-mode => off), when this 
   # period elapses:
   #    display-mode will be seto off from on.
   # A of value > 0 will trigger the auto-timeout. So, if you would like
   # to clear the banner message manually keep this value set to zero.
   seconds-to-display  0

# URL paths can be restried, exluded and forwarded explicitly; the end-user 
# will receive the appropriate error message.
# These option to make a portion of your site unavailable for maintenance
# or other reasons. Each path must begin with a slash (relative path).
# The following should be the order or evaluation: 
#       restrict-paths, exclude-paths, forward-paths.
URLPaths
   # restrict-paths <url paths separated by comma>
   # e.g.
   # restrict-paths   /gallery,/accouting, /customer-review, /myblog

   # with the exclude option, files will be intact in the same location, but not
   # served; the end-user will receive 404 error; or the behavior can be 
   # customized.
   exclude-paths /galleryx, /someotherpath

   # forward-paths <url-from:url-to-forward paths separated by comma>.
   # Note that forwarding to a fully qualified url must not be allowed.
   # e.g.   
   forward-paths  /along-name-of-a-blog-page|/latest-blog, /another-along-name-of-a-blog-page|/best-of-blogs, \ 
              /and-more-and-more-pages|/yet-the-best-blog

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
# 10.12.3.4 <a short descrription of the reason>
`
)
