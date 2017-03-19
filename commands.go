package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/urfave/cli"
	"github.com/yukimemi/core"
	"github.com/yukimemi/file"
)

// GitStatus is git status struct.
type GitStatus struct {
	Path   string
	Diff   bool
	Ahead  bool
	Behind bool
	Status string
}

var (
	// GlobalFlags is global flag for app.
	GlobalFlags = []cli.Flag{
		cli.BoolFlag{
			EnvVar: "GSR_SHOW_STATUS",
			Name:   "status",
			Usage:  "Show status",
		},
		cli.BoolFlag{
			EnvVar: "GSR_SHOW_AHEAD",
			Name:   "ahead",
			Usage:  "Show ahead repo",
		},
		cli.BoolFlag{
			EnvVar: "GSR_SHOW_BEHIND",
			Name:   "behind",
			Usage:  "Show behind repo",
		},
		cli.BoolFlag{
			EnvVar: "GSR_SHOW_ALL",
			Name:   "all",
			Usage:  "Show all directry",
		},
	}

	mu       = new(sync.Mutex)
	aheadRe  = regexp.MustCompile(`\[ahead`)
	behindRe = regexp.MustCompile(`\[behind`)
)

// GlobalAction is Global gsr command.
var GlobalAction = func(c *cli.Context) error {

	var (
		err  error
		root string
		opt  file.Option

		wg  = new(sync.WaitGroup)
		sem = make(chan struct{}, runtime.NumCPU())
	)

	if c.NArg() > 0 {
		root = c.Args().First()
		opt = file.Option{
			Matches: []string{`\.git$`},
			Recurse: true,
		}
	} else {
		cmd := core.Cmd{Cmd: exec.Command("ghq", "root")}
		err = cmd.CmdRun()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		root = strings.TrimRight(cmd.Stdout.String(), "\n")
		opt = file.Option{
			Matches: []string{`\.git$`},
			Depth:   4,
		}
	}

	if !file.IsExistDir(root) {
		msg := fmt.Sprintf("[%v] is not exist", root)
		return fmt.Errorf(msg)
	}

	dirs, err := file.GetDirs(root, opt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	for info := range dirs {
		if info.Err != nil {
			fmt.Fprintln(os.Stderr, info.Err)
			continue
		}
		gs := GitStatus{Path: filepath.Dir(info.Path)}
		select {
		case sem <- struct{}{}:
			// Async.
			wg.Add(1)
			go func(gs GitStatus) {
				defer wg.Done()
				if err := gs.GetStatus(c); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}
				gs.Print(c)
				<-sem
			}(gs)
		default:
			// Sync.
			if err := gs.GetStatus(c); err != nil {
				fmt.Fprintln(os.Stderr, err)
			} else {
				gs.Print(c)
			}
		}
	}

	wg.Wait()

	return nil
}

// GetStatus return git status.
func (gs *GitStatus) GetStatus(c *cli.Context) error {

	var err error

	// Check diff.
	cmd := core.Cmd{Cmd: exec.Command("git", "diff", "--quiet")}
	cmd.Cmd.Dir = gs.Path
	err = cmd.CmdRun()
	if err != nil {
		return err
	}

	if cmd.ExitCode == 0 {
		gs.Diff = false
	} else {
		gs.Diff = true
	}

	if c.Bool("status") {
		cmd := core.Cmd{Cmd: exec.Command("git", "status", "--porcelain", "--branch")}
		cmd.Cmd.Dir = gs.Path
		err = cmd.CmdRun()
		if err != nil {
			return err
		}
		gs.Status = cmd.Stdout.String()
	}

	if c.Bool("ahead") || c.Bool("behind") {
		cmd := core.Cmd{Cmd: exec.Command("git", "status", "--porcelain", "--branch")}
		cmd.Cmd.Dir = gs.Path
		err = cmd.CmdRun()
		if err != nil {
			return err
		}
		stdOut := cmd.Stdout.String()
		if aheadRe.MatchString(stdOut) {
			gs.Ahead = true
		}
		if behindRe.MatchString(stdOut) {
			gs.Behind = true
		}
	}

	return nil
}

// Print is print GitStatus with mutex.
func (gs *GitStatus) Print(c *cli.Context) {
	mu.Lock()
	defer mu.Unlock()

	printStatus := func() {
		if c.Bool("status") {
			fmt.Println(gs.Status)
		}
	}

	// Print path.
	if c.Bool("all") {
		fmt.Println(gs.Path)
		printStatus()
	} else if gs.Diff {
		fmt.Println(gs.Path)
		printStatus()
	} else if c.Bool("ahead") && gs.Ahead {
		fmt.Println(gs.Path)
		printStatus()
	} else if c.Bool("behind") && gs.Behind {
		fmt.Println(gs.Path)
		printStatus()
	}

}
