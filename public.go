package webconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kambahr/go-mathsets"
)

// Refresh is the same as GetConfig. It reads the config from disk
// and fills-in the build-in fields accordingly.
func (c *Config) Refresh() {
	c.GetConfig()
}

// GetJSON returns json of the Config struct.
func (c *Config) GetJSON() string {
	b, err := json.Marshal(&c)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	s := fmt.Sprintf("%s", string(b))

	s = strings.ReplaceAll(s, "\\u003e", ">")
	s = strings.ReplaceAll(s, "\\u003c", "<")

	return s
}

// hasSection check for the line(s) above to see if there
// a section.
func (c *Config) hasSection(lines []string, pos int) bool {

	x := pos - 1
	for {
		if x < 1 {
			break
		}
		l := strings.Trim(lines[x], " ")
		if c.skipLine(l) {
			x--
			continue
		}
		v := strings.Split(l, " ")
		if len(v) < 2 {
			return true
		}
		x--
	}

	return false
}

// getData fills-in the Data section of the config.
func (c *Config) getData(line []string) {

	inx := -1
	for i := 0; i < len(line); i++ {
		l := strings.Trim(line[i], " ")
		if c.skipLine(l) || l == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(l), "data") {
			inx = i
			break
		}
	}

	if inx < 0 {
		// Data section was not found
		return
	}

	// Go forward one
	inx++

	// Get the count of the map
	dataCount := 0
	for i := inx; i < len(line); i++ {
		l := strings.Trim(line[i], " ")
		if c.skipLine(l) || l == "" {
			continue
		}
		dataCount++
	}

	// Split each line: as left/right (key/vlaue) and append to the Data map
	c.Data = make(map[string]string, dataCount)
	for i := inx; i < len(line); i++ {
		l := strings.Trim(line[i], " ")
		if c.skipLine(l) || l == "" {
			continue
		}
		v := strings.Split(line[i], " ")
		key := v[0]
		val := ""
		// account for multi-spaces between key and value.
		for j := 1; j < len(v); j++ {
			if v[j] != "" {
				val = fmt.Sprintf("%s%s", val, v[j])
				break
			}
		}
		c.Data[key] = val
	}
}

// GetConfig reads config values from file /appdata/.cfg.
// All values are part of a struct so lingering text in the config
// file will not be processed.
func (c *Config) GetConfig() {

	f, err := ReadFile(c.ConfigFilePath)
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
		if c.skipLine(l) || l == "" {
			continue
		}

		lLower := strings.ToLower(l)

		if strings.HasPrefix(lLower, "maintenance-window") {
			s := strings.ToLower(c.parseCofigLine(l, "maintenance-window"))

			if s == "on" {
				c.MaintenanceWindowOn = true
			} else {
				c.MaintenanceWindowOn = false
			}
		} else if strings.HasPrefix(lLower, "site") {
			keys := []string{"hostname", "portno", "proto"}
			i++
			i = c.getConfigLeaves(line, i, "site", keys)

		} else if strings.HasPrefix(lLower, "tls") {
			keys := []string{"cert", "key"}
			i++
			i = c.getConfigLeaves(line, i, "tls", keys)

		} else if strings.HasPrefix(lLower, "admin") {
			keys := []string{"allowed-ip-addr", "run-on-startup", "portno"}
			i++
			i = c.getConfigLeaves(line, i, "admin", keys)

		} else if strings.HasPrefix(lLower, "redirect-http-to-https") {

			if strings.ToLower(c.parseCofigLine(l, "redirect-http-to-https")) == "yes" {
				c.RedirectHTTPtoHTTPS = true
			} else {
				c.RedirectHTTPtoHTTPS = false
			}

		} else if strings.HasPrefix(lLower, "messagebanner") {

			keys := []string{"display-mode", "seconds-to-display"}
			i++
			i = c.getConfigLeaves(line, i, "MessageBanner", keys)

		} else if strings.HasPrefix(lLower, "http") {

			keys := []string{"allowed-methods"}
			i++
			i = c.getConfigLeaves(line, i, "HTTP", keys)

		} else if strings.HasPrefix(lLower, "urlpaths") {

			keys := []string{"restrict-paths", "exclude-paths", "forward-paths"}
			i++
			i = c.getConfigLeaves(line, i, "URLPaths", keys)
		}
	}

	if c.MessageBanner.On && c.MessageBanner.SecondsToDisplay > 0 {
		c.MessageBanner.TickCount = c.MessageBanner.SecondsToDisplay
		go c.setTimeoutResetMsgBanner()
	}

	c.getData(line)

	// Get the offenders
	blockedIPPath := fmt.Sprintf("%s/.cfg/blocked-ip", c.AppDataPath)
	if c.fileOrDirExists(blockedIPPath) {
		f, err := ReadFile(blockedIPPath)
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

	f, err := ReadFile(c.ConfigFilePath)
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

	// Refresh
	c.GetConfig()
}
