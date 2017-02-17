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

var CheckErrors []string = []string{"ERROR:", "Error:", "error:", "FATAL:", "Fatal:", "fatal:"}

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

func trimCheck(r bytes.Buffer, trim bool, check bool, checkStart bool, checkList []string) (string, error) {
	var result string
	if trim {
		result = strings.TrimSpace(r.String())
	} else {
		result = r.String()
	}
	if check {
		for _, search := range checkList {
			if checkStart {
				if strings.HasPrefix(result, search) {
					return "", errors.New(result)
				}
			} else {
				if strings.Contains(result, search) {
					return "", errors.New(result)
				}
			}
		}
	}
	return result, nil
}

func doBaseGetWorkingDir() (string, error) {
	var o bytes.Buffer
	cmdGitPresent := exec.Command("git", "rev-parse", "--show-toplevel")
	if err := execCmd(*cmdGitPresent, "", &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, true, true, true, CheckErrors)
}

func doBaseGetRemoteUrl() (string, error) {
	var o bytes.Buffer
	cmdGitUrl := exec.Command("git", "remote", "get-url", "origin")
	if err := execCmd(*cmdGitUrl, "", &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, true, true, true, CheckErrors)
}

func doBaseCreateSandbox(sandboxDir string) error {
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		return errors.Wrap(err, "error: could not create sandbox directory")
	}
	return nil
}

func doBaseRemoveSandbox(sandboxDir string) error {
	if err := os.RemoveAll(sandboxDir); err != nil {
		return errors.Wrap(err, "error: could not remove sandbox directory")
	}
	return nil
}

func doGetHeadCommit(dir string) (string, error) {
	var o bytes.Buffer
	cmdGitHeadCommit := exec.Command("git", "rev-parse", "HEAD")
	if err := execCmd(*cmdGitHeadCommit, dir, &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, true, true, true, CheckErrors)
}

func doGetHeadBranch(dir string) (string, error) {
	var o bytes.Buffer
	checkGitHeadBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err := execCmd(*checkGitHeadBranch, dir, &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, true, true, true, CheckErrors)
}

func doGitClone(dir string, url string) error {
	var e bytes.Buffer
	cmdClone := exec.Command("git", "clone", "--verbose", url, dir)
	if err := execCmd(*cmdClone, "", nil, &e); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	_, err := trimCheck(e, false, true, true, CheckErrors)
	return err
}

func doGitFetch(dir string) error {
	var e bytes.Buffer
	cmdFetch := exec.Command("git", "fetch")
	if err := execCmd(*cmdFetch, dir, nil, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	_, err := trimCheck(e, false, true, true, CheckErrors)
	return err
}

func doGitCheckout(dir string, hash string) error {
	var o bytes.Buffer
	cmdCheckout := exec.Command("git", "checkout", hash)
	if err := execCmd(*cmdCheckout, dir, &o, nil); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	_, err := trimCheck(o, false, true, true, CheckErrors)
	return err
}

func doGetCommitList(dir string) (string, error) {
	var o bytes.Buffer
	cmdCommitList := exec.Command("git", "--no-pager", "log", "--pretty=format:%H::%s::%an")
	if err := execCmd(*cmdCommitList, dir, &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, false, true, true, CheckErrors)
}

func doGitBranchList(dir string) (string, error) {
	var o bytes.Buffer
	cmdBranchList := exec.Command("git", "for-each-ref", "--sort=committerdate", "refs/heads/", "--format=%(refname:short)")
	if err := execCmd(*cmdBranchList, dir, &o, nil); err != nil {
		return "", errors.Wrap(err, "error: an error occured while processing command")
	}
	return trimCheck(o, false, true, true, CheckErrors)
}

func doGitPull(dir string) error {
	var e bytes.Buffer
	cmdFetch := exec.Command("git", "pull")
	if err := execCmd(*cmdFetch, dir, nil, &e); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	_, err := trimCheck(e, false, true, true, CheckErrors)
	return err
}

func (b *browser) initialize() error {
	workingd, err := doBaseGetWorkingDir()
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.workingd = workingd

	gitUrl, err := doBaseGetRemoteUrl()
	if err != nil {
		return errors.Wrap(err, "error: an error occurred while processing command")
	}
	b.gitUrl = gitUrl

	rbhash, err := doGetHeadCommit("")
	if err != nil {
		return errors.Wrap(err, "error: an error occurred while processing command")
	}
	b.rbhash = rbhash

	rbbranch, err := doGetHeadBranch("")
	if err != nil {
		return errors.Wrap(err, "error: an error occurred while processing command")
	}
	b.rbbranch = rbbranch

	if err := b.execClone(false); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}

	return nil
}

