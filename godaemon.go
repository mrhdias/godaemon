package godaemon

import (
  "errors"
  "flag"
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

type fn func()

type Daemon struct {
  Name          string
  PidFile       string
  LogFile       string
  RedirectStrFd bool
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
    log.Fatalln("Unable to write to file:", err)
  }

  log.Println("The", daemon.Name, "was successfully started...")

  logfile, err := os.OpenFile(daemon.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    log.Fatalf("Error opening file: %v", err)
  }
  defer logfile.Close()

  log.SetOutput(logfile)
  log.Println("Set log output to", daemon.LogFile, "file")
}

func (daemon *Daemon) stop() {
  if _, err := os.Stat(daemon.PidFile); errors.Is(err, os.ErrNotExist) {
    fmt.Println("The", daemon.Name, "is already stopped...")
    os.Exit(0)
  }

  pidStr, err := ioutil.ReadFile(daemon.PidFile)
  if err != nil {
    log.Fatalf("unable to read file: %v", err)
  }

  pid, err := strconv.Atoi(string(pidStr))
  if err != nil {
    log.Fatalf("Wrong pid process number: %v", err)
  }

  if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
    log.Fatalf("Unable to kill the process %d: %v", pid, err)
  }

  if err := os.Remove(daemon.PidFile); err != nil {
    log.Fatalf("Unable to delete the file: %v", err)
  }

  fmt.Println("The", daemon.Name, "was successfully stopped.")
}

func (daemon *Daemon) start() {

  if _, err := os.Stat(daemon.PidFile); err == nil {
    fmt.Println("The", daemon.Name, "is already running...")
    os.Exit(0)
  }

  // From https://github.com/golang/go/issues/227
  // short delay to avoid race condition between os.StartProcess and os.Exit
  // can be omitted if the work done above amounts to a sufficient delay
  // time.Sleep(1 * time.Second)

  if err := syscall.FcntlFlock(os.Stdout.Fd(), syscall.F_SETLKW, &syscall.Flock_t{
    Type: syscall.F_WRLCK, Whence: 0, Start: 0, Len: 0}); err != nil {
    log.Fatalln("Failed to lock stdout:", err)
  }

  if os.Getppid() != 1 {
    // I am the parent, spawn child to run as daemon
    binary, err := exec.LookPath(os.Args[0])
    if err != nil {
      log.Fatalln("Failed to lookup binary:", err)
    }
    os.Args[1] = "run"
    if _, err = os.StartProcess(binary, os.Args, &os.ProcAttr{Dir: "", Env: nil,
      Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Sys: nil}); err != nil {
      log.Fatalln("Failed to start process:", err)
    }
    os.Exit(0)
  } else {
    // I am the child, i.e. the daemon, start new session and detach from terminal
    if _, err := syscall.Setsid(); err != nil {
      log.Fatalln("Failed to create new session:", err)
    }
    if daemon.RedirectStrFd {
      redirectStrFd()
    }
  }
}

func (daemon *Daemon) Daemonize(worker fn) {

  startCmd := flag.NewFlagSet("start", flag.ExitOnError)
  stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
  restartCmd := flag.NewFlagSet("restart", flag.ExitOnError)
  runCmd := flag.NewFlagSet("run", flag.ExitOnError)
  // statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

  if len(os.Args) == 1 {
    fmt.Println("Usage: Expected 'start', 'stop', 'restart' or 'run' commands")
    os.Exit(0)
  }

  switch os.Args[1] {
  case "run":
    runCmd.Parse(os.Args[2:])
    daemon.run()
    worker()
  case "start":
    startCmd.Parse(os.Args[2:])
    daemon.start()
  case "stop":
    stopCmd.Parse(os.Args[2:])
    daemon.stop()
    os.Exit(0)
  case "restart":
    restartCmd.Parse(os.Args[2:])
    daemon.stop()
    time.Sleep(1 * time.Second)
    daemon.start()
  default:
    fmt.Println("Usage: Expected 'start', 'stop', 'restart' or 'run' commands")
    os.Exit(0)
  }
}

func New() Daemon {
  daemon := new(Daemon)
  daemon.Name = filepath.Base(os.Args[0])
  daemon.PidFile = fmt.Sprintf("%s.pid", daemon.Name)
  daemon.LogFile = fmt.Sprintf("%s.log", daemon.Name)
  daemon.RedirectStrFd = true
  return *daemon
}
