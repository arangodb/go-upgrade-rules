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

// CheckUpgradeRules checks if it is allowed to upgrade an ArangoDB
// deployment from given `from` version to given `to` version.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
func CheckUpgradeRules(from, to driver.Version) error {
	// Image changed, check if change is allowed
	if from.Major() != to.Major() {
		// E.g. 3.x -> 4.x, we cannot allow automatically
		return fmt.Errorf("Major versions are different")
	}
	if from.Minor() != to.Minor() {
		// Only allow upgrade from 3.x to 3.y when y=x+1
		if from.Minor()+1 != to.Minor() {
			return fmt.Errorf("Minor versions may only increment by 1")
		}
	} else {
		// Patch version only diff. That is allowed in upgrade & downgrade.
	}
	return nil
}

// CheckSoftUpgradeRules checks if it is allowed to upgrade an ArangoDB
// deployment from given `from` version to given `to` version.
// If this is allowed, nil is returned, otherwise and error is
// returning describing why the upgrade is not allowed.
// This function allows to jump more than one minor version.
func CheckSoftUpgradeRules(from, to driver.Version) error {
	// Image changed, check if change is allowed
	if from.Major() != to.Major() {
		// E.g. 3.x -> 4.x, we cannot allow automatically
		return fmt.Errorf("Major versions are different")
	}
	if from.Minor() != to.Minor() {
		// Only allow upgrade from 3.x to 3.y when y=x+1
		if from.Minor() < to.Minor() {
			return fmt.Errorf("Downgrade is not possible")
		}
	} else {
		// Patch version only diff. That is allowed in upgrade & downgrade.
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
