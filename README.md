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
  // daemon.ChDir = "test"
  // daemon.RedirectStrFd = true
  
  // Optional
  daemon.OnStart = func() {
    fmt.Println("On Start...")
  }
  
  // Optional
  daemon.OnStop = func() {
    fmt.Println("On Stop...")
  }
  
  daemon.Manager(func() {
    // Deemon business logic starts here
    fmt.Println("Daemonize my staff...")
    for {
      fmt.Println("Hello Again")
      time.Sleep(5 * time.Second)
    }
  })
}
```
## Example How To Use with the Flag Package
```go
package main

import (
  "flag"
  "fmt"
  "os"
  "time"

  "github.com/mrhdias/godaemon"
)

func main() {

  var action string
  if len(os.Args) == 3 && os.Args[1] == "run" && os.Args[2] == "daemon" {
    // Daemonize
    action = "run"
  } else {
    flag.StringVar(&action, "a", "", "action: start | stop | restart | status")
    flag.Parse()

    if _, ok := map[string]bool{
      "start":   true,
      "stop":    true,
      "restart": true,
      "status":  true}[action]; !ok {

      fmt.Printf("The action \"%s\" not exist:\n", action)
      flag.PrintDefaults()
      os.Exit(0)
    }
	}

  daemon := godaemon.New()
  // Optional
  // daemon.Name = "test"
  // daemon.PidFile = "test.pid"
  // daemon.ChDir = "test"
  daemon.RedirectStrFd = false
  daemon.Action = action

  // Optional
  daemon.OnStart = func() {
    fmt.Println("On Start...")
  }

  // Optional
  daemon.OnStop = func() {
    fmt.Println("On Stop...")
  }

  daemon.Manager(func() {
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
