package profile

import "os"

const (
	profileEnvName = "profile"
	ProductEnvName = "product"
)

var currentProfile string

func init() {
	currentProfile = os.Getenv(profileEnvName)
}

func IsProd() bool {
	if currentProfile == ProductEnvName {
		return true
	}
	return false
}

func GetProfile() string {
	return currentProfile
}
