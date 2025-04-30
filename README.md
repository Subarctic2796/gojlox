# Gojlox
[Crafting Interpreters](https://craftinginterpreters.com) tree walking interpreter but in go, with a few add ons.

## Additions
- `+=, -=, etc..` operators.
- `printf` and other native functions.
- lambda's or annoymous functions. `var a = fun(x) { print x; };`
- repl can evaluate expressions, not only statements. (WIP)
    - `2 + 3` in the repl prints 5.
- static class functions using the `static` keyword before a class method.
- arrays and hashmaps

# Running
```console
git clone --depth=1 https://github.com/Subarctic2796/gojlox.git
cd gojlox
make build
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

# Currently working on
- [ ] make instances hashable
  - [ ] maybe add id or uuid to instance that can then be used for the key

# Current plans
- [ ] add support for expressions in the repl
- [ ] improve performance
  - [ ] move some of the resolver checks to the parser
	- [ ] `ErrAlreadyInScope` check
	- [ ] `ErrLocalInitializesSelf` check
	- [ ] `ErrLocalNotRead` check
  - [ ] use arrays instead of hashmaps for `Env` struct
  - [x] move away from visitor pattern, and just use straight type checks
  - [ ] precompute some binary nodes
  - [ ] make real negative numbers
  - [ ] store scope info in ast nodes
- [ ] add arrays and hashmaps
  - [x] add trailing comma support
  - [x] add hashmaps
    - [ ] make instances hashable
  - [x] add arrays
  - [x] add fancy indexing
    - [x] add slicing `print arr[2:5];`
    - [x] add negative indexing `print a[-2];`
  - [ ] add `for in` loops.
    - [ ] add indexed looping `for (var k, v in hashmap) printf(k, v);`
- [ ] make `;` optional
- [ ] add ability to import other files
- [ ] add type hints (want to make it statically typed if possible)
- [ ] add errors so that scripts can recover
- [ ] add proper variadics
- [ ] add ability to define multiple variables on the same line `var a, b, c = 1, "hi", true;`
- [ ] add test suite
- [ ] add `--tokens` and `--ast` flags to output the tokens and ast respectively to stdout (maybe compile flag also)
  - [ ] add pretty printer for ast
- [ ] consolidate `Return`, `Break`, `Continue` Statements into `ControlStmt`
  - [ ] add `continue` keyword
- [ ] remove `genAst.go` script
- [ ] add native classes (act as modules maybe?)
- [ ] back port clox variable handling ?
- [ ] add compile step (?)
- [ ] create Makefile
- [ ] add better error messages
- [ ] add `else if` branches and `switch` cases
- [ ] add debugging support, dumping env, etc
- [ ] add a native dummy function
- [ ] setup github releases
- [ ] add string interpolation
