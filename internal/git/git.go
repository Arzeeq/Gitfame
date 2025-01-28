package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gitlab.com/slon/shad-go/gitfame/configs"
	"gitlab.com/slon/shad-go/gitfame/internal/flag"
	"gitlab.com/slon/shad-go/gitfame/internal/statistic"
	"gitlab.com/slon/shad-go/gitfame/pkg/progressbar"
)

const cntGoroutines = 8

func CalculateFiles(flag flag.Flag) *statistic.Statistic {
	defer func() {
		if err := recover(); err != nil {
			_, err := io.WriteString(os.Stderr, fmt.Sprintf("%v\n", err))
			if err != nil {
				panic(err)
			}
			os.Exit(1)
		}
	}()
	wg := sync.WaitGroup{}

	stats := statistic.NewStatistic()
	fileChan := make(chan string, 1)

	worker := func() {
		for {
			file := <-fileChan
			switch file {
			case "":
				wg.Done()
				return
			default:
				err := calculateFile(file, flag, stats)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	files, err := ListFiles(flag)
	if err != nil {
		_, e := io.WriteString(os.Stderr, err.Error()+"\n")
		if e != nil {
			panic(e)
		}
		os.Exit(1)
	}

	for i := 0; i < cntGoroutines; i++ {
		wg.Add(1)
		go worker()
	}

	progBar := progressbar.NewProgressBar(5)
	for i, file := range files {
		fileChan <- file
		progBar.Draw(100 * i / len(files))
	}
	close(fileChan)

	wg.Wait()

	return stats
}

func calculateFile(file string, flag flag.Flag, stats *statistic.Statistic) error {
	cmd := exec.Command("git", "blame", "--incremental", flag.Revision, "--", file)
	cmd.Dir = flag.RepositoryPath
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	nameMap := make(map[string]string)      // key: commit hash, value: committer/author name
	commitsMap := make(map[string][]string) // key: committer/author name, value: list of commits
	strs := strings.Split(string(output), "\n")
	strs = strs[:len(strs)-1]
	if len(strs) == 0 {
		return calculateEmptyFile(file, flag, stats)
	}
	for i := 0; i < len(strs); i++ {
		str := strs[i]
		columns := strings.Split(str, " ")
		hash := columns[0]
		lines, err := strconv.Atoi(columns[3])
		if err != nil {
			return err
		}

		if _, ok := nameMap[hash]; !ok {
			var authorOffset int
			if flag.UseCommitter {
				authorOffset = 5
			} else {
				authorOffset = 1
			}

			tmp := strings.Split(strs[i+authorOffset], " ")
			name := strings.Join(tmp[1:], " ")
			nameMap[hash] = name
			commitsMap[name] = append(commitsMap[name], hash)

			i += 9 // skip header
		}
		for ; i < len(strs) && !strings.HasPrefix(strs[i], "filename"); i++ {
		}
		stats.AddLines(nameMap[hash], lines)
	}

	for name, commits := range commitsMap {
		stats.AddFiles(name, 1)
		stats.UpdateCommitSet(name, commits)
	}

	return nil
}

func calculateEmptyFile(file string, flag flag.Flag, stats *statistic.Statistic) error {
	cmd := exec.Command("git", "log", flag.Revision, "--", file)
	cmd.Dir = flag.RepositoryPath
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	strs := strings.Split(string(output), "\n")
	strs = strs[:len(strs)-1]
	hash := strings.Split(strs[0], " ")[1]
	tmp := strings.Split(strs[1], " ")
	name := strings.Join(tmp[1:len(tmp)-1], " ")
	stats.AddFiles(name, 1)
	stats.UpdateCommitSet(name, []string{hash})

	return nil
}

func ListFiles(flag flag.Flag) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", flag.Revision)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = flag.RepositoryPath

	err := cmd.Run()
	if err != nil {
		return nil, &GitFilesError{stderr.String()}
	}

	files := strings.Split(stdout.String(), "\n")
	filtered, err := filter(files[:len(files)-1], flag)
	if err != nil {
		return nil, err
	}

	return filtered, nil
}

func filter(files []string, flag flag.Flag) ([]string, error) {
	// get allowed extensions
	extensions := configs.GetLanguageExtensions()
	allowedExtensions := flag.Extensions
	for _, lang := range flag.Languages {
		ext, ok := extensions[lang]
		if !ok {
			fmt.Fprintf(os.Stderr, "gitfame: warning: %s language is unknown\n", lang)
			continue
		}
		allowedExtensions = append(allowedExtensions, ext...)
	}

	// filter files
	filtered := make([]string, 0)
	for _, file := range files {
		var extension bool
		var exclude bool
		var restrictTo bool

		// check extensions
		extension = (len(flag.Extensions) == 0 && len(flag.Languages) == 0)
		for _, ext := range allowedExtensions {
			if strings.HasSuffix(file, ext) {
				extension = true
				break
			}
		}

		// check exclude
		for _, exc := range flag.Exclude {
			ok, err := filepath.Match(exc, file)
			if err != nil {
				return nil, &GitFilesError{fmt.Sprintf("invalid Glob %s", exc)}
			}
			if ok {
				exclude = true
				break
			}
		}

		// check restrict-to
		restrictTo = (len(flag.RestrictTo) == 0)
		for _, restr := range flag.RestrictTo {
			ok, err := filepath.Match(restr, file)
			if err != nil {
				return nil, &GitFilesError{fmt.Sprintf("invalid Glob %s", restr)}
			}
			if ok {
				restrictTo = true
				break
			}
		}

		// filter file
		if extension && !exclude && restrictTo {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

type GitFilesError struct {
	message string
}

func (e *GitFilesError) Error() string {
	return fmt.Sprintf("gitfame: %s", e.message)
}
