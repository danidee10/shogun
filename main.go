package main

//go:generate go generate ./templates/...

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	gexec "os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/influx6/faux/exec"
	"github.com/influx6/faux/metrics"
	"github.com/influx6/faux/metrics/custom"
	"github.com/influx6/gobuild/build"
	"github.com/influx6/moz/ast"
	"github.com/influx6/moz/gen"
	"github.com/influx6/shogun/internals/samurai"
	"github.com/influx6/shogun/templates"
	"github.com/minio/cli"
)

// vars
var (
	Version          = "0.0.1"
	shogunateDirName = "katanas"
	ignoreAddition   = ".shogun"
	goPath           = os.Getenv("GOPATH")
	goSrcPath        = filepath.Join(goPath, "src")
	goosRuntime      = runtime.GOOS
	packageReg       = regexp.MustCompile(`package \w+`)
	binNameReg       = regexp.MustCompile("\\W+")
	helpTemplate     = `NAME:
{{.Name}} - {{.Usage}}

VERSION:
{{.Version}}

DESCRIPTION:
{{.Description}}

USAGE:
{{.Name}} {{if .Flags}}[flags] {{end}}command{{if .Flags}}{{end}} [arguments...]

COMMANDS:
	{{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
	{{end}}{{if .Flags}}
FLAGS:
	{{range .Flags}}{{.}}
	{{end}}{{end}}

`
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Name = "Shogun"
	app.Author = "Ewetumo Alexander"
	app.Email = "trinoxf@gmail.com"
	app.Usage = "shogun {{command}}"
	app.Description = "Become one with your functions"
	app.CustomAppHelpTemplate = helpTemplate
	app.Action = mainAction
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "in,input",
			Usage: "-in=bob-build to give a one time value to a function as input",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "add",
			Action: addAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir,dirName",
					Usage: "-dirName=bob-build set the name of directory and package",
				},
				cli.BoolFlag{
					Name:  "m,main",
					Usage: "-main to force main as the package name",
				},
				cli.BoolFlag{
					Name:  "v,verbose",
					Usage: "-verbose to show hidden logs and operations",
				},
			},
		},
		{
			Name:   "list",
			Action: listAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d,dir",
					Usage: "-dir=./example to set directory to scan for functions",
				},
				cli.BoolFlag{
					Name:  "v,verbose",
					Usage: "-verbose to show hidden logs and operations",
				},
			},
		},
		{
			Name:   "build",
			Action: buildAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "d,dir",
					Value: "",
					Usage: "-dir=./katanas to build specific directory instead of root.",
				},
				cli.StringFlag{
					Name:  "cmd,cmdDir",
					Value: "",
					Usage: "-cmd=./cmd to build CLI package files into relative directory.",
				},
				cli.BoolFlag{
					Name:  "rm,remove",
					Usage: "-rm to delete package files after building binaries",
				},
				cli.BoolFlag{
					Name:  "single,singlepkg",
					Usage: "-singlePkg=true to only bundle seperate command binaries",
				},
				cli.BoolFlag{
					Name:  "skipsub,skipsubcommandbuild",
					Usage: "-skipsub to only generate combined binary for root and not for sub packages",
				},
				cli.BoolFlag{
					Name:  "skip,skipbuild",
					Usage: "-skip to generate CLI package files without building binaries",
				},
				cli.BoolFlag{
					Name:  "v,verbose",
					Usage: "-verbose to show hidden logs and operations",
				},
				cli.BoolFlag{
					Name:  "f,force",
					Usage: "-f to force rebuild of binary",
				},
			},
		},
		{
			Name:   "help",
			Action: helpAction,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "v,verbose",
					Usage: "-verbose to show hidden logs and operations",
				},
				cli.BoolFlag{
					Name:  "s,source",
					Usage: "-source to show source of binary function",
				},
			},
		},
		{
			Name:   "version",
			Action: versionAction,
			Flags:  []cli.Flag{},
		},
	}

	app.RunAndExitOnError()
}

