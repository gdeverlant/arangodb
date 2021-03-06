//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package service

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sync"
)

const (
	// SetupConfigVersion is the semantic version of the process that created this.
	// If the structure of SetupConfigFile (or any underlying fields) or its semantics change, you must increase this version.
	SetupConfigVersion = "0.2.1"
	setupFileName      = "setup.json"
)

// SetupConfigFile is the JSON structure stored in the setup file of this process.
type SetupConfigFile struct {
	Version          string `json:"version"` // Version of the process that created this. If the structure or semantics changed, you must increase this version.
	ID               string `json:"id"`      // My unique peer ID
	Peers            peers  `json:"peers"`
	StartLocalSlaves bool   `json:"start-local-slaves,omitempty"`
}

// saveSetup saves the current peer configuration to disk.
func (s *Service) saveSetup() error {
	cfg := SetupConfigFile{
		Version:          SetupConfigVersion,
		ID:               s.ID,
		Peers:            s.myPeers,
		StartLocalSlaves: s.StartLocalSlaves,
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		s.log.Errorf("Cannot serialize config: %#v", err)
		return maskAny(err)
	}
	if err := ioutil.WriteFile(filepath.Join(s.DataDir, setupFileName), b, 0644); err != nil {
		s.log.Errorf("Error writing setup: %#v", err)
		return maskAny(err)
	}
	return nil
}

// relaunch tries to read a setup.json config file and relaunch when that file exists and is valid.
// Returns true on relaunch or false to continue with a fresh start.
func (s *Service) relaunch(runner Runner) bool {
	// Is this a new start or a restart?
	if setupContent, err := ioutil.ReadFile(filepath.Join(s.DataDir, setupFileName)); err == nil {
		// Could read file
		var cfg SetupConfigFile
		if err := json.Unmarshal(setupContent, &cfg); err == nil {
			if cfg.Version == SetupConfigVersion {
				s.myPeers = cfg.Peers
				s.ID = cfg.ID
				s.AgencySize = s.myPeers.AgencySize
				s.log.Infof("Relaunching service with id '%s' on %s:%d...", s.ID, s.OwnAddress, s.announcePort)
				s.startHTTPServer()
				wg := &sync.WaitGroup{}
				if cfg.StartLocalSlaves {
					s.startLocalSlaves(wg, cfg.Peers.Peers)
				}
				s.startRunning(runner)
				wg.Wait()
				return true
			}
			s.log.Warningf("%s is outdated. Starting fresh...", setupFileName)
		} else {
			s.log.Warningf("Failed to unmarshal existing %s: %#v", setupFileName, err)
		}
	}
	return false
}
