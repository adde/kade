package version

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	GH_API_URL = "https://api.github.com/repos/adde/kade"
)

var CurrentVersion = GetCurrentVersion()
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

func GetCurrentVersion() string {
	file, err := os.Open("version.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		return line
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}
