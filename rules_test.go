//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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

package upgraderules

import (
	"testing"

	driver "github.com/arangodb/go-driver"
)

func TestCheckUpgradeRules(t *testing.T) {
	tests := []struct {
		From    driver.Version
		To      driver.Version
		Allowed bool
	}{
		// Same version
		{"1.2.3", "1.2.3", true},
		// Different major
		{"2.2.3", "1.2.3", false},
		{"1.2.3", "2.2.3", false},
		// Same major, different minor
		{"3.2.2", "3.3.0", true},
		{"3.2.2", "3.3.10", true},
		{"3.3.2", "3.2.10", false},
		{"3.3.2", "3.5.10", false},
		{"3.2.2", "3.3.11111", true},
		// Same major & minor, different patch
		{"3.2.2", "3.2.88", true},
		{"3.2.88", "3.2.8", true},
		{"3.2.88", "3.2.rc7", true},
	}
	for _, test := range tests {
		err := CheckUpgradeRules(test.From, test.To)
		if test.Allowed {
			if err != nil {
				t.Errorf("%s -> %s should be valid, got %s", test.From, test.To, err)
			}
		} else {
			if err == nil {
				t.Errorf("%s -> %s should be invalid, got valid", test.From, test.To)
			}
		}
	}
}
