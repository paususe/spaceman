package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/smallfish/simpleyaml"
)

/*
	configFiles allows to keep track of existing configurations
*/
type configFiles struct {
	global  string
	local   string
	session string
	used    string
}

// Config object constructor
func Config() *configFiles {
	cfg := new(configFiles)
	cfg.global = "/etc/rhn/spaceman.conf"
	cfg.local = cfg.expandPath("~/.config/spaceman/config.conf")
	cfg.session = cfg.expandPath("~/.config/spaceman/session.conf")
	cfg.used = cfg.local

	return cfg
}

// Expands "~" to "$HOME".
func (cfg *configFiles) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

// Return current configuration file
func (cfg *configFiles) getConfigFile(ctx *cli.Context) string {
	custom := ctx.GlobalString("config")
	if custom != "" {
		cfg.used = custom
	}

	return cfg.used
}

func (cfg *configFiles) checkFail(err error, message string) {
	if err != nil {
		log.Fatal(err)
		panic(message)
	}
}

func (cfg *configFiles) getConfig(ctx *cli.Context, sections ...string) *map[string]interface{} {
	filename := cfg.getConfigFile(ctx)
	if filename != "" {
		filename = cfg.expandPath(filename)
		source, err := ioutil.ReadFile(filename)
		cfg.checkFail(err, "Unable to read configuration file")

		data, err := simpleyaml.NewYaml(source)
		cfg.checkFail(err, "Unable to parse YAML data")

		content := make(map[string]interface{})
		globalConfig, err := data.Map()
		cfg.checkFail(err, "Configuration syntax error: structure expected")
		for _, section := range sections {
			sectionConfig, exist := globalConfig[section]
			if exist {
				content[section] = sectionConfig
			} else {
				log.Printf("Section '%s' does not exist", section)
			}
		}
		if len(content) == 0 {
			log.Fatal(fmt.Sprintf("No configuration found for %s sections", strings.Join(sections, ", ")))
		}

		return &content
	}
	panic("Unable to obtain configuration")
}

var configuration configFiles

func init() {
	configuration = *Config()
}
