package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	workingd   string
	sandboxed  bool
	gitUrl     string
	current    string
	branch     string
	commitList []commit
	branchList []string
}

func NewBrowser(sandboxed bool) (*browser, error) {
	return &browser{
		sandboxed: sandboxed,
	}, nil
}

func execCmd(cmd exec.Cmd, dir string, sout io.Writer, serr io.Writer) error {
	cmd.Stdout = sout
	cmd.Stderr = serr
	if dir != "" {
		cmd.Dir = dir
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "error: failed to execute command")
	}
	return nil
}

func (b *browser) initialize() error {
	var o bytes.Buffer

	cmdGitPresent := exec.Command("git", "rev-parse", "--show-toplevel")
	if err := execCmd(*cmdGitPresent, "", &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.workingd = strings.TrimSpace(o.String())
	if strings.Contains(b.workingd, "fatal") {
		return errors.New(b.workingd)
	}
	o.Reset()

	cmdGitUrl := exec.Command("git", "remote", "get-url", "origin")
	if err := execCmd(*cmdGitUrl, "", &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.gitUrl = strings.TrimSpace(o.String())
	if strings.Contains(b.gitUrl, "fatal") {
		return errors.New(b.gitUrl)
	}
	o.Reset()

	cmdGitHeadCommit := exec.Command("git", "rev-parse", "HEAD")
	if err := execCmd(*cmdGitHeadCommit, "", &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.rbhash = strings.TrimSpace(o.String())
	if strings.Contains(b.rbhash, "fatal") {
		return errors.New(b.rbhash)
	}
	o.Reset()

	checkGitHeadBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err := execCmd(*checkGitHeadBranch, "", &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.rbbranch = strings.TrimSpace(o.String())
	if strings.Contains(b.rbbranch, "fatal") {
		return errors.New(b.rbbranch)
	}
	o.Reset()

	if err := b.execClone(); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}

	return nil
}

func (b *browser) execClone() error {
	var o bytes.Buffer
	var e bytes.Buffer

	if _, err := os.Stat(filepath.Join(b.workingd, SandboxDirName, ".git")); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(b.workingd, SandboxDirName), 0755); err != nil {
			return errors.Wrap(err, "error: could not create sandbox directory")
		}

		cmdClone := exec.Command("git", "clone", "--verbose", b.gitUrl, filepath.Join(b.workingd, SandboxDirName))
		if err := execCmd(*cmdClone, "", nil, &e); err != nil {
			return errors.Wrap(err, "error: an error occured while processing command")
		}
		rawClone := e.String()
		if strings.Contains(rawClone, "fatal") || strings.Contains(rawClone, "error") {
			return errors.New(rawClone)
		}
		e.Reset()
	} else if err != nil {
		return errors.Wrap(err, "error: could not create sandbox directory")
	}

	cmdFetch := exec.Command("git", "fetch")
	if err := execCmd(*cmdFetch, filepath.Join(b.workingd, SandboxDirName), nil, &e); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	rawFetch := e.String()
	if strings.Contains(rawFetch, "fatal") || strings.Contains(rawFetch, "error") {
		return errors.New(rawFetch)
	}
	e.Reset()

	cmdCommitList := exec.Command("git", "--no-pager", "log", "--pretty=format:%H::%s::%an")
	if err := execCmd(*cmdCommitList, filepath.Join(b.workingd, SandboxDirName), &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	rawCommitList := o.String()
	if strings.Contains(rawCommitList, "fatal") {
		return errors.New(rawCommitList)
	}
	o.Reset()
	clines := strings.Split(rawCommitList, "\n")
	for _, line := range clines {
		parts := strings.Split(line, "::")
		b.commitList = append(b.commitList, commit{
			hash:    parts[0],
			message: parts[1],
			author:  parts[2],
		})
	}

	cmdBranchList := exec.Command("git", "for-each-ref", "--sort=committerdate", "refs/heads/", "--format=%(refname:short)")
	if err := execCmd(*cmdBranchList, filepath.Join(b.workingd, SandboxDirName), &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	rawBranchList := o.String()
	if strings.Contains(rawBranchList, "fatal") {
		return errors.New(rawBranchList)
	}
	o.Reset()
	blines := strings.Split(rawBranchList, "\n")
	for _, line := range blines {
		b.branchList = append(b.branchList, strings.TrimSpace(line))
	}

	return nil
}

func (b *browser) execMove(hash string) error {
	var o bytes.Buffer
	cmdCheckout := exec.Command("git", "checkout", hash)
	if err := execCmd(*cmdCheckout, filepath.Join(b.workingd, SandboxDirName), &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	checkoutOutput := o.String()
	if strings.Contains(checkoutOutput, "error") {
		return errors.New(checkoutOutput)
	}
	return nil
}

func (b *browser) execLog() error {
	for _, commit := range b.commitList {
		fmt.Printf("%v - %v %v\n", commit.hash, commit.author, commit.message)
	}
	return nil
}

func (b *browser) execBranchList() error {
	for _, branch := range b.branchList {
		fmt.Printf("%v\n", branch)
	}
	return nil
}

func (b *browser) execBranch(branch string) error {
	var o bytes.Buffer
	var e bytes.Buffer

	cmdFetch := exec.Command("git", "pull")
	if err := execCmd(*cmdFetch, filepath.Join(b.workingd, SandboxDirName), nil, &e); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	rawFetch := e.String()
	if strings.Contains(rawFetch, "fatal") || strings.Contains(rawFetch, "error") {
		return errors.New(rawFetch)
	}
	e.Reset()

	cmdCheckout := exec.Command("git", "checkout", branch)
	if err := execCmd(*cmdCheckout, filepath.Join(b.workingd, SandboxDirName), &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	checkoutOutput := o.String()
	if strings.Contains(checkoutOutput, "error") {
		return errors.New(checkoutOutput)
	}
	o.Reset()

	cmdCommitList := exec.Command("git", "--no-pager", "log", "--pretty=format:%H::%s::%an")
	if err := execCmd(*cmdCommitList, filepath.Join(b.workingd, SandboxDirName), &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	rawCommitList := o.String()
	if strings.Contains(rawCommitList, "fatal") {
		return errors.New(rawCommitList)
	}
	o.Reset()
	clines := strings.Split(rawCommitList, "\n")
	for _, line := range clines {
		parts := strings.Split(line, "::")
		b.commitList = append(b.commitList, commit{
			hash:    parts[0],
			message: parts[1],
			author:  parts[2],
		})
	}

	return nil
}

func (b *browser) execNext() error {
	for i, commit := range b.commitList {
		if strings.Compare(commit.hash, b.current) == 0 {
			if i < len(b.commitList)-1 {
				b.execMove(b.commitList[i+1].hash)
			}
		}
	}
	return nil
}

func (b *browser) execPrev() error {
	for i, commit := range b.commitList {
		if strings.Compare(commit.hash, b.current) == 0 {
			if i > 0 {
				b.execMove(b.commitList[i-1].hash)
			}
		}
	}
	return nil
}

func (b *browser) execClean() error {
	return nil
}

func (b *browser) execCommand(args []string) error {
	switch args[0] {
	case "first":
		if len(b.commitList) > 0 {
			return errors.New("error: no commits to browse")
		}
		return b.execMove(b.commitList[0].hash)
	case "last":
		if len(b.commitList) > 0 {
			return errors.New("error: no commits to browse")
		}
		return b.execMove(b.commitList[len(b.commitList)-1].hash)
	case "list":
		return b.execLog()
	case "next":
		if len(args) != 2 {
			return errors.New("error: invalid arguments")
		}
		return b.execNext()
	case "prev":
		if len(args) != 2 {
			return errors.New("error: invalid arguments")
		}
		return b.execPrev()
	case "move":
		if len(args) != 2 {
			return errors.New("error: invalid arguments")
		}
		return b.execMove(args[1])
	case "branch":
		if len(args) != 2 {
			return errors.New("error: invalid arguments")
		}
		b.execBranch(args[1])
	case "branch-list":
		return b.execBranchList()
	case "clean":
		return b.execClean()
	default:
		return errors.New("error: invalid command")
	}
	return nil
}

func (b *browser) runRepl() error {
	for {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")

		in, _ := consoleReader.ReadString('\n')
		in = strings.ToLower(in)

		if err := b.execCommand(strings.Split(in, " ")); err != nil {
			return errors.Wrap(err, "error: could not execute command")
		}

		if strings.HasPrefix(in, "quit") {
			b.execClean()
			break
		}
	}
	return nil
}

func main() {
	var debug bool = true

	//Parse Configuration
	var interactive *bool = flag.Bool("i", false, "starts REPL to interactively issue commands")
	var sandboxdir *bool = flag.Bool("s", true, "sandboxed directory to browse repo")

	flag.Parse()

	var outputFmt string = "%v\n"
	if debug {
		outputFmt = "%+v\n"
	}

	args := flag.Args()
	if *interactive {
		if len(args) > 0 {
			fmt.Printf(outputFmt, errors.New("error: too many argument"))
			return
		}
	} else {
		if len(args) < 1 || len(args) > 2 {
			fmt.Printf(outputFmt, errors.New("error: too many argument"))
			return
		}
	}

	if debug {
		fmt.Println("debug: args are ", *interactive, *sandboxdir, args)
	}

	//Run Command
	b, err := NewBrowser(*sandboxdir)
	if err != nil {
		fmt.Printf(outputFmt, errors.Wrap(err, "error: failed to create browser"))
		return
	}

	if err := b.initialize(); err != nil {
		fmt.Printf(outputFmt, errors.Wrap(err, "error: failed to initialize browser"))
		return
	}

	if debug {
		fmt.Println(b)
	}

	if *interactive {
		if err := b.runRepl(); err != nil {
			fmt.Println("error occurred: be sure to clean up with 'git gobrowse clean'")
			fmt.Printf(outputFmt, errors.Wrap(err, "error: repl error caused process to fail"))
			return
		}
	} else {
		if err := b.execCommand(args); err != nil {
			fmt.Println("error occurred: be sure to clean up with 'git gobrowse clean'")
			fmt.Printf(outputFmt, errors.Wrap(err, "error: exec error caused process to fail"))
			return
		}
	}
}