func (b *browser) execClone(force bool) error {
	if _, err := os.Stat(filepath.Join(b.workingd, SandboxDirName, ".git")); os.IsNotExist(err) {
		if err := doBaseCreateSandbox(filepath.Join(b.workingd, SandboxDirName)); err != nil {
			return errors.Wrap(err, "error: could not create sandbox directory")
		}
		if err := doGitClone(filepath.Join(b.workingd, SandboxDirName), b.gitUrl); err != nil {
			return errors.Wrap(err, "error: an error occured while processing command")
		}
	} else if force {
		if err := doBaseRemoveSandbox(filepath.Join(b.workingd, SandboxDirName)); err != nil {
			return errors.Wrap(err, "error: could not remove sandbox directory")
		}
		if err := doBaseCreateSandbox(filepath.Join(b.workingd, SandboxDirName)); err != nil {
			return errors.Wrap(err, "error: could not create sandbox directory")
		}
		if err := doGitClone(filepath.Join(b.workingd, SandboxDirName), b.gitUrl); err != nil {
			return errors.Wrap(err, "error: an error occured while processing command")
		}
	} else if err != nil {
		return errors.Wrap(err, "error: could not create sandbox directory")
	}

	if err := doGitFetch(filepath.Join(b.workingd, SandboxDirName)); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}

	current, err := doGetHeadCommit(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.current = current

	rawCommitList, err := doGetCommitList(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.commitList = make([]commit, 0)
	clines := strings.Split(rawCommitList, "\n")
	for _, line := range clines {
		parts := strings.Split(line, "::")
		b.commitList = append(b.commitList, commit{
			hash:    parts[0],
			message: parts[1],
			author:  parts[2],
		})
	}

	rawBranchList, err := doGitBranchList(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.branchList = make([]string, 0)
	blines := strings.Split(rawBranchList, "\n")
	for _, line := range blines {
		if len(strings.TrimSpace(line)) > 0 {
			b.branchList = append(b.branchList, strings.TrimSpace(line))
		}
	}

	return nil
}

func (b *browser) execMove(hash string) error {
	if err := doGitCheckout(filepath.Join(b.workingd, SandboxDirName), hash); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	current, err := doGetHeadCommit(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.current = current
	return nil
}

func (b *browser) execLog() error {
	for _, commit := range b.commitList {
		if commit.hash == b.current {
			fmt.Printf("* %v - %v %v\n", commit.hash, commit.author, commit.message)
		} else {
			fmt.Printf("%v - %v %v\n", commit.hash, commit.author, commit.message)
		}
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
	if err := doGitPull(filepath.Join(b.workingd, SandboxDirName)); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	if err := doGitCheckout(filepath.Join(b.workingd, SandboxDirName), branch); err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}

	rawCommitList, err := doGetCommitList(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.commitList = make([]commit, 0)
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
	current, err := doGetHeadCommit(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.current = current
	for i, commit := range b.commitList {
		if strings.TrimSpace(commit.hash) == strings.TrimSpace(b.current) {
			if i > 0 {
				b.execMove(b.commitList[i-1].hash)
				break
			}
		}
	}
	return nil
}

func (b *browser) execPrev() error {
	current, err := doGetHeadCommit(filepath.Join(b.workingd, SandboxDirName))
	if err != nil {
		return errors.Wrap(err, "error: an error occured while processing command")
	}
	b.current = current
	for i, commit := range b.commitList {
		if strings.TrimSpace(commit.hash) == strings.TrimSpace(b.current) {
			if i < len(b.commitList)-1 {
				b.execMove(b.commitList[i+1].hash)
				break
			}
		}
	}
	return nil
}

func (b *browser) execClean() error {
	//Todo: Implement
	return nil
}

func (b *browser) execHelp() error {
	//Todo: Implement
	return nil
}

func (b *browser) execCommand(args []string) error {
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}
	switch args[0] {
	case "first":
		if len(b.commitList) <= 0 {
			return errors.New("error: no commits to browse")
		}
		return b.execMove(b.commitList[len(b.commitList)-1].hash)
	case "last":
		if len(b.commitList) <= 0 {
			return errors.New("error: no commits to browse")
		}
		return b.execMove(b.commitList[0].hash)
	case "list":
		return b.execLog()
	case "next":
		return b.execNext()
	case "prev":
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
	case "clone":
		return b.execClone(true)
	case "help":
		return b.execHelp()
	default:
		fmt.Println("error: invalid command")
		return nil
	}
	return nil
}

func (b *browser) runRepl() error {
	for {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")

		in, _ := consoleReader.ReadString('\n')
		in = strings.ToLower(in)

		if strings.HasPrefix(in, "quit") {
			b.execClean()
			break
		}
		if err := b.execCommand(strings.Split(in, " ")); err != nil {
			return errors.Wrap(err, "error: could not execute command")
		}
	}
	return nil
}

func main() {
	var debug bool = false
	var sandboxdir bool = false

	//Parse Configuration
	var interactive *bool = flag.Bool("i", false, "starts REPL to interactively issue commands")
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
			fmt.Printf(outputFmt, errors.New("error: invalid argument"))
			return
		}
	}

	if debug {
		fmt.Println("debug: args are ", *interactive, sandboxdir, args)
	}

	//Run Command
	b, err := NewBrowser(sandboxdir)
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
