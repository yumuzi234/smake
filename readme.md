`smake` is the default make tool for `shanhu.io` development.

```
go install shanhu.io/smake
```

It automatically runs the following commands for all non-vendor
packages under the current directory:

```
go install -v
gofmt -s -w -l
go install -v
smlchk
golint
govet
go tags -R -f=tags
```

It replaces the old `makefiles` that we use.