func addAction(c *cli.Context) error {
	if c.NArg() == 0 {
		fmt.Printf("⠙ Run `shogun add bob.go` to add `bob.go` to root directory\n")
		fmt.Printf("⠙ Run `shogun add -dir=voz bob.go` to add `bob.go` to `voz` directory\n")
		fmt.Printf("⠙ Run `shogun add -dir=voz bob.go dog.go` to add `bob.go` and `dog.go` to `voz` directory\n")
		fmt.Printf("⠙ Run `shogun add -dir=voz bob.go ...[filenames]` to add more files to `voz` directory\n\n")

		fmt.Println("⡿ Run `shogun -h` to see what it takes to make me work.")
		return nil
	}

	events := metrics.New()

	if c.Bool("verbose") {
		events = metrics.New(custom.StackDisplay(os.Stdout))
	}

	currentDir, err := os.Getwd()
	if err != nil {
		events.Emit(metrics.Errorf("Failed to read current directory: %q", err))
		return err
	}

	var packageName string

	packageDir := c.String("dirName")
	if packageDir == "" && c.Bool("main") {
		packageName = "main"
	}

	if packageDir != "" && !c.Bool("main") {
		packageName = toPackageName(packageDir)
	}

	if packageDir == "" && !c.Bool("main") {
		packageName = filepath.Base(currentDir)
	}

	var directives []gen.WriteDirective

	for i := 0; i < c.NArg(); i++ {
		arg := c.Args().Get(i)
		if arg == "" {
			continue
		}

		arg = strings.TrimSuffix(arg, filepath.Ext(arg))
		fileName := fmt.Sprintf("%s.go", arg)

		directives = append(directives, gen.WriteDirective{
			Dir:      packageDir,
			FileName: fileName,
			Writer: gen.SourceTextWith(
				string(templates.Must("shogun-add.tml")),
				template.FuncMap{},
				struct {
					Package string
				}{
					Package: packageName,
				},
			),
		})

	}

	if err := ast.SimpleWriteDirectives("./", false, directives...); err != nil {
		events.Emit(metrics.Errorf("Failed to create new file: %q", err))
		return err
	}

	return nil
}

func helpAction(c *cli.Context) error {
	if c.NArg() == 0 || c.Args().First() == "" {
		fmt.Println("⡿ Run `shogun -h` to see what it takes to make me work.")
		return nil
	}

	events := metrics.New()

	if c.Bool("verbose") {
		events = metrics.New(custom.StackDisplay(os.Stdout))
	}

	var buildDone bool
	if _, err := gexec.LookPath(c.Args().First()); err != nil {
		if err := buildAction(c); err != nil {
			fmt.Println("⡿ Run `shogun build -dir=''` to build package directory first before running `shogun [] [Command]`.")
			return nil
		}

		buildDone = true
	}

	if !buildDone {
		if err := buildAction(c); err != nil {
			// do nothing for now
		}
	}

	binaryPath := binPath()

	var command string

	if c.Bool("source") {
		command = fmt.Sprintf("%s/%s help -s %s", binaryPath, c.Args().First(), strings.Join(c.Args().Tail(), " "))
	} else {
		command = fmt.Sprintf("%s/%s help %s", binaryPath, c.Args().First(), strings.Join(c.Args().Tail(), " "))
	}

	binCmd := exec.New(
		exec.Async(),
		exec.Command(command),
		exec.Output(os.Stdout),
		exec.Err(os.Stderr),
	)

	fmt.Printf("⡿ Executing %+q:\n", command)
	exitCode, err := binCmd.ExecWithExitCode(context.Background(), events)
	if err != nil {
		events.Emit(metrics.Error(err))
		return fmt.Errorf("Command Error: %+q", err)
	}

	if exitCode > 0 {
		return fmt.Errorf("Command Error: ExitCode: %d\n", exitCode)
	}

	return nil
}

