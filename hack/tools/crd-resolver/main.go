package main

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"regexp"
	goruntime "runtime"
	"strings"
)

var (
	root string
)

func init() {
	// Get the root of the current file to use in CRD paths.
	_, filename, _, _ := goruntime.Caller(0) //nolint
	root = path.Join(path.Dir(filename), "..", "..", "..")
}

func main() {
	fmt.Printf("%s", getFilePathToAPI(root, "github.com/kube-bind", "kube-bind", "deploy/crd"))
}

func getFilePathToAPI(root, org, pkg, apis string) string {
	modBits, err := os.ReadFile(filepath.Join(root, "go.mod")) //nolint:gosec
	if err != nil {
		return ""
	}

	var packageVersion string
	packageVersionRegex := regexp.MustCompile(fmt.Sprintf(`^(\W+)%s/%s v(.+)`, org, pkg))

	for _, line := range strings.Split(string(modBits), "\n") {
		matches := packageVersionRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			packageVersion = matches[2]
			break
		}
	}

	if packageVersion == "" {
		return ""
	}

	gopath := envOr("GOPATH", build.Default.GOPATH)

	return filepath.Join(gopath, "pkg", "mod", org, fmt.Sprintf("%s@v%s", pkg, packageVersion), apis)
}

func envOr(envKey, defaultValue string) string {
	if value, ok := os.LookupEnv(envKey); ok {
		return value
	}

	return defaultValue
}
