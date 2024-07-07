package filters

import (
	"log/slog"
	"relevan-see/config"

	"github.com/gobwas/glob"
)

type Modification struct {
	Name    string
	Type    string // see https://git-scm.com/docs/git-diff#Documentation/git-diff.txt---diff-filterACDMRTUXB82308203
	OldHash string
	NewHash string
}

type Filter interface {
	Filter(files []Modification) []Modification
}

func Init(root string, config config.ConfigData) ([]Filter, error) {
	result := make([]Filter, len(config.Filters))
	for index, filter := range config.Filters {
		switch filter.Name {
		case `ignorable`:
			ignorable, err := CreateIgnorable(filter.Data)
			if err != nil {
				return result, err
			}
			result[index] = ignorable
		case `comments`:
			comments, err := CreateComments(root, filter.Data)
			if err != nil {
				return result, err
			}
			result[index] = comments
		default:
			slog.Warn("Ignoring unknown filter", slog.String(`name`, filter.Name))
		}
	}
	return result, nil
}

func ToPatterns(list []string) []glob.Glob {
	var result []glob.Glob
	for _, pattern := range list {
		g, err := glob.Compile(pattern)
		if err == nil {
			result = append(result, g)
		} else {
			slog.Warn("Invalid glob pattern", slog.String("pattern", pattern))
		}
	}
	return result
}
