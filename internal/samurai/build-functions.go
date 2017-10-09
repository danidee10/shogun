package samurai

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/influx6/faux/exec"
	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/vfiles"
	"github.com/influx6/gobuild/build"
	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
	"github.com/influx6/shogun/internal"
	"github.com/influx6/shogun/templates"
)

var (
	ignoreAddition = ".shogun"
	cmdDir         = "cmd"
	goosRuntime    = runtime.GOOS
	packageReg     = regexp.MustCompile(`package \w+`)
)

// BuildFunctions holds all build directives from processed packages.
type BuildFunctions struct {
	Dir  string
	Main BuildList
	Subs map[string]BuildList
}

// BuildPackage builds a shogun binarie commandline files for giving directory and 1 level directory.
func BuildPackage(vlog metrics.Metrics, events metrics.Metrics, dir string, binaryPath string, skipBuild bool, ctx build.Context) (BuildFunctions, error) {
	var list BuildFunctions
	list.Subs = make(map[string]BuildList)

	var err error
	list.Main, err = BuildPackageForDir(vlog, events, dir, binaryPath, cmdDir, skipBuild, ctx)
	if err != nil {
		events.Emit(metrics.Error(err).With("dir", dir).With("binary_path", binaryPath))
		return list, err
	}

	if err := vfiles.WalkDirSurface(dir, func(rel string, abs string, info os.FileInfo) error {
		if !info.IsDir() {
			return nil
		}

		res, err2 := BuildPackageForDir(vlog, events, abs, cmdDir, binaryPath, skipBuild, ctx)
		if err2 != nil {
			events.Emit(metrics.Error(err2).With("dir", abs).With("binary_path", binaryPath))
			return err2
		}

		list.Subs[rel] = res
		return nil
	}); err != nil {
		events.Emit(metrics.Error(err).With("dir", dir))
		return list, err
	}

	return list, nil
}

// BuildList holds a procssed package list of write directives.
type BuildList struct {
	Hash        string
	Path        string
	PkgPath     string
	PkgFilePath string
	List        []gen.WriteDirective
	Functions   []internal.Function
}

// BuildPackageForDir generates needed package files for creating new function based executable binaries.
func BuildPackageForDir(vlog metrics.Metrics, events metrics.Metrics, dir string, cmd string, binaryPath string, skipBuild bool, ctx build.Context) (BuildList, error) {
	var list BuildList
	list.Path = dir

	pkgs, err := ast.FilteredPackageWithBuildCtx(vlog, dir, ctx)
	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			return list, nil
		}

		return list, err
	}

	var hash []byte

	for _, pkgItem := range pkgs {
		pkgHash, err := generateHash(pkgItem.Files)
		if err != nil {
			return list, err
		}

		hash = append(hash, []byte(pkgHash)...)

		var binaryName, binaryExeName string
		if binAnnons := pkgItem.AnnotationsFor("@binaryName"); len(binAnnons) != 0 {
			if len(binAnnons[0].Arguments) == 0 {
				err := fmt.Errorf("InvalidBinaryName(File: %q): expected format @binaryName(name => NAME)", pkgItem.FilePath)
				return list, err
			}

			binaryName = strings.ToLower(binAnnons[0].Param("name"))
		} else {
			binaryName = pkgItem.Name
		}

		binaryExeName = binaryName
		if goosRuntime == "windows" {
			binaryExeName = fmt.Sprintf("%s.exec", binaryName)
		}

		packageBinaryPath := filepath.Join(cmd, binaryName)
		packageBinaryFilePath := filepath.Join(cmd, binaryName, "pkg")

		for _, declr := range pkgItem.Packages {

			// Retrieve function list.
			fnsList, err := pullFunctionFromDeclr(pkgItem, &declr)
			if err != nil {
				return list, err
			}

			list.Functions = append(list.Functions, fnsList...)

			source := strings.Replace(declr.Source, strings.Join(declr.Comments, "\n"), "", -1)
			packageIndex := strings.Index(source, "package")
			packagePart := packageReg.FindString(source)

			source = source[packageIndex:]
			source = strings.TrimSpace(strings.Replace(source, packagePart, "", 1))

			list.List = append(list.List, gen.WriteDirective{
				FileName: filepath.Base(declr.FilePath),
				Dir:      packageBinaryFilePath,
				Writer: gen.SourceTextWith(
					string(templates.Must("shogun-src.tml")),
					template.FuncMap{},
					struct {
						Source string
					}{
						Source: source,
					},
				),
			})
		}

		list.List = append(list.List, gen.WriteDirective{
			FileName: fmt.Sprintf("pkg_%s.go", binaryName),
			Dir:      packageBinaryFilePath,
			Writer: gen.SourceTextWith(
				string(templates.Must("shogun-src-pkg.tml")),
				template.FuncMap{},
				struct {
					BinaryName string
					Functions  []internal.Function
				}{
					BinaryName: binaryName,
					Functions:  list.Functions,
				},
			),
		})

		list.List = append(list.List, gen.WriteDirective{
			FileName: "main.go",
			Dir:      packageBinaryPath,
			Writer: gen.SourceTextWith(
				string(templates.Must("shogun-main.tml")),
				template.FuncMap{},
				struct {
				}{},
			),
			After: func() error {
				if skipBuild {
					return nil
				}

				fmt.Printf("----------------------------------------\n")
				fmt.Printf("Building binary for shogunate: %q\n", binaryName)

				if err := exec.New(exec.Command("go build -x -o %s %s", filepath.Join(binaryPath, binaryExeName), filepath.Join(dir, packageBinaryPath, "main.go")), exec.Async()).Exec(context.Background(), vlog); err != nil {
					fmt.Printf("Building binary for shogun %q failed\n", binaryName)
					return err
				}

				fmt.Printf("Built binary for shogun %q into %q\n", binaryName, binaryPath)

				fmt.Printf("Cleaning up shogun binary build files... %q\n", binaryName)
				if err := os.RemoveAll(filepath.Join(dir, packageBinaryPath)); err != nil {
					fmt.Printf("Failed to properly cleanup build files %q\n", binaryName)
					return err
				}

				fmt.Printf("Shogun %q build ready\n\n", binaryName)
				return nil
			},
		})
	}

	return list, nil
}

func binHash(nlog metrics.Metrics, binPath string) (string, error) {
	var response bytes.Buffer

	if err := exec.New(exec.Command("%s hash", binPath), exec.Async(), exec.Output(&response)).Exec(context.Background(), nlog); err != nil {
		return "", err
	}

	return strings.TrimSpace(response.String()), nil
}

func generateHash(files []string) (string, error) {
	var hashes []byte

	for _, file := range files {
		hash, err := generateFileHash(file)
		if err != nil {
			return "", err
		}

		hashes = append(hashes, []byte(hash)...)
	}

	return base64.StdEncoding.EncodeToString(hashes), nil
}

func generateFileHash(file string) (string, error) {
	hasher := sha1.New()
	fl, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer fl.Close()

	_, err = io.Copy(hasher, fl)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
