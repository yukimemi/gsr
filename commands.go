package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
			EnvVar: "GSR_SHOW_ALL",
			Name:   "all",
			Usage:  "Show all directry",
		},
	}

	mu = new(sync.Mutex)
)

// GlobalAction is Global gsr command.
var GlobalAction = func(c *cli.Context) error {

	var (
		err  error
		root string

		wg  = new(sync.WaitGroup)
		opt = file.Option{
			Matches: []string{`\.git$`},
			Recurse: true,
		}
	)

	if c.NArg() > 0 {
		root = c.Args().First()
	} else {
		cmd := core.Cmd{Cmd: exec.Command("ghq", "root")}
		err = cmd.CmdRun()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		root = strings.TrimRight(cmd.Stdout.String(), "\n")
	}

	if !file.IsExistDir(root) {
		msg := fmt.Sprintf("[%v] is not exist", root)
		fmt.Fprintln(os.Stderr, msg)
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
		wg.Add(1)
		go func(gs GitStatus) {
			defer wg.Done()
			if err := gs.GetStatus(c); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			gs.Print(c)
		}(gs)
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
		cmd := core.Cmd{Cmd: exec.Command("git", "status", "--short")}
		cmd.Cmd.Dir = gs.Path
		err = cmd.CmdRun()
		if err != nil {
			return err
		}
		gs.Status = cmd.Stdout.String()
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
	}

}
