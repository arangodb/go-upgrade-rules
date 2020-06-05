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
		Soft bool
	}{
		// Same version
		{"1.2.3", "1.2.3", true, false},
		// Different major
		{"2.2.3", "1.2.3", false, false},
		{"1.2.3", "2.2.3", false, false},
		// Same major, different minor
		{"3.2.2", "3.3.0", true, false},
		{"3.2.2", "3.3.10", true, false},
		{"3.3.2", "3.2.10", false, false},
		{"3.3.2", "3.5.10", false, false},
		{"3.2.2", "3.3.11111", true, false},
		// Same major & minor, different patch
		{"3.2.2", "3.2.88", true, false},
		{"3.2.88", "3.2.8", true, false},
		{"3.2.88", "3.2.rc7", true, false},
		// Soft
		{"3.2.1", "3.3.1", true, true},
		{"3.2.1", "3.4.8", true, true},
		{"3.2.1", "3.5.rc7", true, true},
	}
	for _, test := range tests {
		checkUpgradeRules := CheckUpgradeRules
		checkUpgradeRulesWithLicense := CheckUpgradeRulesWithLicense

		if test.Soft {
			checkUpgradeRules = CheckSoftUpgradeRules
			checkUpgradeRulesWithLicense = CheckSoftUpgradeRulesWithLicense
		}

		// Without license
		{
			err := checkUpgradeRules(test.From, test.To)
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
		// Same license
		{
			err := checkUpgradeRulesWithLicense(test.From, test.To, LicenseCommunity, LicenseCommunity)
			if test.Allowed {
				if err != nil {
					t.Errorf("(C->C) %s -> %s should be valid, got %s", test.From, test.To, err)
				}
			} else {
				if err == nil {
					t.Errorf("(C->C) %s -> %s should be invalid, got valid", test.From, test.To)
				}
			}
		}
		{
			err := checkUpgradeRulesWithLicense(test.From, test.To, LicenseEnterprise, LicenseEnterprise)
			if test.Allowed {
				if err != nil {
					t.Errorf("(E->E) %s -> %s should be valid, got %s", test.From, test.To, err)
				}
			} else {
				if err == nil {
					t.Errorf("(E->E) %s -> %s should be invalid, got valid", test.From, test.To)
				}
			}
		}
		// Community license -> Enterprise
		{
			err := checkUpgradeRulesWithLicense(test.From, test.To, LicenseCommunity, LicenseEnterprise)
			if test.Allowed {
				if err != nil {
					t.Errorf("(C->E) %s -> %s should be valid, got %s", test.From, test.To, err)
				}
			} else {
				if err == nil {
					t.Errorf("(C->E) %s -> %s should be invalid, got valid", test.From, test.To)
				}
			}
		}
		// Enterprise license -> Community
		{
			err := checkUpgradeRulesWithLicense(test.From, test.To, LicenseEnterprise, LicenseCommunity)
			if err == nil {
				t.Errorf("(E->C) %s -> %s should be invalid, got valid", test.From, test.To)
			}
		}
	}
}
