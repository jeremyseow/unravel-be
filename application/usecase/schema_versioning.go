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
