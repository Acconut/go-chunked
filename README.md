<h1 align="center">Chunked</h1>

An experimental data storage built using [Go](http://golang.org). Chunked is similar to traditional key-value store with the exception that it doesn't allow you to specify the key you want to store a value under.



| Build Status | Test Coverage | Documentation |
|:-:|:-:|:-:|
| [![Build Status](https://travis-ci.org/Acconut/go-chunked.svg?branch=master)](https://travis-ci.org/Acconut/go-chunked) | [![Coverage Status](https://coveralls.io/repos/Acconut/go-chunked/badge.png?branch=master)](https://coveralls.io/r/Acconut/go-chunked?branch=master) | [Godoc](http://godoc.org/github.com/Acconut/go-chunked) |

# Usage

```bash
go get github.com/Acconut/go-chunked
```

```go
package main

import (
    "github.com/Acconut/go-chunked"
    "fmt"
    "log"
)

func main() {

    // Create new database
    db, err := chunked.Create("./db", chunked.DefaultConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // Insert value
    key, err := db.Append([]byte("hello world"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Read value
    value, err := db.Read(key)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(value)) // -> hello world
    
    // Close database
    if err = db.Close(); err != nil {
        log.Fatal(err)
    }
}
```

# Interals

Chunked splites your values into pieces so that each one of them fits into an array with the length defined in `Blocksize`.
If you have a blocksize of 10 and a value with the length of 22 then you're going to have three blocks: Two with length of 10 and one of 2.
The first block contains a reference to the second and the second block to the thrid one. So you can start reading all of them just from knowing the position of the first one.
The maximum number of blocks in a chunk-file is defined using `Chunksize`. Using a chunksize of 20, the first 20 blocks are stored in the first chunk-file.
21st to 40th block can be found in the second file and so on.

# License
MIT