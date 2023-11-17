// Copyright 2023 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientcmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ironcore-dev/ironcore/orictl/decoder"
	"k8s.io/client-go/util/homedir"
)

const (
	RecommendedConfigPathFlag   = "config"
	RecommendedConfigPathEnvVar = "ORCITL_MACHINE_CONFIG"
	RecommendedHomeDir          = ".orictl-machine"
	RecommendedFileName         = "config"
)

var (
	RecommendedConfigDir = filepath.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedHomeFile  = filepath.Join(RecommendedConfigDir, RecommendedFileName)
)

type Column struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

type TableConfig struct {
	PrependMachineColumns []Column `json:"prependMachineColumns,omitempty"`
	AppendMachineColumns  []Column `json:"appendMachineColumns,omitempty"`
}

type Config struct {
	TableConfig *TableConfig `json:"tableConfig,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{}
}

func ReadConfig(data []byte) (*Config, error) {
	cfg := &Config{}
	if err := decoder.Decode(data, cfg); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}
	return cfg, nil
}

func ReadConfigFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return ReadConfig(data)
}

func GetConfig(filename string) (*Config, error) {
	if filename != "" {
		return ReadConfigFile(filename)
	}

	if configPath := os.Getenv(RecommendedConfigPathEnvVar); configPath != "" {
		cfg, err := ReadConfigFile(configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			return cfg, nil
		}
	}

	cfg, err := ReadConfigFile(RecommendedHomeFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		return cfg, nil
	}

	return DefaultConfig(), nil
}
