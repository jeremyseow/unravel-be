package usecase

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jeremyseow/unravel-be/application/domain"
)

const initialVersion = "1.0.0"

func computeNextVersion(latestActive *domain.Schema, draft domain.Schema) (string, error) {
	if latestActive == nil {
		return initialVersion, nil
	}
	bump := determineBump(latestActive.Parameters, draft.Parameters)
	return applyBump(latestActive.SchemaVersion, bump)
}

func determineBump(old, next []domain.SchemaParameter) string {
	oldMap := toParamMap(old)
	nextMap := toParamMap(next)

	bump := "patch"

	for key, oldRequired := range oldMap {
		nextRequired, exists := nextMap[key]
		if !exists {
			return "major" // parameter removed
		}
		if !oldRequired && nextRequired {
			return "major" // optional → required
		}
		if oldRequired && !nextRequired && bump != "major" {
			bump = "minor" // required → optional
		}
	}

	for key, nextRequired := range nextMap {
		if _, exists := oldMap[key]; !exists {
			if nextRequired {
				return "major" // new required parameter
			}
			if bump != "major" {
				bump = "minor" // new optional parameter
			}
		}
	}

	return bump
}

func applyBump(version, bump string) (string, error) {
	major, minor, patch, err := parseSemver(version)
	if err != nil {
		return "", err
	}
	switch bump {
	case "major":
		return fmt.Sprintf("%d.0.0", major+1), nil
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1), nil
	default:
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1), nil
	}
}

func parseSemver(v string) (major, minor, patch int, err error) {
	core := strings.SplitN(v, "-", 2)[0] // strip pre-release
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver: %q", v)
	}
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return
	}
	if minor, err = strconv.Atoi(parts[1]); err != nil {
		return
	}
	patch, err = strconv.Atoi(parts[2])
	return
}

func toParamMap(params []domain.SchemaParameter) map[string]bool {
	m := make(map[string]bool, len(params))
	for _, p := range params {
		m[p.ParameterKey] = p.IsRequired
	}
	return m
}

// annotateChanges sorts schemas by semver ascending (draft last) and attaches a
// SchemaChanges diff to each version that has a predecessor.
func annotateChanges(schemas []domain.Schema) []domain.Schema {
	if len(schemas) < 2 {
		return schemas
	}

	published := make([]domain.Schema, 0, len(schemas))
	var draft *domain.Schema
	for i := range schemas {
		if schemas[i].SchemaVersion == domain.DraftVersion {
			copy := schemas[i]
			draft = &copy
		} else {
			published = append(published, schemas[i])
		}
	}

	// Sort published versions oldest → newest by semver.
	sortSemverAsc(published)

	// Compute changes for each published version after the first.
	for i := 1; i < len(published); i++ {
		changes := diffParams(published[i-1].Parameters, published[i].Parameters)
		published[i].Changes = &changes
	}

	result := published

	// Draft changes are computed vs the latest published version.
	if draft != nil {
		if len(published) > 0 {
			changes := diffParams(published[len(published)-1].Parameters, draft.Parameters)
			draft.Changes = &changes
		}
		result = append(result, *draft)
	}

	return result
}

func diffParams(old, next []domain.SchemaParameter) domain.SchemaChanges {
	oldMap := toParamMap(old)
	nextMap := toParamMap(next)

	var added, removed, promoted, demoted []string

	for key, nextRequired := range nextMap {
		oldRequired, existed := oldMap[key]
		if !existed {
			added = append(added, key)
			continue
		}
		if !oldRequired && nextRequired {
			promoted = append(promoted, key)
		} else if oldRequired && !nextRequired {
			demoted = append(demoted, key)
		}
	}
	for key := range oldMap {
		if _, exists := nextMap[key]; !exists {
			removed = append(removed, key)
		}
	}

	bump := determineBump(old, next)
	return domain.SchemaChanges{
		BumpType:       bump,
		AddedParams:    added,
		RemovedParams:  removed,
		PromotedParams: promoted,
		DemotedParams:  demoted,
	}
}

func sortSemverAsc(schemas []domain.Schema) {
	for i := 1; i < len(schemas); i++ {
		for j := i; j > 0 && semverLess(schemas[j].SchemaVersion, schemas[j-1].SchemaVersion); j-- {
			schemas[j], schemas[j-1] = schemas[j-1], schemas[j]
		}
	}
}

// semverLess returns true when a is an earlier release than b.
func semverLess(a, b string) bool {
	ma, mia, pa, err1 := parseSemver(a)
	mb, mib, pb, err2 := parseSemver(b)
	if err1 != nil || err2 != nil {
		return a < b
	}
	if ma != mb {
		return ma < mb
	}
	if mia != mib {
		return mia < mib
	}
	return pa < pb
}
