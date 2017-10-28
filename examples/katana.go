// +build shogun

// Package katanas provides exported functions as tasks runnable from commandline.
//
// @binaryName(name => katana-shell)
//
package katanas

import (
	"fmt"
	"io"

	"github.com/influx6/faux/context"
	ty "github.com/influx6/shogun/examples/types"
)

type wondra struct {
	Name string
}

func Draw() {}

// Slash is the default tasks due to below annotation.
// @default
func Slash() error {
	fmt.Println("Welcome to Katana slash!")
	return nil
}

// Buba is bub.
func Buba(ctx context.CancelContext, name string) {
	fmt.Printf("Welcome Buba %q.\n", name)
}

// Bubas is slice of bub.
func Bubas(ctx context.CancelContext, names []string) {
	fmt.Printf("Welcome Buba city %q.\n", names)
}

// BubasMetro is slice of bub.
func BubasMetro(ctx context.CancelContext, names []string, w io.WriteCloser) {
	fmt.Printf("Welcome Buba metro city %q.\n", names)
}

func Bob(ctx context.CancelContext) error {
	return nil
}

// Jija does something.
// @flag(name => time, env => JIJA_TIME, type => Duration, desc => specifies time for jija)
func Jija(ctx context.CancelContext, mp ty.Woofer) error {
	return nil
}

func JijaPointer(ctx context.CancelContext, mp *ty.Woofer) error {
	return nil
}

func Juga(ctx context.CancelContext, r io.Reader) error {
	return nil
}

func Doba(ctx context.CancelContext, mp ty.IBlob) error {
	return nil
}

func Uiga(ctx context.CancelContext, n string, w io.WriteCloser) error {
	return nil
}

func Biga(ctx context.CancelContext, r io.Reader, w io.WriteCloser) error {
	return nil
}

func Nack(ctx context.CancelContext, mp map[string]interface{}) error {
	return nil
}

func Rulla(ctx context.CancelContext, mp wondra, w io.WriteCloser) error {
	return nil
}

func Hulla(ctx context.CancelContext, mp *wondra, w io.WriteCloser) error {
	return nil
}

func Guga(ctx context.CancelContext, mp ty.IBlob, w io.WriteCloser) error {
	return nil
}
