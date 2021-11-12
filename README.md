# godaemon
A simple package to daemonize Go applications.

## Installation
```
go get github.com/mrhdias/godaemon
```
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
  // Optional
  // daemon.Name = "test"
  // daemon.PidFile = "test.pid"
  // daemon.RedirectStrFd = true
  
  // Optional
  daemon.OnStart = func() {
    fmt.Println("On Start...")
  }
  
  // Optional
  daemon.OnStop = func() {
    fmt.Println("On Stop...")
  }
  
  daemon.Daemonize(func() {
    // Deemon business logic starts here
    fmt.Println("Daemonize my staff...")
    for {
      fmt.Println("Hello Again")
      time.Sleep(5 * time.Second)
    }
  })
}
```
## Test
```
go run test.go
Usage: test start | stop | restart | status | run
```
