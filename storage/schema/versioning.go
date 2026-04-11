package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
)

// paramInfo captures the structural aspects of a parameter that determine
// whether a schema change is breaking, additive, or documentation-only.
// Note: data_type lives in entity_parameters (the catalog), not in the
// mapping table, so type-change detection is intentionally out of scope here
// and is enforced separately through parameter update rules.
type paramInfo struct {
	isRequired bool
}

// toParamMap converts a slice of mapping rows into a key→paramInfo map
// for easy comparison between two schema versions.
func toParamMap(mappings []model.EntitySchemasParametersMappings) map[string]paramInfo {
	m := make(map[string]paramInfo, len(mappings))
	for _, p := range mappings {
		m[p.ParameterKey] = paramInfo{
			isRequired: p.IsRequired != nil && *p.IsRequired,
		}
	}
	return m
}

// determineVersionBump compares the parameter sets of the current active
// version against the incoming draft and returns the required semver bump:
//
//	"major" — breaking change: parameter removed, new required parameter added,
//	           or existing parameter changed from optional → required.
//	"minor" — backward-compatible addition: new optional parameter added,
//	           or existing parameter relaxed from required → optional.
//	"patch" — no structural change; only schema/parameter metadata was edited.
func determineVersionBump(current, next map[string]paramInfo) string {
	isMajor := false
	isMinor := false

	// Removed parameters break consumers that depend on them.
	for key := range current {
		if _, exists := next[key]; !exists {
			isMajor = true
		}
	}

	for key, nextP := range next {
		curP, exists := current[key]
		if !exists {
			// New parameter: required = breaking for producers, optional = safe.
			if nextP.isRequired {
				isMajor = true
			} else {
				isMinor = true
			}
		} else {
			// Existing parameter: tightening is breaking, relaxing is additive.
			if !curP.isRequired && nextP.isRequired {
				isMajor = true // optional → required: existing producers may not send it
			} else if curP.isRequired && !nextP.isRequired {
				isMinor = true // required → optional: safe relaxation
			}
		}
	}

	switch {
	case isMajor:
		return "major"
	case isMinor:
		return "minor"
	default:
		return "patch"
	}
}

// applyBump increments the appropriate component of a semver string and
// resets lower components to zero per the semver spec.
func applyBump(version, bump string) string {
	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	default: // patch
		patch++
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// sortByVersion sorts SchemaRecords newest-first using semver comparison.
// Handles major.minor.patch today; compareSemver already supports pre-release
// suffixes (e.g. 1.0.0-rc.1) for when the CHECK constraint is relaxed.
func sortByVersion(records []SchemaRecord) {
	sort.Slice(records, func(i, j int) bool {
		return compareSemver(records[i].SchemaVersion, records[j].SchemaVersion) > 0
	})
}

// compareSemver returns 1 if a > b, -1 if a < b, 0 if equal.
// Release versions (no pre-release suffix) rank above any pre-release.
func compareSemver(a, b string) int {
	parseVersion := func(v string) (major, minor, patch int, pre string) {
		parts := strings.SplitN(v, "-", 2)
		nums := strings.Split(parts[0], ".")
		toInt := func(s string) int { n, _ := strconv.Atoi(s); return n }
		if len(nums) > 0 {
			major = toInt(nums[0])
		}
		if len(nums) > 1 {
			minor = toInt(nums[1])
		}
		if len(nums) > 2 {
			patch = toInt(nums[2])
		}
		if len(parts) == 2 {
			pre = parts[1]
		}
		return
	}

	aMaj, aMin, aPat, aPre := parseVersion(a)
	bMaj, bMin, bPat, bPre := parseVersion(b)

	for _, pair := range [3][2]int{{aMaj, bMaj}, {aMin, bMin}, {aPat, bPat}} {
		if pair[0] != pair[1] {
			if pair[0] > pair[1] {
				return 1
			}
			return -1
		}
	}

	switch {
	case aPre == "" && bPre != "":
		return 1
	case aPre != "" && bPre == "":
		return -1
	case aPre != bPre:
		if aPre > bPre {
			return 1
		}
		return -1
	}
	return 0
}
