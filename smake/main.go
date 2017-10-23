package smake

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"shanhu.io/misc/goload"
	"shanhu.io/sml/goenv"
)

func pkgFromDir(src, dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	p, err := filepath.Rel(src, abs)
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(p), nil
}

func relPkgs(rootPkg string, pkgs []string) ([]string, error) {
	var ret []string
	prefix := rootPkg + "/"
	for _, pkg := range pkgs {
		if pkg == rootPkg {
			ret = append(ret, ".")
			continue
		}

		if strings.HasPrefix(pkg, prefix) {
			rel := strings.TrimPrefix(pkg, prefix)
			ret = append(ret, "./"+rel)
			continue
		}

		return nil, fmt.Errorf("%q is not in %q", pkg, rootPkg)
	}
	return ret, nil
}

type context struct {
	dir string
	env []string
}

func execPkgs(c *context, pkgs []string, tasks [][]string) error {
	for _, args := range tasks {
		line := strings.Join(args, " ")
		fmt.Println(line)

		if len(pkgs) > 0 {
			args = append(args, pkgs...)
		}
		p, err := exec.LookPath(args[0])
		if err != nil {
			return err
		}
		cmd := exec.Cmd{
			Path:   p,
			Args:   args,
			Dir:    c.dir,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Env:    c.env,
		}

		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func smakeDir(gopath, srcRoot, dir string) error {
	rootPkg, err := pkgFromDir(srcRoot, dir)
	if err != nil {
		return err
	}

	pkgs, err := goload.ListPkgs(rootPkg)
	if err != nil {
		return err
	}

	pkgs, err = relPkgs(rootPkg, pkgs)
	if err != nil {
		return err
	}

	c := &context{
		dir: dir,
		env: []string{fmt.Sprintf("GOPATH=%s", gopath)},
	}

	if err := execPkgs(c, pkgs, [][]string{
		{"go", "install", "-v"},
		{"gofmt", "-s", "-w", "-l"},
		{"go", "install", "-v"},
	}); err != nil {
		return err
	}

	if err := execPkgs(c, nil, [][]string{
		{"smlchk", fmt.Sprintf("-path=%s", rootPkg)},
	}); err != nil {
		return err
	}

	if err := execPkgs(c, pkgs, [][]string{
		{"golint"},
		{"go", "vet"},
		{"gotags", "-R", "-f", "tags"},
	}); err != nil {
		return err
	}

	return nil
}

func absGOPATH() (string, error) {
	gopath, err := goenv.GOPATH()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(gopath)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func smake() error {
	gopath, err := absGOPATH()
	if err != nil {
		return err
	}
	src := filepath.Join(gopath, "src")
	return smakeDir(gopath, src, ".")
}

// Main is the entry point for smake.
func Main() {
	if err := smake(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
