package webconfig

type messageBanner struct {
	On               bool
	SecondsToDisplay int

	// When the value of On chnages from false to true
	// Tickout is set to SecondsToDisplay and then
	// decremented every second until the banner is closed
	// (on set to false).
	TickCount int
}

type httpx struct {
	AllowedMethods []string
}

// tlsFiles defines the location of the certificate and
// private-key fiels. Both files must be
// in PEM format.
type tlsFiles struct {
	CertFilePath string
	KeyFilePath  string
}
type urlPaths struct {
	Restrict []string
	Forward  []string
	Exclude  []string
}

// admin defines the ip addresses
// of machiens that connect to the server.
// The admin pages (or app), whether hosted inside
// the public site or in a separete environment
// should be secured as such that it would be only
// available via the localhost (i.e. on the local
// machine or ssh tunnel) or a list of recognized
// ip addresses.
type admin struct {
	AllowedIP []string
}

// siteStats holds the basic stat that can be
// set in connstat. Every server can create one,
// and handle connections accordingly; the active,
// idel, and new states must be set and udated by
// the web server.
type siteStats struct {
	Active uint
	Idle   uint
	New    uint
}

// Config is defines the fields that are typicaly required for
// web configuration.  All config values have to have a
// presentation in this struct.
type Config struct {
	refreshRate        uint // in seconds
	WebRootPath        string
	AppDataPath        string
	ConnStat           siteStats
	ConfigFilePath     string
	ConfigFileLastHash string

	HTTP httpx

	URLPaths urlPaths

	// These are the offender ip addr. Their connections
	// are drop immedietely, without any message returned to them.
	BlockedIP []string

	// Admin pages are only accessable from the local machine,
	// unless the ip of the remote machine is added to this array.
	//
	Admin admin

	RedirectHTTPtoHTTPS bool

	MaintenanceWindowOn bool

	MessageBanner messageBanner
	HostName      string
	Proto         string
	PortNo        int
	TLS           tlsFiles
}

const (
	cfgTemplateAll string = `
# ------------------------------------------------------------------
# About this config
#   --delimiter between key and value is space (one or many).
#   --comment must be at the begining of the line with a #.
#   --one key/valuee per line.
#   --keys are in lowercase.
#   --to continue to another line, place a backslash (\) at the end 
#      of the statement.
#   
#   Some of the config values are directely related to implmentation
#   of feature within a Go website. Please, see the following template
#   for implementation of these feature.
#   https://github.com/kambahr/go-webstandard
#----------------------------------------------------------------------

# This is the hostname that will be accessed from outside
# i.e. mydomain.com.
hostname         localhost

# Ignored if port numbered is passed as a cmdline arg: --portno.
# Note that when run as a service --portno arg is used.
portno           1265
proto            http

# This will s the entire site; used for times that the whole
# site needs to be worked on. Your app will have to response 
# to requests (and display a maint-page) accordingly.
maintenance-window     on
 
# location of certificate and private files;
# both in the PEM format and must be full path.
# The paths can be an local paths; but 
# /appdata/certs/<domain anme>/ is recommended.
TLS 
   #cert /usr/local/mydomain/appdata/tls/certx.pem
   #key /usr/local/mydomain/appdata/tls/keyx.pem

Admin
   # List of ip address that will be allowed to access 
   # the admin module; separated by comma; otherwise, the admin
   # section of the website will only be served to the local machine.
   allowed-ip-addr  <ip add 1>, <ip add 2>

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

# If GET is omitted; the master page will still be written to the
# response with error 405.
HTTP
   allowed-methods     GET, OPTIONS, CONNECT, HEAD
   
`
	cnfTemplateBlockedIP string = `
# ip addresses in this file will be blocked from connecting the website.
# the following is the format:
# <ip address><minimum of one space><description>
# Examples:
# 10.12.3.4 Tried to reach a config path
# 10.12.3.4 Too frequent calls
# 10.12.3.4 Spoof attempt for /admin, /.env, wp- paths (bad intentions)
`
)
