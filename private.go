package webconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// setTimeoutResetMsgBanner starts a count-down to reset the
// value of display-mode back to off.
func (c *Config) setTimeoutResetMsgBanner() {
lblAgain:
	if !c.MessageBanner.On || c.MessageBanner.TickCount < 1 ||
		c.MessageBanner.SecondsToDisplay < 1 /* means the webserver will do this */ {
		return
	}

	c.MessageBanner.TickCount--

	if c.MessageBanner.TickCount < 1 {
		c.MessageBanner.On = false

		c.UpdateConfigValue("MessageBanner", "display-mode", "off")
	}

	time.Sleep(time.Second)
	goto lblAgain
}

// refreshConfig reads the config values from the appdata/.cfg file
// so that the website [service] does not have to be restarted if
// a value changes.
func (c *Config) refreshConfig() {
lblAgain:

	time.Sleep(time.Duration(c.refreshRate) * time.Second)

	c.GetConfig()

	goto lblAgain // avoid recursion
}

// parseCofigLine extracts the value from a line of the config data.
func (c *Config) parseCofigLine(line string, key string) string {

	line = c.trimLine(line)

	i := strings.Index(line, " ")

	if i < 0 {
		// not found
		return line
	}

	return c.trimLine(line[len(line[:i]):])
}

// skipLine tells if the line is not related or a comment.
func (c *Config) skipLine(l string) bool {

	l = strings.Trim(l, " ")

	if strings.HasPrefix(l, "#") || l == "" {
		return true
	}

	return false
}

