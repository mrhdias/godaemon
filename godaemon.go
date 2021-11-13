package godaemon

import (
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "os/exec"
  "path/filepath"
  "strconv"
  "syscall"
  "time"
)

type Daemon struct {
  Name          string
  PidFile       string
  ChDir         string
  Action        string
  RedirectStrFd bool
  OnStart       func()
  OnStop        func()
}

func getPidFromFile(pidFile string) int {
  pidStr, err := ioutil.ReadFile(pidFile)
  if err != nil {
    log.Fatalln("Unable to read file:", err)
  }
  pid, err := strconv.Atoi(string(pidStr))
  if err != nil {
    log.Fatalln("Wrong pid process number:", err)
  }

  return pid
}

func redirectStrFd() {
  file, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
  if err != nil {
    log.Fatalln("Failed to open /dev/null:", err)
  }
  syscall.Dup2(int(file.Fd()), int(os.Stdin.Fd()))
  syscall.Dup2(int(file.Fd()), int(os.Stdout.Fd()))
  syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd()))
  file.Close()
}

func (daemon *Daemon) run() {
  fmt.Printf("pid: %d %s\n", os.Getpid(), daemon.PidFile)

  if err := os.WriteFile(daemon.PidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
    fmt.Println("Unable to write to file:", err)
    os.Exit(1)
  }

  fmt.Println("The", daemon.Name, "was successfully started...")

  if daemon.RedirectStrFd {
    redirectStrFd()
  }
}

func (daemon *Daemon) status() {
  if _, err := os.Stat(daemon.PidFile); errors.Is(err, os.ErrNotExist) {
    fmt.Println("The", daemon.Name, "is stopped...")
    os.Exit(0)
  }

  pid := getPidFromFile(daemon.PidFile)

  process, err := os.FindProcess(int(pid))
  if err != nil {
    fmt.Println("Failed to find process:", err)
  } else {
    if err := process.Signal(syscall.Signal(0)); err != nil {
      fmt.Println("Process.Signal on pid", pid, "returned:", err)
    } else {
      fmt.Println("The", daemon.Name, "is running...")
    }
  }
}

func (daemon *Daemon) stop() {
  if _, err := os.Stat(daemon.PidFile); errors.Is(err, os.ErrNotExist) {
    fmt.Println("The", daemon.Name, "is already stopped...")
    os.Exit(0)
  }

  daemon.OnStop()

  pid := getPidFromFile(daemon.PidFile)

  if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
    fmt.Printf("Unable to kill the process %d: %v\n", pid, err)
    os.Exit(1)
  }

  if err := os.Remove(daemon.PidFile); err != nil {
    fmt.Printf("Unable to delete the file: %v\n", err)
    os.Exit(1)
  }

  fmt.Println("The", daemon.Name, "was successfully stopped.")
}

func (daemon *Daemon) start() {

  if _, err := os.Stat(daemon.PidFile); err == nil {
    fmt.Println("The", daemon.Name, "is already running...")
    os.Exit(0)
  }

  // Borrow from https://github.com/golang/go/issues/227
  // short delay to avoid race condition between os.StartProcess and os.Exit
  // can be omitted if the work done above amounts to a sufficient delay
  // time.Sleep(1 * time.Second)

  if err := syscall.FcntlFlock(os.Stdout.Fd(), syscall.F_SETLKW, &syscall.Flock_t{
    Type: syscall.F_WRLCK, Whence: 0, Start: 0, Len: 0}); err != nil {
    fmt.Println("Failed to lock stdout:", err)
    os.Exit(1)
  }

  if os.Getppid() != 1 {
    // I am the parent, spawn child to run as daemon
    binary, err := exec.LookPath(os.Args[0])
    if err != nil {
      fmt.Println("Failed to lookup binary:", err)
      os.Exit(1)
    }

    if _, err = os.StartProcess(binary, []string{os.Args[0], "run", "daemon"}, &os.ProcAttr{Dir: daemon.ChDir, Env: nil,
      Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: nil}); err != nil {
      fmt.Println("Failed to start process:", err)
      os.Exit(1)
    }
    os.Exit(0)
  } else {
    // I am the child, i.e. the daemon, start new session and detach from terminal
    if _, err := syscall.Setsid(); err != nil {
      fmt.Println("Failed to create new session:", err)
      os.Exit(1)
    }
  }
}

func (daemon *Daemon) Manager(worker func()) {

  daemon.Action = os.Args[1]
  switch os.Args[1] {
  case "run":
    if len(os.Args) == 3 && os.Args[2] == "daemon" {
      daemon.run()
    }
    daemon.OnStart()
    worker()
  case "start":
    daemon.start()
  case "stop":
    daemon.stop()
  case "restart":
    daemon.stop()
    time.Sleep(1 * time.Second)
    daemon.start()
  case "status":
    daemon.status()
  default:
    fmt.Println("Usage:", daemon.Name, "start | stop | restart | status | run")
  }

  if _, err := os.Stat(daemon.PidFile); err == nil {
    if err := os.Remove(daemon.PidFile); err != nil {
      log.Fatalf("Unable to delete the file: %v\n", err)
    }
  }
  os.Exit(0)
}

func New() Daemon {
  daemon := new(Daemon)
  daemon.Name = filepath.Base(os.Args[0])

  if len(os.Args) == 1 {
    fmt.Println("Usage:", daemon.Name, "start | stop | restart | status | run")
    os.Exit(0)
  }

  daemon.Action = os.Args[1]
  daemon.PidFile = fmt.Sprintf("%s.pid", daemon.Name)
  daemon.ChDir = ""
  daemon.Action = ""
  daemon.RedirectStrFd = true
  daemon.OnStart = func() {}
  daemon.OnStop = func() {}

  return *daemon
}
