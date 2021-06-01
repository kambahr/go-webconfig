package webconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kambahr/go-mathsets"
)

// Refresh is the same as GetConfig. It reads the config from disk
// and fills-in the build-in fields accordingly.
func (c *Config) Refresh() {
	c.GetConfig()
}

// GetConfig reads config values from file /appdata/.cfg.
// All values are part of a struct so lingering text in the config
// file will not be processed.
func (c *Config) GetConfig() {

	f, err := ioutil.ReadFile(c.ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// do not process, if the file has not changed.
	hs := fmt.Sprintf("%x", mathsets.Hash256Twice(f))
	if hs == c.ConfigFileLastHash {
		return
	}
	c.ConfigFileLastHash = hs

	linex := strings.Split(string(f), "\n")
	var line []string

	// Put the continuation of lines together.
	// Also remove tabs and trime lines.
	for i := 0; i < len(linex); i++ {

		linex[i] = c.trimLine(linex[i])

		if strings.HasSuffix(linex[i], "\\") {
			// this and the next line
			if (i + 1) >= len(linex) {
				break
			}

			// Take out the \ at the end
			linex[i] = linex[i][0 : len(linex[i])-1]

			linex[i+1] = c.trimLine(linex[i+1])

			s := fmt.Sprintf("%s%s", linex[i], linex[i+1])
			line = append(line, s)
			i++
			continue
		}

		line = append(line, linex[i])
	}
	for i := 0; i < len(line); i++ {

		l := strings.Trim(line[i], " ")
		if c.skipLine(l) {
			continue
		}
		// Values could be any where in the file.
		if strings.HasPrefix(l, "hostname") {
			c.HostName = c.parseCofigLine(l, "hostname")

		} else if strings.HasPrefix(l, "portno") {
			s := c.parseCofigLine(l, "portno")
			c.PortNo, _ = strconv.Atoi(s)

		} else if strings.HasPrefix(l, "proto") {
			c.Proto = strings.ToUpper(c.parseCofigLine(l, "proto"))

		} else if strings.HasPrefix(l, "maintenance-window") {
			s := strings.ToLower(c.parseCofigLine(l, "maintenance-window"))

			if s == "on" {
				c.MaintenanceWindowOn = true
			} else {
				c.MaintenanceWindowOn = false
			}

		} else if strings.ToLower(l) == "tls" {
			keys := []string{"cert", "key"}
			i++
			i = c.getConfigLeaves(line, i, keys)

		} else if strings.HasPrefix(l, "redirect-http-to-https") {

			if strings.ToLower(c.parseCofigLine(l, "redirect-http-to-https")) == "yes" {
				c.RedirectHTTPtoHTTPS = true
			} else {
				c.RedirectHTTPtoHTTPS = false
			}

		} else if strings.ToLower(l) == strings.ToLower("MessageBanner") {

			keys := []string{"display-mode", "seconds-to-display"}
			i++
			i = c.getConfigLeaves(line, i, keys)

		} else if strings.ToLower(l) == strings.ToLower("HTTP") {

			keys := []string{"allowed-methods"}
			i++
			i = c.getConfigLeaves(line, i, keys)

		} else if strings.ToLower(l) == strings.ToLower("admin") {

			keys := []string{"allowed-ip-addr"}
			i++
			i = c.getConfigLeaves(line, i, keys)

		} else if strings.ToLower(l) == strings.ToLower("URLPaths") {

			keys := []string{"restrict-paths", "exclude-paths", "forward-paths"}
			i++
			i = c.getConfigLeaves(line, i, keys)
		}
	}

	if c.MessageBanner.On && c.MessageBanner.SecondsToDisplay > 0 {
		c.MessageBanner.TickCount = c.MessageBanner.SecondsToDisplay
		go c.setTimeoutResetMsgBanner()
	}

	// Get the offenders
	blockedIPPath := fmt.Sprintf("%s/.cfg/blocked-ip", c.AppDataPath)
	if c.fileOrDirExists(blockedIPPath) {
		f, err := ioutil.ReadFile(blockedIPPath)
		if err != nil {
			log.Fatal(err)
		}

		line = strings.Split(string(f), "\n")
		c.BlockedIP = make([]string, 0)
		for i := 0; i < len(line); i++ {
			l := strings.Trim(line[i], " ")
			if strings.HasPrefix(l, "#") || l == "" {
				continue
			}
			v := strings.Split(l, " ")
			ip := v[0]
			c.BlockedIP = append(c.BlockedIP, ip)
		}
	}
}

// UpdateConfigValue updates a value in the /.cfg/.all config file.
// parent is the name of the section (header). it should be blank, if
// if there is not section name.
// e.g.
//   The following has no parent.
// 		hostname         localhost
//
//   and this one has a parent name and key/value
//       TLS
//          cert /usr/local/mydomain/appdata/tls/certx.pem
//          key /usr/local/mydomain/appdata/tls/keyx.pem
func (c *Config) UpdateConfigValue(parent string, key string, newValue string) {

	f, err := ioutil.ReadFile(c.ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	key = strings.ToLower(key)
	line := strings.Split(string(f), "\n")

	for i := 0; i < len(line); i++ {
		line[i] = strings.Replace(line[i], "\t", " ", -1)
		line[i] = strings.TrimLeft(line[i], " ")
		line[i] = strings.TrimRight(line[i], " ")

		if c.skipLine(line[i]) {
			continue
		}
		l := strings.ToLower(line[i])
		if strings.HasPrefix(l, strings.ToLower(parent)) {
			for {
				i++
				line[i] = strings.Replace(line[i], "\t", " ", -1)
				line[i] = strings.TrimLeft(line[i], " ")
				line[i] = strings.TrimRight(line[i], " ")
				if i >= len(line) {
					break
				}
				if c.skipLine(line[i]) {
					continue
				}
				l = strings.ToLower(line[i])
				if strings.HasPrefix(l, key) {
					line[i] = fmt.Sprintf("   %s      %s", key, newValue)
					goto lblDone
				}
			}
		}
	}
lblDone:
	// Write the lines to disk
	fPath := fmt.Sprintf("%s/.cfg/.all.swap", c.AppDataPath)

	fx, err := os.Create(fPath)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(line); i++ {
		s := fmt.Sprintf("%s\n", line[i])
		fx.WriteString(s)
	}
	err = fx.Close()
	if err == nil {
		// replace the file
		err = os.Rename(fPath, c.ConfigFilePath)
		if err != nil {
			fmt.Println(err)
		}
	}
}
