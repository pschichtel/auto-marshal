package util

import (
	"golang.org/x/mod/modfile"
	"os"
	"path/filepath"
)

func ResolvedPwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return resolvePath(pwd)
}

func resolvePath(path string) string {
	resolvedWithoutLinks, err := filepath.EvalSymlinks(path)
	if err != nil {
		panic(err)
	}

	resolved, err := filepath.Abs(resolvedWithoutLinks)
	if err != nil {
		panic(err)
	}

	return resolved
}

func relativize(base string, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		panic(err)
	}

	return rel
}

func FindModuleRoot(path string) string {
	modPath := filepath.Join(path, "go.mod")
	_, err := os.Stat(modPath)
	if err != nil {
		nextPath := resolvePath(filepath.Join(path, ".."))
		if nextPath == path {
			return path
		}
		return FindModuleRoot(nextPath)
	} else {
		return path
	}
}

func detectModule(moduleRoot string) string {
	modFile := filepath.Join(moduleRoot, "go.mod")
	content, err := os.ReadFile(modFile)
	if err != nil {
		panic(err)
	}
	file, err := modfile.Parse(modFile, content, func(path, version string) (string, error) {
		return version, nil
	})
	if err != nil {
		panic(err)
	}
	return file.Module.Mod.Path
}

func DetectPackagePath(pwd string, moduleRoot string) string {
	modulePath := detectModule(moduleRoot)
	relativePackagePath := relativize(moduleRoot, pwd)
	if relativePackagePath == "." {
		return modulePath
	}
	return modulePath + "/" + relativePackagePath
}
