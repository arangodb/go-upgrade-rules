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
	"fmt"
	"strconv"
	"strings"

	driver "github.com/arangodb/go-driver"
)

// License is a strongly typed ArangoDB license type
type License int

const (
	// LicenseCommunity is this license for the ArangoDB community edition
	LicenseCommunity License = iota
	// LicenseEnterprise is this license for the ArangoDB enterprise edition
	LicenseEnterprise
)

// minPatchForMajorUpgrade defines the minimum source patch version required
// for an upgrade to the next major X.0. Key is "fromMajor.fromMinor" (e.g. "3.12").
// Example: 3.12.7+ is required to upgrade to 4.0.
var minPatchForMajorUpgrade = map[string]int{
	"3.12": 7, // 3.12.x → 4.0 requires at least 3.12.7
}

// getPatch extracts the numeric patch version (third component) from a driver.Version.
//
// driver.Version does not expose a Patch() method, so we parse the string form.
// Examples of supported formats:
//
//	"3.12.7"        → 7
//	"3.12.7-rc1"    → 7
//	"3.12.7rc1"     → 7
//
// If the version does not contain a valid numeric patch component,
// the function safely returns 0.
func getPatch(v driver.Version) int {
	// Convert version to string (e.g. "3.12.7" or "3.12.7-rc1")
	s := fmt.Sprint(v)

	// Split by dot: ["3", "12", "7"] or ["3", "12", "7-rc1"]
	parts := strings.Split(s, ".")

	// If there are fewer than 3 components, patch is not present
	if len(parts) < 3 {
		return 0
	}

	// Extract the patch component (may include suffix like "-rc1")
	patchStr := parts[2]

	// Strip any non-digit suffix (e.g. "7-rc1" → "7")
	// We stop at the first non-numeric character.
	for i, r := range patchStr {
		if r < '0' || r > '9' {
			patchStr = patchStr[:i]
			break
		}
	}

	// Convert cleaned numeric string to integer
	p, err := strconv.Atoi(patchStr)
	if err != nil {
		// If conversion fails, return 0 as a safe fallback
		return 0
	}

	return p
}

// checkMinPatchForMajorUpgrade returns an error if from→to is a major upgrade
// that requires a minimum patch and from does not satisfy it.
func checkMinPatchForMajorUpgrade(from, to driver.Version) error {
	if to.Major() != from.Major()+1 || to.Minor() != 0 {
		return nil
	}
	key := fmt.Sprintf("%d.%d", from.Major(), from.Minor())
	minPatch, hasRule := minPatchForMajorUpgrade[key]
	if !hasRule {
		return fmt.Errorf("Major versions are different: major upgrade from %d.%d to %d.0 is not allowed", from.Major(), from.Minor(), to.Major())
	}
	if getPatch(from) >= minPatch {
		return nil
	}
	return fmt.Errorf("Upgrade to %d.0 requires at least %d.%d.%d", to.Major(), from.Major(), from.Minor(), minPatch)
}

// CheckUpgradeRules checks if it is allowed to upgrade an ArangoDB
// deployment from given `from` version to given `to` version.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
func CheckUpgradeRules(from, to driver.Version) error {

	// ---- Major version change ----
	if from.Major() != to.Major() {

		// Disallow downgrade
		if to.Major() < from.Major() {
			return fmt.Errorf("Major versions are different: downgrade of major version is not allowed")
		}

		// Only allow upgrade to next major version
		if to.Major() != from.Major()+1 {
			return fmt.Errorf("Major versions are different: major versions may only increment by 1")
		}

		// Only allow upgrade to X.0 (e.g. 3.12.x → 4.0.x)
		if to.Minor() != 0 {
			return fmt.Errorf("Major versions are different: major upgrades are only allowed to X.0")
		}

		if err := checkMinPatchForMajorUpgrade(from, to); err != nil {
			return err
		}
		return nil
	}

	// ---- Minor version change (same major) ----
	if from.Minor() != to.Minor() {

		// Disallow downgrade
		if to.Minor() < from.Minor() {
			return fmt.Errorf("Downgrade of minor version is not allowed")
		}

		// Only allow increment by 1
		if to.Minor() != from.Minor()+1 {
			return fmt.Errorf("Minor versions may only increment by 1")
		}
	}

	// Patch changes always allowed
	return nil
}

// CheckSoftUpgradeRules checks if it is allowed to upgrade an ArangoDB
// deployment from given `from` version to given `to` version.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
// This function allows to jump more than one minor version.
func CheckSoftUpgradeRules(from, to driver.Version) error {

	// ---- Major version change ----
	if from.Major() != to.Major() {

		// Disallow major downgrade
		if to.Major() < from.Major() {
			return fmt.Errorf("Major versions are different: downgrade of major version is not allowed")
		}

		// Only allow upgrade to next major version
		if to.Major() != from.Major()+1 {
			return fmt.Errorf("Major versions are different: major versions may only increment by 1")
		}

		// Only allow upgrade to X.0
		if to.Minor() != 0 {
			return fmt.Errorf("Major versions are different: major upgrades are only allowed to X.0")
		}

		// Enforce minimum patch requirement (e.g. 3.12.7 → 4.0)
		if err := checkMinPatchForMajorUpgrade(from, to); err != nil {
			return err
		}

		return nil
	}

	// ---- Same major version ----
	// Only major/minor rules: no minimum-patch for minor upgrades.
	// (Minimum-patch applies only to major upgrade 3.12.7+ → 4.0.)

	if to.Minor() < from.Minor() {
		return fmt.Errorf("Downgrade of minor version is not allowed")
	}

	return nil
}

// CheckUpgradeRulesWithLicense checks if it is allowed to upgrade an ArangoDB
// deployment from given `fromVersion` version to given `toVersion` version.
// If also includes the given `fromLicense` and `toLicense` in this check.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
func CheckUpgradeRulesWithLicense(fromVersion, toVersion driver.Version, fromLicense, toLicense License) error {
	if fromLicense != toLicense && fromLicense == LicenseEnterprise {
		return fmt.Errorf("Upgrade from Enterprise to Community edition is not possible")
	}
	return CheckUpgradeRules(fromVersion, toVersion)
}

// CheckUpgradeRulesWithLicense checks if it is allowed to upgrade an ArangoDB
// deployment from given `fromVersion` version to given `toVersion` version.
// If also includes the given `fromLicense` and `toLicense` in this check.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
// This function allows to jump more than one minor version.
func CheckSoftUpgradeRulesWithLicense(fromVersion, toVersion driver.Version, fromLicense, toLicense License) error {
	if fromLicense != toLicense && fromLicense == LicenseEnterprise {
		return fmt.Errorf("Upgrade from Enterprise to Community edition is not possible")
	}
	return CheckSoftUpgradeRules(fromVersion, toVersion)
}
