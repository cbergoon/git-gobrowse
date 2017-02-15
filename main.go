package main

import (
  "flag"
  "fmt"

  "github.com/pkg/errors"
)

const SandboxDirName string = ".sandbox"

type commit struct {
  hash    string
  message string
  author  string
}

type browser struct {
  rbbranch   string
  rbhash     string
  sandboxed  bool
  gitUrl     string
  current    string
  nextHash   string
  branch     string
  commitList []commit
  treeList   []string
}

func NewBrowser(sandboxed bool) (*browser, error) {
  //Todo: Set Git URL
  //Todo: Set Commit List
  //Todo: Set Sandboxed
  return nil, nil
}

func (b *browser) initialize() error {
  //Ensure Valid Directory
  //Todo: Ensure '.git' Director Present; Error if DNE
  return nil
}

func (b *browser) execClone() error {
  return nil
}

func (b *browser) execMove(hash string) error {
  return nil
}

func (b *browser) execLog() error {
  return nil
}

func (b *browser) execBranch() error {
  return nil
}

func (b *browser) execCommand(args []string) error {
  switch args[0] {
  case "first":
  //Todo: Move to First Commit
  case "last":
  //Todo: Move to Last Commit
  case "list":
  //Todo: List Commits
  case "next":
    if len(args) != 2 {
      //Todo: Error Invalid Arguments
    }
  //Todo: Move to Next Commit
  case "move":
    if len(args) != 2 {
      //Todo: Error Invalid Arguments
    }
  //Todo: Move to Commit/Hash/Tag
  case "branch":
    if len(args) != 2 {
      //Todo: Error Invalid Arguments
    }
  //Todo: Checkout value
  case "clean":
  //Todo: Clean Up Mess
  default:
  //Todo: Error Invalid Command
  }
}

func (b *browser) runRepl() error {
  //While command != quit
  return nil
}

func main() {
  //Parse Configuration
  var interactive *bool = flag.Bool("i", false, "starts REPL to interactively issue commands")
  var sandboxdir *bool = flag.Bool("s", false, "sandboxed directory to browse repo")
  var debug *bool = flag.Bool("d", false, "debug mode for command")

  flag.Parse()

  var outputFmt string = "%v\n"
  if *debug {
    outputFmt = "%+v\n"
  }

  args := flag.Args()
  if *interactive {
    if len(args) > 0 {
      fmt.Printf(outputFmt, errors.New("error: too many argument"))
    }
  } else {
    if len(args) < 1 || len(args) > 2 {
      fmt.Printf(outputFmt, errors.New("error: too many argument"))
    }
  }

  if *debug {
    fmt.Println("debug: args are ", *interactive, *sandboxdir, args)
  }

  //Run Command
  b, err := NewBrowser(*sandboxdir)
  if err != nil {
    //Todo: Error
  }

  if err := b.initialize(); err != nil {
    fmt.Printf(outputFmt, err)
    return
  }

  if *interactive {
    if err := b.runRepl(); err != nil {
      //Todo: Error
    }
  } else {
    if err := b.execCommand(args); err != nil {
      //Todo: Error
    }
  }
}
