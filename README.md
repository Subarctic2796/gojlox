# Gojlox
Crafting Interpreters tree walking interpreter but in go

# Running
```console
git clone --depth=1 https://github.com/Subarctic2796/gojlox.git
cd gojlox
go build .
./main
```

# Notes
If you get a compiler error about anything relating to `ast` then you may need to regenerate the ast files.
I hard-coded the path in `tools/genAst.go`, so you should change it.
Once you change it the `path` and run it, make sure to reformat the generated files.
## Generating AST files
```console
cd tools
go run .
```

# Currently on
Chap 10
challenge: 2
