# godaemon
A simple package to daemonize Go applications.

## Installation

	go get github.com/mrhdias/godaemon

## Example
```go
package main

import (
	"fmt"
	"time"
 
	"github.com/mrhdias/godaemon"
)

func main() {
  daemon := godaemon.New()
  daemon.Daemonize(func() {
    fmt.Println("Start...")
    for {
      fmt.Println("Hello Again")
      time.Sleep(5 * time.Second)
    }
  })
}
```
