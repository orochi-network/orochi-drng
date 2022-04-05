// Copyright 2019 P2Sub Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/orochi-network/orochimaru/config"
	"github.com/orochi-network/orochimaru/logger"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	log = logger.GetSugarLogger()
}

// OrochiAppConfig conf wrapper for P2Sub
type OrochiAppConfig struct {
	cfg *config.Config
}

// FlagConfig flags configuration
type FlagConfig struct {
	name        string
	dataType    string
	value       interface{}
	required    bool
	description string
}

var AppConfig *OrochiAppConfig
var confOnce sync.Once

// GetOrochiAppConfig get singleton instance of Config
func GetOrochiAppConfig() *OrochiAppConfig {
	confOnce.Do(func() {
		AppConfig = &OrochiAppConfig{cfg: config.New()}
	})
	return AppConfig
}

// GetKeyFile get key file
func (p *OrochiAppConfig) GetKeyFile() string {
	return p.cfg.GetString("node::key_file")
}

// SetKeyFile set key file
func (p *OrochiAppConfig) SetKeyFile(keyFile string) bool {
	return p.cfg.Set("node::key_file", keyFile)
}

// GetBindPort get bind port of current node
func (p *OrochiAppConfig) GetBindPort() uint {
	return p.cfg.GetUint("node::bind_port")
}

// SetBindPort set bind port of current node
func (p *OrochiAppConfig) SetBindPort(bindPort uint) bool {
	return p.cfg.Set("node::bind_port", bindPort)
}

// GetBindHost get bind host
func (p *OrochiAppConfig) GetBindHost() string {
	return p.cfg.GetString("node::bind_host")
}

// SetBindHost set bind host
func (p *OrochiAppConfig) SetBindHost(bindHost string) bool {
	return p.cfg.Set("node::bind_host", bindHost)
}

// GetDirectConnect get direct connect node's identity
func (p *OrochiAppConfig) GetDirectConnect() string {
	return p.cfg.GetString("node::direct_connect")
}

// SetDirectConnect set direct connect node's identity
func (p *OrochiAppConfig) SetDirectConnect(nodeAddress string) bool {
	return p.cfg.Set("node::direct_connect", nodeAddress)
}

// GetDomain get domain of node discovery
func (p *OrochiAppConfig) GetDomain() string {
	return p.cfg.GetString("node::domain")
}

// SetDomain set domain of node discovery
func (p *OrochiAppConfig) SetDomain(domain string) bool {
	return p.cfg.Set("node::domain", domain)
}

func (f FlagConfig) valToBool() bool {
	if v, ok := f.value.(bool); ok {
		return v
	}
	return false
}

func (f FlagConfig) valToString() string {
	if v, ok := f.value.(string); ok {
		return v
	}
	return ""
}

func (f FlagConfig) valToInt() int {
	if v, ok := f.value.(int); ok {
		return v
	}
	return 0
}

func (f FlagConfig) valToUint() uint {
	if v, ok := f.value.(uint); ok {
		return v
	}
	return 0
}

func nameToFlag(name string) string {
	parts := strings.Split(name, "::")
	if len(parts) == 2 {
		// node::key-file
		return strings.ReplaceAll(parts[1], "_", "-")
	}
	log.Panic(errors.New("wrong format of flag name"))
	return ""
}

// Init common components
func init() {
	AppConfig = GetOrochiAppConfig()

	// All flags configuration
	flagConfigs := []FlagConfig{
		{
			name:        "node::key_file",
			dataType:    "string",
			value:       "",
			description: "File name to save/load key configuration",
			required:    true,
		},
		{
			name:        "node::direct_connect",
			dataType:    "string",
			value:       "",
			description: "Direct connect to a given node",
		},
		{
			name:        "node::domain",
			dataType:    "string",
			value:       "P2Sub::alpha::0.0.1",
			description: "Rendezvous string used to discover same node",
		},
		{
			name:        "node::bind_port",
			dataType:    "uint",
			value:       0,
			description: "Bind port of current node",
			required:    true,
		},
		{
			name:        "node::bind_host",
			dataType:    "string",
			value:       "0.0.0.0",
			description: "Bind host of current node",
			required:    true,
		},
	}

	// Transform flag config to arguments
	for _, flagConf := range flagConfigs {
		if flagConf.dataType == "string" {
			flag.String(nameToFlag(flagConf.name), flagConf.valToString(), flagConf.description)
		} else if flagConf.dataType == "bool" {
			flag.Bool(nameToFlag(flagConf.name), flagConf.valToBool(), flagConf.description)
		} else if flagConf.dataType == "uint" {
			flag.Uint(nameToFlag(flagConf.name), flagConf.valToUint(), flagConf.description)
		} else {
			flag.Int(nameToFlag(flagConf.name), flagConf.valToInt(), flagConf.description)
		}
	}

	// Parse flags
	flag.Parse()

	isFlagOn := make(map[string]bool)

	flag.Visit(func(f *flag.Flag) {
		isFlagOn[f.Name] = true
	})

	//Save configuration
	for _, flagConf := range flagConfigs {

		if flagConf.required && !isFlagOn[nameToFlag(flagConf.name)] {
			flag.Usage()
			os.Exit(1)
		}
		rawValue := flag.Lookup(nameToFlag(flagConf.name)).Value.(flag.Getter).Get()
		if isFlagOn[nameToFlag(flagConf.name)] {
			log.Infof("Flag config: %s value: %v", flagConf.name, rawValue)
		}

		AppConfig.cfg.Set(flagConf.name, rawValue)
	}
}
