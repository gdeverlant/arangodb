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

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arangodb-helper/arangodb/client"
	shell "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/shavac/gexpect"
)

const (
	ctrlC           = "\u0003"
	whatCluster     = "cluster"
	whatSingle      = "single server"
	testModeProcess = "localprocess"
	testModeDocker  = "docker"
)

var (
	isVerbose bool
	testModes []string
)

func init() {
	isVerbose = os.Getenv("VERBOSE") != ""
	testModes = strings.Split(os.Getenv("TEST_MODES"), " ")
	if len(testModes) == 1 && testModes[0] == "" {
		testModes = nil
	}
}

func needTestMode(t *testing.T, testMode string) {
	for _, x := range testModes {
		if x == testMode {
			return
		}
	}
	if len(testModes) == 0 {
		return
	}
	t.Skipf("Test mode '%s' not set", testMode)
}

// Spawn a command an return its process.
func Spawn(t *testing.T, command string) *gexpect.SubProcess {
	args, err := shell.Split(os.ExpandEnv(command))
	if err != nil {
		t.Fatal(describe(err))
	}
	if isVerbose {
		t.Log(args, len(args))
	}
	p, err := gexpect.NewSubProcess(args[0], args[1:]...)
	if err != nil {
		t.Fatal(describe(err))
	}
	if err := p.Start(); err != nil {
		p.Close()
		t.Fatal(describe(err))
	}
	return p
}

// SetUniqueDataDir creates a temp dir and sets the DATA_DIR environment variable to it.
func SetUniqueDataDir(t *testing.T) string {
	dataDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(describe(err))
	}
	os.Setenv("DATA_DIR", dataDir)
	return dataDir
}

// WaitUntilStarterReady waits until all given starter processes have reached the "Your cluster is ready state"
func WaitUntilStarterReady(t *testing.T, what string, starters ...*gexpect.SubProcess) bool {
	g := sync.WaitGroup{}
	result := true
	for _, starter := range starters {
		starter := starter // Used in nested function
		g.Add(1)
		go func() {
			defer g.Done()
			if _, err := starter.ExpectTimeout(time.Minute, regexp.MustCompile(fmt.Sprintf("Your %s can now be accessed with a browser at", what))); err != nil {
				result = false
				t.Errorf("Starter is not ready in time: %s", describe(err))
			}
		}()
	}
	g.Wait()
	return result
}

// SendIntrAndWait stops all all given starter processes by sending a Ctrl-C into it.
// It then waits until the process has terminated.
func SendIntrAndWait(t *testing.T, starters ...*gexpect.SubProcess) bool {
	g := sync.WaitGroup{}
	result := true
	for _, starter := range starters {
		starter := starter // Used in nested function
		g.Add(1)
		go func() {
			defer g.Done()
			if err := starter.WaitTimeout(time.Second * 30); err != nil {
				result = false
				t.Errorf("Starter is not stopped in time: %s", describe(err))
			}
		}()
	}
	time.Sleep(time.Second)
	for _, starter := range starters {
		starter.Term.SendIntr()
		//starter.Send(ctrlC)
	}
	g.Wait()
	return result
}

// describe returns a string description of the given error.
func describe(err error) string {
	if err == nil {
		return "nil"
	}
	cause := errors.Cause(err)
	c, _ := json.Marshal(cause)
	cStr := fmt.Sprintf("%#v (%s)", cause, string(c))
	if cause.Error() != err.Error() {
		return fmt.Sprintf("%v caused by %v", err, cStr)
	} else {
		return cStr
	}
}

// NewStarterClient creates a new starter API instance for the given endpoint, failing the test on errors.
func NewStarterClient(t *testing.T, endpoint string) client.API {
	ep, err := url.Parse(endpoint)
	if err != nil {
		t.Fatalf("Failed to parse starter endpoint: %s", describe(err))
	}
	c, err := client.NewArangoStarterClient(*ep)
	if err != nil {
		t.Fatalf("Failed to create starter client: %s", describe(err))
	}
	return c
}

// ShutdownStarter calls the starter the shutdown via the HTTP API.
func ShutdownStarter(t *testing.T, endpoint string) {
	c := NewStarterClient(t, endpoint)
	if err := c.Shutdown(context.Background(), false); err != nil {
		t.Errorf("Shutdown failed: %s", describe(err))
	}
	WaitUntilStarterGone(t, endpoint)
}

// WaitUntilStarterGone waits until the starter at given endpoint no longer responds to queries.
func WaitUntilStarterGone(t *testing.T, endpoint string) {
	c := NewStarterClient(t, endpoint)
	failures := 0
	for {
		if _, err := c.Version(context.Background()); err != nil {
			// Version request failed
			failures++
		} else {
			failures = 0
		}
		if failures > 2 {
			// Several failures, we assume the starter is really gone now
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
}