func mainAction(c *cli.Context) error {
	if c.NArg() == 0 || c.Args().First() == "" {
		fmt.Printf("⠙ Nothing to do...\n\n")
		fmt.Println("⡿ Run `shogun -h` to see what it takes to make me work.")
		return nil
	}

	var buildDone bool
	if _, err := gexec.LookPath(c.Args().First()); err != nil {
		if err := buildAction(c); err != nil {
			fmt.Println("⡿ Run `shogun build -dir=''` to build package directory first before running `shogun [] [Command]`.")
			return nil
		}

		buildDone = true
	}

	if !buildDone {
		if err := buildAction(c); err != nil {
			// do nothing for now
			return err
		}
	}

	events := metrics.New()

	if c.Bool("verbose") {
		events = metrics.New(custom.StackDisplay(os.Stdout))
	}

	binaryPath := binPath()

	var command string

	if c.String("in") != "" {
		command = fmt.Sprintf("%s/%s -in %s %s", binaryPath, c.Args().First(), c.String("in"), strings.Join(c.Args().Tail(), " "))
	} else {
		command = fmt.Sprintf("%s/%s %s", binaryPath, c.Args().First(), strings.Join(c.Args().Tail(), " "))
	}

	var response, responseErr bytes.Buffer
	binCmd := exec.New(
		exec.Async(),
		exec.Command(command),
		exec.Output(&response),
		exec.Err(&responseErr),
		exec.Input(os.Stdin),
	)

	fmt.Printf("⡿ Executing %+q:\n", command)
	exitCode, err := binCmd.ExecWithExitCode(context.Background(), events)
	if err != nil {
		events.Emit(metrics.Error(err))
		return fmt.Errorf("Command Error: %+q\n %+s", err, responseErr.String())
	}

	if exitCode > 0 {
		return fmt.Errorf("Command Error: exitCode: %d\n%+q\n", exitCode, responseErr.String())
	}

	if responseErr.Len() != 0 {
		return fmt.Errorf("Command Error: exitCode: %d\n%+q\n", exitCode, responseErr.String())
	}

	fmt.Println(response.String())
	return nil
}

