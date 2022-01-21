package webconfig

import (
	"fmt"
	"os"
)

// NewPage initalizes the NewWebConfig. It creates the
// default directories and starts the internal daemon.
func NewWebConfig(webRootPath string) *Config {
	var c Config
	c.WebRootPath = webRootPath

	c.refreshRate = 15

	// Create the appdata if it does not exist
	c.AppDataPath = fmt.Sprintf("%s/appdata", c.WebRootPath)
	if !fileOrDirExists(c.AppDataPath) {
		os.Mkdir(c.AppDataPath, os.ModePerm)
	}
	c.ConfigFilePath = fmt.Sprintf("%s/.cfg/.all", c.AppDataPath)

	cfgDir := fmt.Sprintf("%s/.cfg", c.AppDataPath)
	if !fileOrDirExists(cfgDir) {
		os.Mkdir(cfgDir, os.ModePerm)
	}

	certDir := fmt.Sprintf("%s/appdata/certs", c.WebRootPath)
	if !fileOrDirExists(certDir) {
		os.Mkdir(certDir, os.ModePerm)
	}
	selfcertDir := fmt.Sprintf("%s/appdata/certs/self", c.WebRootPath)
	if !fileOrDirExists(selfcertDir) {
		os.Mkdir(selfcertDir, os.ModePerm)
	}

	if !fileOrDirExists(c.ConfigFilePath) {
		c.writeDefaultConfig()
	}

	c.GetConfig()

	go c.refreshConfig()

	// Need to do this on start.
	if c.MessageBanner.On && c.MessageBanner.SecondsToDisplay > 0 {
		c.MessageBanner.TickCount = c.MessageBanner.SecondsToDisplay
		go c.setTimeoutResetMsgBanner()
	}

	return &c
}
