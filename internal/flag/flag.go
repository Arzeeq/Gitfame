package flag

import (
	"fmt"
	"slices"

	"github.com/spf13/pflag"
)

type Flag struct {
	RepositoryPath string
	Revision       string
	OrderBy        string
	UseCommitter   bool
	Format         string
	Extensions     []string
	Languages      []string
	Exclude        []string
	RestrictTo     []string
}

func Parse() (Flag, error) {
	var flag Flag
	pflag.StringVar(&flag.RepositoryPath, "repository", "./", "path to repository")
	pflag.StringVar(&flag.Revision, "revision", "HEAD", "pointer to commit")
	pflag.StringVar(&flag.OrderBy, "order-by", "lines", fmt.Sprintf("the sorting order of the rows: %v", getOrderByValues()))
	pflag.BoolVar(&flag.UseCommitter, "use-committer", false, "use committer instead of author in calculations")
	pflag.StringVar(&flag.Format, "format", "tabular", fmt.Sprintf("output format: %v", getFormatValues()))
	pflag.StringSliceVar(&flag.Extensions, "extensions", []string{}, "list of extensions that need to be counted")
	pflag.StringSliceVar(&flag.Languages, "languages", []string{}, "list of languages that need to be counted")
	pflag.StringSliceVar(&flag.Exclude, "exclude", []string{}, "list of patterns that exclude files from the calculation")
	pflag.StringSliceVar(&flag.RestrictTo, "restrict-to", []string{}, "list of patterns that excludes all files that do not satisfy any of the patterns in the set")
	pflag.Parse()
	return flag, validate(flag)
}

func validate(flag Flag) error {
	if v := getOrderByValues(); !slices.Contains(v, flag.OrderBy) {
		return &FlagError{fmt.Sprintf("'%s' is unknown, should be one of %v", flag.OrderBy, v)}
	}
	if v := getFormatValues(); !slices.Contains(v, flag.Format) {
		return &FlagError{fmt.Sprintf("'%s' is unknown, should be one of %v", flag.Format, v)}
	}

	return nil
}

func getOrderByValues() []string {
	return []string{"lines", "commits", "files"}
}

func getFormatValues() []string {
	return []string{"tabular", "csv", "json", "json-lines"}
}

type FlagError struct {
	message string
}

func (e *FlagError) Error() string {
	return fmt.Sprintf("gitfame: %s", e.message)
}

func (e *FlagError) SetMessage(msg string) {
	e.message = msg
}
