![Shogun](./media/shogun.png)

Shogun
---------

Shogun provides a function executor similar to `make` with a nice twist of small haikus and expanded capabilities
like io.Reader arguments, shogun does not just lets you run functions as tasks but lets you take such binaries
generated by shogun be to lunched anywhere like a beautiful lambda house.

*Inspired by [mage](https://github.com/magefile/mage) and Amazon Lambda functions*

## Install

```bash
go install -u github.com/influx6/shogun
```

## Shogun Go Files
Shogun requires all go files which wish to expose katana functions with the shogun build tag as below.

```
// +build shogun
```

Shogun by default will save binaries into the `GOBIN` or `GOPATH/bin` path extracted from the environment, however this can be changed by setting a `SHOGUNBIN` environment
varable. More so, Shogun names all binaries the name of the parent package unless one declares an explicit annotation `@binaryName` at the package level.

```
// +build shogun

// Package do does something.
//
//@binaryName(shogunate_bin)
package do
```

All binaries created by shogun are self complete and can equally be called directly without the `shogun` command, which makes it very usable for easy deployable self contained executables that can be used in place where behaviors need to be exposed as functions.

Beyond this, shogun copies content of these files out and tries to main as much as
possible the peculiar style of the source.

In shogun, you can tag a function as the default entry level function by using the  `@default` annotation.


## Functions

Shogun focuses on the execution of functions, that supports a limited but flexible
set of function format. More so, to match needs of most with `Context` objects, the
function formats support the usage of Context as first place arguments.

### Using Context

Only the following packages and interfaces are allowed for context usage.
If you need context then it must always be the first argument.

- context "context.Context"
- github.com/influx6/fuax/context "context.CancelContext"
- github.com/influx6/fuax/context "context.ValueBagContext"

When using `context.Context` as the context type which is part of the Go core packages, as far as the context is the only argument of any function if any json
sent as input then all json key-value pairs will be copied into the context.

Shogun will watch for `-time` flags and will use this to set timeouts for the 3
giving context else the context will not have expiration deadlines.

### Function Formats

Only the following function types format are allowed and others will be ignored, shogun provides flexibility enough without sacrificing simplicity so you can write and leverage behaviour in a suitable manner.

-	`func()`
- `func() error`

- `func(Context)`
- `func(Context) error`

- `func(map[string]interface{}) error`
- `func(Context, map[string]interface{}) error`

- `func(Interface) error`
- `func(Context, Interface) error`
- `func(Interface, io.WriteCloser) error`
- `func(Context, Interface, io.WriteCloser) error`

- `func(Struct) error`
- `func(Context, Struct) error`
- `func(Struct, io.WriteCloser) error`
- `func(Context, Struct, io.WriteCloser) error`

- `func(package.Type) error`
- `func(Context, package.Type) error`
- `func(Struct, io.WriteCloser) error`
- `func(Context, package.Type, io.WriteCloser) error`

- `func(io.Reader) error`
- `func(io.Reader, io.WriteCloser) error`
- `func(Context, io.Reader, io.WriteCloser) error`

*Where `Context` => represents the context package used of the 3 allowed.*
*Where `Struct`   => represents any struct declared in package*
*where `Interface` => represents any interface declared in package*
*where `package.Type` => represents any type(Struct, Interface, OtherType) imported from other package, except functions*

Any other thing beyond this type formats won't be allowed and will be ignored in
function list and execution.

## Commands

But using the `shogun` command, we can do the following:

- Build a package shogun files

```bash
shogun build
```

Shogun will hash shogun files and ensure only when changes occur will a new build be made and binary will be stored in binary location as dictated by environment variable `SHOGUNBIN` or default `GOBIN`/`GOPATH/bin` .

- Run function within shogun files expecting no input

```bash
shogun {{FUNCTIONNAME}}
```

- Run function within shogun files with standard input

```bash
echo "We lost the war" | shogun {{FUNCTIONNAME}}
```

- Run function within shogun files with json input

```bash
shogun {{FUNCTIONNAME}} -json {"name":"bat"}
```

- List all functions with

```bash
shogun list
```

- List all functions with short commentary

```bash
shogun help {{FunctionName}}
```

- List all functions with full commentary

```bash
shogun help -f {{FunctionName}}
```

- List all functions with full commentary and source snippet

```bash
shogun help -f=true -s=true {{FunctionName}}
```
