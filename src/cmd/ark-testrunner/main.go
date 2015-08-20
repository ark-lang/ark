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

	var dirs []string
	files := make(map[string][]string)

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
			return nil
		}

		if strings.HasSuffix(file, ".toml") {
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

			job, err := ParseJob(path)
			if err != nil {
				fmt.Printf("%s\n", err.Error())
				os.Exit(1)
			}
			job.Sourcefile = filepath.Join(*testDirectory, dir, job.Sourcefile)

			jobs = append(jobs, job)
		}
	}

	// Do jobs
	var results []Result
	for _, job := range jobs {
		idx := strings.LastIndex(job.Sourcefile, ".ark")
		basedir := filepath.Dir(job.Sourcefile)
		module := filepath.Base(job.Sourcefile[:idx])
		outpath := filepath.Join(basedir, fmt.Sprintf("%s.test", module))

		res := Result{Job: job}

		// Compile the test program
		buildArgs := []string{"build"}
		buildArgs = append(buildArgs, job.CompilerArgs...)
		buildArgs = append(buildArgs, []string{"-b", basedir, "-o", outpath, module}...)
		buildCmd := exec.Command("ark", buildArgs...)

		// Output handling
		outBuf := new(bytes.Buffer)
		if *showOutput {
			fmt.Printf("Building test: %s\n", job.Name)
			buildCmd.Stdout = io.MultiWriter(outBuf, os.Stdout)
			buildCmd.Stderr = io.MultiWriter(outBuf, os.Stdout)
		} else {
			buildCmd.Stdout = outBuf
			buildCmd.Stderr = outBuf
		}

		// Start the build
		if err := buildCmd.Start(); err != nil {
			fmt.Printf("Error while starting build command:\n%s\n", err.Error())
			os.Exit(1)
		}

		// Check the exit status
		if err := buildCmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					res.CompilerError = status.ExitStatus()
				}
			} else {
				fmt.Printf("Error while running build command:\n%s\n", err.Error())
				os.Exit(1)
			}
		}
		res.CompilerOutput = outBuf.String()

		if res.CompilerError != 0 {
			results = append(results, res)
			res.RunError = -1
			continue
		}

		// Run the test program
		runCmd := exec.Command(fmt.Sprintf("./%s", outpath), job.RunArgs...)

		// Output handling
		outBuf.Reset()
		if *showOutput {
			fmt.Printf("\nRunning test: %s\n", job.Name)
			runCmd.Stdout = io.MultiWriter(outBuf, os.Stdout)
			runCmd.Stderr = io.MultiWriter(outBuf, os.Stdout)
		} else {
			runCmd.Stdout = outBuf
			runCmd.Stderr = outBuf
		}

		// Start the test
		if err := runCmd.Start(); err != nil {
			fmt.Printf("Error while starting test binary:\n%s\n", err.Error())
			os.Exit(1)
		}

		// Check the exit status
		if err := runCmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					res.RunError = status.ExitStatus()
				}
			} else {
				fmt.Printf("Error while running test binary:\n%s\n", err.Error())
				os.Exit(1)
			}
		}
		res.RunOutput = outBuf.String()

		if *showOutput {
			fmt.Printf("\n")
		}

		// Remove test executable
		if err := os.Remove(outpath); err != nil {
			fmt.Printf("Error while removing test executable:\n%s\n", err.Error())
		}

		results = append(results, res)
	}

	// Check results
	numSucceses := 0

	fmt.Printf("Test name       | Build error | Build output | Run error | Run output | Result \n")
	fmt.Printf("----------------|-------------|--------------|-----------|------------|---------\n")
	for _, res := range results {
		if len(res.Job.Name) > 15 {
			fmt.Printf("%s... |", res.Job.Name[:12])
		} else {
			fmt.Printf("%-15s |", res.Job.Name)
		}

		// Check build errors
		buildWrongError := false
		if res.CompilerError != res.Job.CompilerError {
			buildWrongError = true
			fmt.Printf("   %s%3d%s", util.TEXT_RED, res.CompilerError, util.TEXT_RESET)
		} else {
			fmt.Printf("   %s%3d%s", util.TEXT_GREEN, res.CompilerError, util.TEXT_RESET)
		}
		fmt.Printf(" (%3d) |", res.Job.CompilerError)

		// Check build output
		buildWrongOutput := false
		if res.Job.CompilerOutput == "" {
			fmt.Printf("          n/a |")
		} else if res.CompilerOutput != res.Job.CompilerOutput {
			buildWrongOutput = true
			fmt.Printf(" %sMismatch%s |", util.TEXT_RED, util.TEXT_RESET)
		} else {
			fmt.Printf("    %sMatch%s |", util.TEXT_GREEN, util.TEXT_RESET)
		}

		// Check run errors
		runWrongError := false
		if res.RunError != res.Job.RunError {
			runWrongError = true
			fmt.Printf(" %s%3d%s", util.TEXT_RED, res.RunError, util.TEXT_RESET)
		} else {
			fmt.Printf(" %s%3d%s", util.TEXT_GREEN, res.RunError, util.TEXT_RESET)
		}
		fmt.Printf(" (%3d) |", res.Job.RunError)

		// Check run output
		runWrongOutput := false
		if res.Job.RunOutput == "" {
			fmt.Printf("        n/a |")
		} else if res.RunOutput != res.Job.RunOutput {
			runWrongOutput = true
			fmt.Printf("   %sMismatch%s |", util.TEXT_RED, util.TEXT_RESET)
		} else {
			fmt.Printf("      %sMatch%s |", util.TEXT_GREEN, util.TEXT_RESET)
		}

		// Output result
		failure := buildWrongError || buildWrongOutput || runWrongError || runWrongOutput
		if failure {
			fmt.Printf(" %sFailure%s\n", util.TEXT_RED, util.TEXT_RESET)
		} else {
			fmt.Printf(" %sSuccess%s\n", util.TEXT_GREEN, util.TEXT_RESET)
			numSucceses += 1
		}
	}

	fmt.Printf("\nTotal: %d / %d tests ran succesfully\n", numSucceses, len(results))

	if numSucceses < len(results) {
		os.Exit(1)
	}
	os.Exit(0)
}
