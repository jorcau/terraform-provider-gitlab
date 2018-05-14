package gitlab

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	gitlab "github.com/xanzy/go-gitlab"
)

var accessLevelNameToValue = map[string]gitlab.AccessLevelValue{
	"guest":     gitlab.GuestPermissions,
	"reporter":  gitlab.ReporterPermissions,
	"developer": gitlab.DeveloperPermissions,
	"master":    gitlab.MasterPermissions,
	"owner":     gitlab.OwnerPermission,
}

var accessLevelID = map[string]gitlab.AccessLevelValue{
	"guest":     gitlab.GuestPermissions,
	"reporter":  gitlab.ReporterPermissions,
	"developer": gitlab.DeveloperPermissions,
	"master":    gitlab.MasterPermissions,
	"owner":     gitlab.OwnerPermission,
}

var accessLevelValueToName = map[gitlab.AccessLevelValue]string{
	gitlab.GuestPermissions:     "guest",
	gitlab.ReporterPermissions:  "reporter",
	gitlab.DeveloperPermissions: "developer",
	gitlab.MasterPermissions:    "master",
	gitlab.OwnerPermission:      "owner",
}

var accessLevel = map[gitlab.AccessLevelValue]string{
	gitlab.GuestPermissions:     "guest",
	gitlab.ReporterPermissions:  "reporter",
	gitlab.DeveloperPermissions: "developer",
	gitlab.MasterPermissions:    "master",
	gitlab.OwnerPermission:      "owner",
}

// copied from ../github/util.go
func validateValueFunc(values []string) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (we []string, errors []error) {
		value := v.(string)
		valid := false
		for _, role := range values {
			if value == role {
				valid = true
				break
			}
		}

		if !valid {
			errors = append(errors, fmt.Errorf("%s is an invalid value for argument %s", value, k))
		}
		return
	}
}

func stringToVisibilityLevel(s string) *gitlab.VisibilityValue {
	lookup := map[string]gitlab.VisibilityValue{
		"private":  gitlab.PrivateVisibility,
		"internal": gitlab.InternalVisibility,
		"public":   gitlab.PublicVisibility,
	}

	value, ok := lookup[s]
	if !ok {
		return nil
	}
	return &value
}
