# Gojlox
[Crafting Interpreters](https://craftinginterpreters.com) tree walking interpreter but in go, with a few add ons.

## Additions
- `+=, -=, etc..` operators.
- `printf` and other native functions.
- lambda's or annoymous functions. `var a = fun(x) { print x; };`
- repl can evaluate expressions, not only statements. (WIP)
    - `2 + 3` in the repl prints 5.
- static class functions using the `static` keyword before a class method.

# Running
```console
git clone --depth=1 https://github.com/Subarctic2796/gojlox.git
cd gojlox
go build .
./gojlox
```

# Notes
If you get a compiler error about anything relating to `ast` then you may need to regenerate the ast files.
The `genAst.go` script uses the `os.Getwd` function from the standard library to find where to put the files
it needs to generate. So, it has all the issues with using that function, it also doesn't handle symlinks. The
script also assumes that a directory called `ast` is in the direct parent directory from where the script is run
so make sure that that directory exists there.
Once you run it, make sure to reformat the generated files.
## Generating AST files
```console
cd tools
go run .
```

# Current plans
- [ ] add support for expressions in the repl
- [ ] improve performance
  - [ ] move some of the resolver checks to the parser
  - [ ] use arrays instead of hashmaps for `Env` struct
  - [ ] move away from visitor pattern, and just use straight type checks
- [ ] add arrays and hashmaps
- [ ] make `;` optional
- [ ] add ability to import other files
- [ ] add type hints (want to make it statically typed if possible)
- [ ] add errors so that scripts can recover
- [ ] add proper variadics
- [ ] add ability to define multiple variables on the same line
