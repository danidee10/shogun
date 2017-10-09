package samurai

import (
	"os"

	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/vfiles"
	"github.com/influx6/gobuild/build"
	"github.com/influx6/moz/ast"
)

// PackageHashList holds a list of hashes from a main package and
// all other subpackages retrieved.
type PackageHashList struct {
	Dir  string
	Main HashList
	Subs map[string]HashList
}

// ListPackageHash returns all functions retrieved from the directory filtered by the build.Context.
func ListPackageHash(vlog, events metrics.Metrics, targetDir string, ctx build.Context) (PackageHashList, error) {
	var list PackageHashList
	list.Dir = targetDir
	list.Subs = make(map[string]HashList)

	// Build shogunate directory itself first.
	var err error
	list.Main, err = HashPackages(vlog, events, targetDir, ctx)
	if err != nil {
		events.Emit(metrics.Errorf("Failed to generate function list : %+q", err))
		return list, err
	}

	if err = vfiles.WalkDirSurface(targetDir, func(rel string, abs string, info os.FileInfo) error {
		if !info.IsDir() {
			return nil
		}

		res, err2 := HashPackages(vlog, events, abs, ctx)
		if err2 != nil {
			return err2
		}

		list.Subs[rel] = res
		return nil
	}); err != nil {
		events.Emit(metrics.Error(err).With("dir", targetDir))
		return list, err
	}

	return list, nil
}

// HashList holds the list of processed functions from individual packages.
type HashList struct {
	Path     string
	Hash     string
	Packages map[string]string
}

// HashPackages iterates all directories and generates package hashes of all declared functions
// matching the shegun format.
func HashPackages(vlog, events metrics.Metrics, dir string, ctx build.Context) (HashList, error) {
	var pkgFuncs HashList
	pkgFuncs.Path = dir
	pkgFuncs.Packages = make(map[string]string)

	pkgs, err := ast.FilteredPackageWithBuildCtx(vlog, dir, ctx)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			return pkgFuncs, nil
		}

		events.Emit(metrics.Error(err).With("dir", dir))
		return pkgFuncs, err
	}

	var hash []byte

	for _, pkgItem := range pkgs {
		pkgHash, err := generateHash(pkgItem.Files)
		if err != nil {
			return pkgFuncs, err
		}

		pkgFuncs.Packages[pkgItem.Path] = pkgHash
		hash = append(hash, []byte(pkgHash)...)
	}

	pkgFuncs.Hash = string(hash)

	return pkgFuncs, nil
}
