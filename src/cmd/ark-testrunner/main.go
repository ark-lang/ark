package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ark-lang/ark/src/util"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
)

type Job struct {
	Name, Sourcefile          string
	CompilerArgs, RunArgs     []string
	CompilerError, RunError   int
	Input                     string
	CompilerOutput, RunOutput string
}

type Result struct {
	Job            Job
	CompilerError  int
	RunError       int
	CompilerOutput string
	RunOutput      string
}

func ParseJob(filename string) (Job, error) {
	var job Job
	if _, err := toml.DecodeFile(filename, &job); err != nil {
		return Job{}, err
	}

	return job, nil
}

var (
	showOutput    = flag.Bool("show-output", false, "Enable to show output of tests")
	testDirectory = flag.String("test-directory", "./tests/", "The directory in which tests are located")
)

func main() {
	flag.Parse()
	os.Exit(realmain())
}

func realmain() int {
	var dirs []string
	files := make(map[string][]string)

	// Find all toml files in test directory
	filepath.Walk(*testDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relpath, err := filepath.Rel(*testDirectory, path)
		if err != nil {
			return err
		}

		dir, file := filepath.Split(relpath)
		if info.IsDir() {
			if file == "." {
				file = ""
			} else {
				file += "/"
			}

			dirs = append(dirs, file)
		} else if strings.HasSuffix(file, ".toml") {
			files[dir] = append(files[dir], file)
		}
		return nil
	})

	// Sort directories
	sort.Strings(dirs)

	var jobs []Job
	for _, dir := range dirs {
		// Sort files
		sort.Strings(files[dir])

		for _, file := range files[dir] {
			path := filepath.Join(*testDirectory, dir, file)

			// Parse job file
			job, err := ParseJob(path)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				return 1
			}
			job.Sourcefile = filepath.Join(*testDirectory, dir, job.Sourcefile)

			jobs = append(jobs, job)
		}
	}

	// Do jobs
	outBuf := new(bytes.Buffer)
	var results []Result
	for _, job := range jobs {
		idx := strings.LastIndex(job.Sourcefile, ".ark")
		basedir, module := filepath.Split(job.Sourcefile[:idx])
		outpath := filepath.Join(basedir, fmt.Sprintf("%s.test", module))

		// Compile the test program
		buildArgs := []string{"build"}
		buildArgs = append(buildArgs, job.CompilerArgs...)
		buildArgs = append(buildArgs, []string{"-b", basedir, "-o", outpath, module}...)

		outBuf.Reset()
		if *showOutput {
			fmt.Printf("Building test: %s\n", job.Name)
		}

		var err error
		res := Result{Job: job}

		res.CompilerError, err = runCommand(outBuf, "ark", buildArgs...)
		if err != nil {
			fmt.Printf("Error while building test:\n%s\n", err.Error())
			return 1
		}
		res.CompilerOutput = outBuf.String()

		if res.CompilerError != 0 {
			results = append(results, res)
			res.RunError = -1
			continue
		}

		// Run the test program
		outBuf.Reset()
		if *showOutput {
			fmt.Printf("\nRunning test: %s\n", job.Name)
		}

		res.RunError, err = runCommand(outBuf, fmt.Sprintf("./%s", outpath), job.RunArgs...)
		if err != nil {
			fmt.Printf("Error while running test:\n%s\n", err.Error())
			return 1
		}
		res.RunOutput = outBuf.String()

		if *showOutput {
			fmt.Printf("\n")
		}

		// Remove test executable
		if err := os.Remove(outpath); err != nil {
			fmt.Printf("Error while removing test executable:\n%s\n", err.Error())
			return 1
		}

		results = append(results, res)
	}

	// Check results
	numSucceses := 0

	fmt.Printf("Test name       | Build error | Run error | Build output | Run output | Result  \n")
	fmt.Printf("----------------|-------------|-----------|--------------|------------|---------\n")
	for _, res := range results {
		failure := false
		if len(res.Job.Name) > 15 {
			fmt.Printf("%s... |", res.Job.Name[:12])
		} else {
			fmt.Printf("%-15s |", res.Job.Name)
		}

		// Check build errors
		if res.CompilerError == res.Job.CompilerError {
			fmt.Printf("   %s%3d%s (%3d) |", util.TEXT_GREEN, res.CompilerError, util.TEXT_RESET, res.Job.CompilerError)
		} else {
			fmt.Printf("   %s%3d%s (%3d) |", util.TEXT_RED, res.CompilerError, util.TEXT_RESET, res.Job.CompilerError)
			failure = true
		}

		// Check run errors
		if res.RunError == res.Job.RunError {
			fmt.Printf(" %s%3d%s (%3d) |", util.TEXT_GREEN, res.RunError, util.TEXT_RESET, res.Job.RunError)
		} else {
			fmt.Printf(" %s%3d%s (%3d) |", util.TEXT_RED, res.RunError, util.TEXT_RESET, res.Job.RunError)
			failure = true
		}

		// Check build output
		if res.Job.CompilerOutput == "" {
			fmt.Printf("          n/a |")
		} else if res.CompilerOutput == res.Job.CompilerOutput {
			fmt.Printf("    %sMatch%s |", util.TEXT_GREEN, util.TEXT_RESET)
		} else {
			fmt.Printf(" %sMismatch%s |", util.TEXT_RED, util.TEXT_RESET)
			failure = true
		}

		// Check run output
		if res.Job.RunOutput == "" {
			fmt.Printf("        n/a |")
		} else if res.RunOutput == res.Job.RunOutput {
			fmt.Printf("      %sMatch%s |", util.TEXT_GREEN, util.TEXT_RESET)
		} else {
			fmt.Printf("   %sMismatch%s |", util.TEXT_RED, util.TEXT_RESET)
			failure = true
		}

		// Output result
		if failure {
			fmt.Printf(" %sFailure%s\n", util.TEXT_RED, util.TEXT_RESET)
		} else {
			fmt.Printf(" %sSuccess%s\n", util.TEXT_GREEN, util.TEXT_RESET)
			numSucceses += 1
		}
	}

	fmt.Printf("\nTotal: %d / %d tests ran succesfully\n", numSucceses, len(results))
	if numSucceses < len(results) {
		return 1
	}
	return 0
}

func runCommand(out io.Writer, cmd string, args ...string) (int, error) {
	// Run the test program
	command := exec.Command(cmd, args...)

	// Output handling
	ow := out
	if *showOutput {
		ow = io.MultiWriter(out, os.Stdout)
	}
	command.Stdout, command.Stderr = ow, ow

	// Start the test
	if err := command.Start(); err != nil {
		return -1, err
	}

	// Check the exit status
	if err := command.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), nil
			}
		} else {
			return -1, err
		}
	}

	return 0, nil
}
