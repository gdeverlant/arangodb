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
	"os"
	"testing"
	"time"
)

// TestProcessSingle runs `arangodb --starter.mode=single`
func TestProcessSingle(t *testing.T) {
	needTestMode(t, testModeProcess)
	dataDir := SetUniqueDataDir(t)
	defer os.RemoveAll(dataDir)

	start := time.Now()

	child := Spawn(t, "${STARTER} --starter.mode=single")
	defer child.Close()

	if ok := WaitUntilStarterReady(t, whatSingle, child); ok {
		t.Logf("Single server start took %s", time.Since(start))
		testSingle(t, insecureStarterEndpoint(0), false)
	}

	if isVerbose {
		t.Log("Waiting for termination")
	}
	SendIntrAndWait(t, child)
}

// TestProcessSingleShutdownViaAPI runs `arangodb --starter.mode=single`, stopping it through the `/shutdown` API.
func TestProcessSingleShutdownViaAPI(t *testing.T) {
	needTestMode(t, testModeProcess)
	dataDir := SetUniqueDataDir(t)
	defer os.RemoveAll(dataDir)

	start := time.Now()

	child := Spawn(t, "${STARTER} --starter.mode=single")
	defer child.Close()

	if ok := WaitUntilStarterReady(t, whatSingle, child); ok {
		t.Logf("Single server start took %s", time.Since(start))
		testSingle(t, insecureStarterEndpoint(0), false)
	}

	if isVerbose {
		t.Log("Waiting for termination")
	}
	ShutdownStarter(t, insecureStarterEndpoint(0))
}

// TestProcessSingleAutoKeyFile runs `arangodb --starter.mode=single --ssl.auto-key`
func TestProcessSingleAutoKeyFile(t *testing.T) {
	needTestMode(t, testModeProcess)
	dataDir := SetUniqueDataDir(t)
	defer os.RemoveAll(dataDir)

	start := time.Now()

	child := Spawn(t, "${STARTER} --starter.mode=single --ssl.auto-key")
	defer child.Close()

	if ok := WaitUntilStarterReady(t, whatSingle, child); ok {
		t.Logf("Single server start took %s", time.Since(start))
		testSingle(t, secureStarterEndpoint(0), true)
	}

	if isVerbose {
		t.Log("Waiting for termination")
	}
	SendIntrAndWait(t, child)
}

// TestOldProcessSingle runs `arangodb --mode=single`
func TestOldProcessSingle(t *testing.T) {
	needTestMode(t, testModeProcess)
	dataDir := SetUniqueDataDir(t)
	defer os.RemoveAll(dataDir)

	start := time.Now()

	child := Spawn(t, "${STARTER} --mode=single")
	defer child.Close()

	if ok := WaitUntilStarterReady(t, whatSingle, child); ok {
		t.Logf("Single server start took %s", time.Since(start))
		testSingle(t, insecureStarterEndpoint(0), false)
	}

	if isVerbose {
		t.Log("Waiting for termination")
	}
	SendIntrAndWait(t, child)
}

// TestOldProcessSingleAutoKeyFile runs `arangodb --mode=single --sslAutoKeyFile`
func TestOldProcessSingleAutoKeyFile(t *testing.T) {
	needTestMode(t, testModeProcess)
	dataDir := SetUniqueDataDir(t)
	defer os.RemoveAll(dataDir)

	start := time.Now()

	child := Spawn(t, "${STARTER} --mode=single --sslAutoKeyFile")
	defer child.Close()

	if ok := WaitUntilStarterReady(t, whatSingle, child); ok {
		t.Logf("Single server start took %s", time.Since(start))
		testSingle(t, secureStarterEndpoint(0), true)
	}

	if isVerbose {
		t.Log("Waiting for termination")
	}
	SendIntrAndWait(t, child)
}
