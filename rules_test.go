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
	"strings"
	"testing"

	driver "github.com/arangodb/go-driver/v2/arangodb"
)

func TestCheckUpgradeRules(t *testing.T) {
	// Each row: From version, To version, and two booleans:
	//
	//   Allowed (3rd): true = this upgrade should be VALID (allowed)
	//                  false = this upgrade should be REJECTED (invalid)
	//
	//   Soft (4th):    true  = test with CheckSoftUpgradeRules (allows skipping minor versions)
	//                  false = test with CheckUpgradeRules (strict: minor may only +1)
	//
	// Same upgrade (e.g. 4.0.0→4.2.0) can have two rows: one with Soft=false (strict says invalid),
	// one with Soft=true (soft says valid).
	tests := []struct {
		From    driver.Version
		To      driver.Version
		Allowed bool // true = expect allowed, false = expect rejected
		Soft    bool // true = use CheckSoftUpgradeRules, false = use CheckUpgradeRules
	}{
		// --- Same version ---
		{"1.2.3", "1.2.3", true, false},
		{"1.2.3", "1.2.3", true, true},

		// --- Major: downgrade always invalid ---
		{"2.2.3", "1.2.3", false, false},
		{"2.2.3", "1.2.3", false, true},
		{"1.2.3", "2.2.3", false, false}, // 1→2 but to 2.2 not 2.0, invalid

		// --- Major: 3.x → 4.0 only with 3.12.7+ (minPatchForMajorUpgrade) ---
		// Same upgrade tested with both checkers: 4th=false → CheckUpgradeRules, 4th=true → CheckSoftUpgradeRules
		{"3.12.7", "4.0.0", true, false}, // 4th=false: run with CheckUpgradeRules (strict), expect allowed
		{"3.12.7", "4.0.0", true, true},  // 4th=true:  run with CheckSoftUpgradeRules (soft), expect allowed
		{"3.12.7-rc1", "4.0.0", true, false},
		{"3.12.7-rc1", "4.0.0", true, true},
		{"3.12.8", "4.0.0", true, false},
		{"3.12.0", "4.0.0", false, false}, // strict: patch < 7
		{"3.12.0", "4.0.0", false, true},
		{"3.12.6", "4.0.0", false, false},
		{"3.12.6", "4.0.0", false, true},
		{"3.12.rc7", "4.0.0", false, false},
		{"3.12.rc7", "4.0.0", false, true},

		{"3.11.0", "4.0.0", false, false}, // unlisted source minor must remain disallowed
		{"3.11.0", "4.0.0", false, true},
		{"3.12.7", "4.1.0", false, false}, // strict: major upgrade to non-X.0 must be rejected
		{"3.12.7", "4.1.0", false, true},
		{"3.12.0", "4.1.0", false, true},

		// --- Same major, minor: strict allows only +1 ---
		{"3.2.2", "3.3.0", true, false},
		{"3.2.2", "3.3.10", true, false},
		{"3.3.2", "3.2.10", false, false}, // minor downgrade
		{"3.3.2", "3.5.10", false, false}, // skip minor
		{"3.2.2", "3.3.11111", true, false},

		// --- Major: skipping a major (e.g., 3.x → 5.x) must be rejected ---
		{"3.12.7", "5.0.0", false, false}, // strict: major may only increment by 1
		{"3.12.7", "5.0.0", false, true},  // soft: still disallow skipping major versions

		// --- Same major & minor, patch: always allowed ---
		{"3.2.2", "3.2.88", true, false},
		{"3.2.88", "3.2.8", true, false},
		{"3.2.88", "3.2.rc7", true, false},

		// --- Soft: can skip minor within same major ---
		{"3.2.1", "3.3.1", true, true},
		{"3.2.1", "3.4.8", true, true},
		{"3.2.1", "3.5.rc7", true, true},

		// --- 4.0 → 4.2: strict invalid, soft valid ---
		{"4.0.0", "4.2.0", false, false},
		{"4.0.0", "4.2.0", true, true},
		{"4.0.7", "4.1.0", true, true},
		{"4.0.0", "4.1.0", true, true},
		{"4.0.6", "4.1.0", true, true},

		// --- Downgrade ---
		{"4.1.0", "4.0.0", false, false},
		{"4.1.0", "4.0.0", false, true},
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

func TestMalformedSourcePatchForMajorUpgrade(t *testing.T) {
	tests := []struct {
		name  string
		from  driver.Version
		to    driver.Version
		check func(driver.Version, driver.Version) error
	}{
		{
			name:  "strict checker rejects non-numeric patch component",
			from:  "3.12.rc1",
			to:    "4.0.0",
			check: CheckUpgradeRules,
		},
		{
			name:  "soft checker rejects non-numeric patch component",
			from:  "3.12.rc1",
			to:    "4.0.0",
			check: CheckSoftUpgradeRules,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.check(test.from, test.to)
			if err == nil {
				t.Fatalf("expected invalid version format error, got nil")
			}
			if !strings.Contains(err.Error(), "Invalid source version format") {
				t.Fatalf("expected invalid source version format error, got: %s", err)
			}
		})
	}
}