func listAction(c *cli.Context) error {
	events := metrics.New()

	if c.Bool("verbose") {
		events = metrics.New(custom.StackDisplay(os.Stdout))
	}

	currentDir, err := os.Getwd()
	if err != nil {
		events.Emit(metrics.Errorf("Failed to read current directory: %q", err))
		return err
	}

	ctx := build.Default
	ctx.BuildTags = append(ctx.BuildTags, "shogun")
	ctx.RequiredTags = append(ctx.RequiredTags, "shogun")

	// Build shogunate directory itself first.
	functions, err := samurai.ListFunctions(events, events, filepath.Join(currentDir, c.String("dir")), ctx)
	if err != nil {
		events.Emit(metrics.Errorf("Failed to generate function list : %+q", err))
		if err == samurai.ErrSkipDir {
			return nil
		}

		return err
	}

	result := gen.SourceTextWithName(
		"shogun-pkg-list",
		string(templates.Must("shogun-pkg-list.tml")),
		template.FuncMap{},
		functions,
	)

	_, err = result.WriteTo(os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func buildAction(c *cli.Context) error {
	events := metrics.New()

	if c.Bool("verbose") {
		events = metrics.New(custom.StackDisplay(os.Stdout))
	}

	skipBuild := c.Bool("skipbuild")
	forceBuild := c.Bool("force")
	tgDir := c.String("dir")
	cmdDir := c.String("cmdDir")

	binaryPath := binPath()
	currentDir, err := os.Getwd()
	if err != nil {
		events.Emit(metrics.Error(err).With("dir", currentDir).With("binary_path", binaryPath))
		return err
	}

	if igerr := checkAndAddIgnore(currentDir); igerr != nil {
		events.Emit(metrics.Errorf("Failed to add changes to .gitignore: %+q", igerr).With("dir", currentDir))
		return igerr
	}

	targetDir := filepath.Join(currentDir, tgDir)

	if cmdDir == "" {
		cmdDir = filepath.Join(".shogun", "cmd")
	}

	ctx := build.Default
	ctx.BuildTags = append(ctx.BuildTags, "shogun")
	ctx.RequiredTags = append(ctx.RequiredTags, "shogun")

	// Build hash list for directories.
	hashList, err := samurai.ListPackageHash(events, events, targetDir, ctx)
	if err != nil {
		events.Emit(metrics.Error(err).With("dir", currentDir).With("binary_path", binaryPath))
		if err == samurai.ErrSkipDir {
			return nil
		}

		return err
	}

	// Build directories for commands.
	directive, err := samurai.BuildPackage(events, events, targetDir, cmdDir, currentDir, binaryPath, skipBuild, c.Bool("remove"), c.Bool("singlepkg"), c.Bool("skipsub"), ctx)
	if err != nil {
		events.Emit(metrics.Error(err).With("dir", currentDir).With("binary_path", binaryPath))
		return err
	}

	var subUpdated bool

	for _, sub := range directive.Subs {
		if hashData, ok := hashList.Subs[sub.Path]; ok {
			hashFile := filepath.Join(goSrcPath, filepath.Dir(sub.PkgPath), ".hashfile")
			prevHash, err := readFile(hashFile)

			if err == nil && prevHash == hashData.Hash && !forceBuild {
				continue
			}
		}

		// if PkgPath is empty then possibly not one we want to handle, all must
		// have a place to store
		if sub.PkgPath == "" || len(sub.List) == 0 {
			continue
		}

		subUpdated = true
		if err := ast.SimpleWriteDirectives("./", true, sub.List...); err != nil {
			events.Emit(metrics.Error(err).With("dir", currentDir).With("binary_path", binaryPath))
			return err
		}
	}

	// Validate hash of main cmd.
	if !forceBuild {
		hashFile := filepath.Join(goSrcPath, filepath.Dir(directive.Main.PkgPath), ".hashfile")
		prevHash, err := readFile(hashFile)
		if err == nil && prevHash == hashList.Main.Hash && !subUpdated {
			return nil
		}
	}

	if err := ast.SimpleWriteDirectives("./", true, directive.Main.List...); err != nil {
		events.Emit(metrics.Error(err).With("dir", currentDir).With("binary_path", binaryPath))
		return err
	}

	return nil
}

func checkAndAddIgnore(currentDir string) error {
	ignoreFile := filepath.Join(currentDir, ".gitignore")
	if _, ierr := os.Stat(ignoreFile); ierr != nil {
		if igerr := addtoGitIgnore(ignoreFile); igerr != nil {
			return igerr
		}
	}

	ignoreFileData, err := ioutil.ReadFile(ignoreFile)
	if err != nil {
		return err
	}

	if !bytes.Contains(ignoreFileData, []byte(ignoreAddition)) {
		if igerr := addtoGitIgnore(ignoreFile); igerr != nil {
			return igerr
		}
	}

	return nil
}

func addtoGitIgnore(ignoreFile string) error {
	gitignore, err := os.OpenFile(ignoreFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer gitignore.Close()
	gitignore.Write([]byte(ignoreAddition))
	gitignore.Write([]byte("\n"))
	return nil
}

func hasDir(dir string) bool {
	if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
		return true
	}

	return false
}

func readFile(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	return string(bytes.TrimSpace(content)), err
}

func hasFile(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

func versionAction(c *cli.Context) {
	fmt.Println(color.BlueString(fmt.Sprintf("shogun %s %s/%s", Version, runtime.GOOS, runtime.GOARCH)))
}

func toPackageName(name string) string {
	return strings.ToLower(binNameReg.ReplaceAllString(name, ""))
}

func binPath() string {
	shogunBinPath := os.Getenv("SHOGUNBIN")
	gobin := os.Getenv("GOBIN")
	gopath := os.Getenv("GOPATH")

	if runtime.GOOS == "windows" {
		gobin = filepath.ToSlash(gobin)
		gopath = filepath.ToSlash(gopath)
		shogunBinPath = filepath.ToSlash(shogunBinPath)
	}

	if shogunBinPath == "" && gobin == "" {
		return fmt.Sprintf("%s/bin", gopath)
	}

	if shogunBinPath == "" && gobin != "" {
		return gobin
	}

	return shogunBinPath
}