// writeDefaultConfig creates a default config.
// The template is in defs.go (not on disk).
func (c *Config) writeDefaultConfig() {
	cdir := fmt.Sprintf("%s/.cfg", c.AppDataPath)
	os.Mkdir(cdir, os.ModePerm)
	f, err := os.Create(c.ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(cfgTemplateAll)
	f.Close()

	// Also create the blocked-ip file
	blockedIPPath := fmt.Sprintf("%s/.cfg/blocked-ip", c.AppDataPath)
	f, err = os.Create(blockedIPPath)
	if err != nil {
		log.Fatal(err)
	}
	f.WriteString(cnfTemplateBlockedIP)
	f.Close()
}

// getConfigLeaves get the config values under a section;
// example:
//   TLS
//      cert /usr/local/mydomain/appdata/tls/certx.pem
//      key  /usr/local/mydomain/appdata/tls/keyx.pem
func (c *Config) getConfigLeaves(lines []string, i int, section string, keys []string) int {

	section = strings.ToLower(section)

	// Make the keys lower
	for j := 0; j < len(keys); j++ {
		keys[j] = strings.ToLower(keys[j])
	}

	for {
		if i >= len(lines) {
			return i
		}
		for {
			if i >= len(lines) {
				return i
			}
			if c.skipLine(lines[i]) {
				i++
			} else {
				break
			}
		}
		if i >= len(lines) {
			return i
		}
		l := c.trimLine(lines[i])
		hitValue := false
		for j := 0; j < len(keys); j++ {
			if strings.HasPrefix(l, keys[j]) {
				hitValue = true
				break
			}
		}
		if !hitValue {
			break
		}

		// See if the section matches; as there could be same values but
		// different section; have keep going up as here could be spaces upwards.
		hitSection := false
		x := i - 1
		for {
			if x < 1 {
				break
			}
			if c.skipLine(lines[x]) {
				x--
				continue
			}
			sx := strings.ToLower(c.trimLine(lines[x]))
			if sx == section {
				hitSection = true
				break
			}
			x--
		}
		if !hitSection {
			i++
			if i >= len(lines) {
				return i
			}
			continue
		}

		if section == "site" {
			if strings.HasPrefix(l, "hostname") {
				c.Site.HostName = c.parseCofigLine(l, "hostname")

			} else if strings.HasPrefix(l, "alternate-hostnames") {
				// empty the array first
				c.Site.AlternateHostNames = make([]string, 0)
				s := c.parseCofigLine(l, "alternate-hostnames")
				s = strings.Trim(s, " ")
				if s != "" {
					v := strings.Split(s, ",")
					for i := 0; i < len(v); i++ {
						if v[i] == "" {
							continue
						}
						v[i] = strings.ToLower(strings.Trim(v[i], " "))
						c.Site.AlternateHostNames = append(c.Site.AlternateHostNames, v[i])
					}
				}
			} else if strings.HasPrefix(l, "portno") {
				s := c.parseCofigLine(l, "portno")
				c.Site.PortNo, _ = strconv.Atoi(s)

			} else if strings.HasPrefix(l, "proto") {
				c.Site.Proto = c.parseCofigLine(l, "proto")
			}

		} else if section == "tls" {
			if strings.HasPrefix(l, "cert") {
				// cert PEM file
				c.TLS.CertFilePath = c.parseCofigLine(l, "cert")

			} else if strings.HasPrefix(l, "key") {
				// private key PEM file
				c.TLS.KeyFilePath = c.parseCofigLine(l, "key")

			}

		} else if section == "admin" {
			if strings.HasPrefix(l, "run-on-startup") {
				s := c.parseCofigLine(l, "run-on-startup")

				if s == "yes" {
					c.Admin.RunOnStartup = true
				} else {
					c.Admin.RunOnStartup = false
				}
			} else if strings.HasPrefix(l, "portno") {
				s := c.parseCofigLine(l, "portno")
				ix, _ := strconv.Atoi(s)
				c.Admin.PortNo = uint(ix)

			} else if strings.HasPrefix(l, "allowed-ip-addr") {
				s := c.parseCofigLine(l, "allowed-ip-addr")
				c.Admin.AllowedIP = strings.Split(s, ",")
				for j := 0; j < len(c.Admin.AllowedIP); j++ {
					c.Admin.AllowedIP[j] = strings.Trim(c.Admin.AllowedIP[j], " ")
				}
			}
		} else if section == "messagebanner" {
			if strings.HasPrefix(l, "display-mode") {
				s := c.parseCofigLine(l, "display-mode")

				if s == "on" {
					c.MessageBanner.On = true
				} else {
					c.MessageBanner.On = false
				}
			}
			if strings.HasPrefix(l, "seconds-to-display") {
				s := c.parseCofigLine(l, "seconds-to-display")
				c.MessageBanner.SecondsToDisplay, _ = strconv.Atoi(s)
			}

		} else if section == "http" {
			if strings.HasPrefix(l, "allowed-methods") {
				s := c.parseCofigLine(l, "allowed-methods")
				c.HTTP.AllowedMethods = strings.Split(s, ",")
				for j := 0; j < len(c.HTTP.AllowedMethods); j++ {
					c.HTTP.AllowedMethods[j] = strings.Trim(c.HTTP.AllowedMethods[j], " ")
				}
			}
		} else if section == "urlpaths" {
			if strings.HasPrefix(l, "forward-paths") {
				s := c.parseCofigLine(l, "forward-paths")
				c.URLPaths.Forward = strings.Split(s, ",")
				for j := 0; j < len(c.URLPaths.Forward); j++ {
					c.URLPaths.Forward[j] = strings.TrimLeft(c.URLPaths.Forward[j], " ")
					c.URLPaths.Forward[j] = strings.TrimRight(c.URLPaths.Forward[j], " ")

					// Must begin with /
					v := strings.Split(c.URLPaths.Forward[j], "|")
					left := ""
					right := ""
					if len(v) > 1 {
						left = v[0]
						right = v[1]
					}
					if !strings.HasPrefix(right, "/") {
						// Replace the right-side with an error so that it will be processed.
						// The error is only visible internally during debugging.
						c.URLPaths.Forward[j] = fmt.Sprintf("%s|~@error: fully qualified url-forwarding not allowed", left)
					}

				}
			} else if strings.HasPrefix(l, "restrict-paths") {
				s := c.parseCofigLine(l, "restrict-paths")
				c.URLPaths.Restrict = strings.Split(s, ",")
				for j := 0; j < len(c.URLPaths.Restrict); j++ {
					c.URLPaths.Restrict[j] = strings.TrimLeft(c.URLPaths.Restrict[j], " ")
					c.URLPaths.Restrict[j] = strings.TrimRight(c.URLPaths.Restrict[j], " ")
				}
			} else if strings.HasPrefix(l, "exclude-paths") {
				s := c.parseCofigLine(l, "exclude-paths")
				c.URLPaths.Exclude = strings.Split(s, ",")
				for j := 0; j < len(c.URLPaths.Exclude); j++ {
					c.URLPaths.Exclude[j] = strings.TrimLeft(c.URLPaths.Exclude[j], " ")
					c.URLPaths.Exclude[j] = strings.TrimRight(c.URLPaths.Exclude[j], " ")
				}
			} else if strings.HasPrefix(l, "conditional-http-service") {
				s := c.parseCofigLine(l, "conditional-http-service")
				json.Unmarshal([]byte(s), &c.URLPaths.ServeOnlyTo)
			}
		}

		// go to next line
		i++
	}

	// In case a parent was on the last line
	i--

	return i
}
