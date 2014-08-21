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

# License
MIT