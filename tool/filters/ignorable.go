package filters

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gobwas/glob"
)

type Ignorable struct {
	patterns []glob.Glob
}

func (ignorable Ignorable) Filter(modified []Modification) []Modification {
	slog.Debug("Ignorable filter.")
	slog.Debug("-----------------")
	var result []Modification
	for _, modification := range modified {
		var include = true
		// TODO what types to handle
		if modification.Type != `D` {
			slog.Debug("Considering", slog.String("name", modification.Name))
			for _, glob := range ignorable.patterns {
				if glob.Match(modification.Name) {
					slog.Debug("Dropping", slog.String("name", modification.Name), slog.String("glob", fmt.Sprintf("%v", glob)))
					include = false
					break
				}
			}
		}
		if include {
			slog.Debug("Keeping", slog.String("name", modification.Name))
			result = append(result, modification)
		}
	}
	return result
}

func CreateIgnorable(config json.RawMessage) (Ignorable, error) {
	var data []string
	err := json.Unmarshal(config, &data)
	return Ignorable{patterns: ToPatterns(data)}, err
}
