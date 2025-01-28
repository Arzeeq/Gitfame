package statistic

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"gitlab.com/slon/shad-go/gitfame/internal/flag"
)

func NewStatistic() *Statistic {
	return &Statistic{authors: make(map[string]Author)}
}

type Statistic struct {
	mu      sync.Mutex
	authors map[string]Author
}

func (s *Statistic) AddLines(name string, lines int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	author, ok := s.authors[name]
	if !ok {
		author = Author{Name: name, commitSet: make(map[string]struct{})}
	}
	author.Lines += lines
	s.authors[name] = author
}

func (s *Statistic) AddFiles(name string, files int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	author, ok := s.authors[name]
	if !ok {
		author = Author{Name: name, commitSet: make(map[string]struct{})}
	}
	author.Files += files
	s.authors[name] = author
}

func (s *Statistic) UpdateCommitSet(name string, commits []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	author, ok := s.authors[name]
	if !ok {
		author = Author{Name: name, commitSet: make(map[string]struct{})}
	}
	for _, commit := range commits {
		author.commitSet[commit] = struct{}{}
	}
	s.authors[name] = author
}

func (s *Statistic) GetSorted(flags flag.Flag) []Author {
	s.mu.Lock()
	defer s.mu.Unlock()
	authors := make([]Author, 0)
	for _, author := range s.authors {
		authors = append(authors, author)
	}
	if flags.OrderBy == "lines" {
		sort.SliceStable(authors, func(i, j int) bool {
			if authors[i].Lines == authors[j].Lines {
				if authors[i].Commits() == authors[j].Commits() {
					if authors[i].Files == authors[j].Files {
						return authors[i].Name < authors[j].Name
					}
					return authors[i].Files > authors[j].Files
				}
				return authors[i].Commits() > authors[j].Commits()
			}

			return authors[i].Lines > authors[j].Lines
		})
	} else if flags.OrderBy == "commits" {
		sort.SliceStable(authors, func(i, j int) bool {
			if authors[i].Commits() == authors[j].Commits() {
				if authors[i].Lines == authors[j].Lines {
					if authors[i].Files == authors[j].Files {
						return authors[i].Name < authors[j].Name
					}
					return authors[i].Files > authors[j].Files
				}
				return authors[i].Lines > authors[j].Lines
			}

			return authors[i].Commits() > authors[j].Commits()
		})
	} else if flags.OrderBy == "files" {
		sort.SliceStable(authors, func(i, j int) bool {
			if authors[i].Files == authors[j].Files {
				if authors[i].Lines == authors[j].Lines {
					if authors[i].Commits() == authors[j].Commits() {
						return authors[i].Name < authors[j].Name
					}
					return authors[i].Commits() > authors[j].Commits()
				}
				return authors[i].Lines > authors[j].Lines
			}

			return authors[i].Files > authors[j].Files
		})
	} else {
		ErrUnknownOrderBy := flag.FlagError{}
		ErrUnknownOrderBy.SetMessage(fmt.Sprintf("unknow order by value '%s'", flags.OrderBy))
		panic(ErrUnknownOrderBy)
	}

	return authors
}

type Author struct {
	Name      string              `json:"name"`
	Lines     int                 `json:"-"`
	Files     int                 `json:"-"`
	commitSet map[string]struct{} `json:"-"`
}

func (a *Author) Commits() int {
	return len(a.commitSet)
}

func (a *Author) MarshalJSON() ([]byte, error) {
	type Alias Author
	return json.Marshal(&struct {
		Alias
		Lines   int `json:"lines"`
		Commits int `json:"commits"`
		Files   int `json:"files"`
	}{
		Alias:   Alias(*a),
		Lines:   a.Lines,
		Commits: len(a.commitSet),
		Files:   a.Files,
	})
}
