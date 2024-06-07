package version

import (
	"encoding/json"
	"io"
	"net/http"

	_ "embed"
)

const (
	GH_API_URL = "https://api.github.com/repos/adde/kade"
)

//go:embed version.txt
var CurrentVersion string
var LatestVersion = GetLatestVersion()

func IsLatestVersion() bool {
	return CurrentVersion >= LatestVersion
}

func GetLatestVersion() string {
	res, err := http.Get(GH_API_URL + "/releases/latest")
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	err = json.Unmarshal(body, &release)
	if err != nil {
		return ""
	}

	return release.TagName
}
