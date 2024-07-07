package filters

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/gobwas/glob"
)

type Comments struct {
	comments []Comment
	root     string
}

type Comment struct {
	patterns   []glob.Glob
	singleLine *regexp.Regexp
	multiLine  *regexp.Regexp
}

type CommentData struct {
	Patterns       []string `json:"patterns"`
	SingleLine     string   `json:"single"`
	MultiLineStart string   `json:"multi-start"`
	MultiLineEnd   string   `json:"multi-end"`
}

type Needle func(Comment) bool

func search(name string) Needle {
	return func(comment Comment) bool {
		for _, glob := range comment.patterns {
			if glob.Match(name) {
				slog.Debug("Matches:", slog.String("glob", fmt.Sprintf("%v", glob)), slog.String("name", name))
				return true
			}
		}
		return false
	}
}

func regexpEscape(literal string) string {
	for _, c := range "*+.?^{}[]" {
		esc := string(c)
		literal = strings.ReplaceAll(literal, esc, `\`+esc)
	}
	return literal
}

func fetchRevision(root string, hash string, name string) (string, error) {
	var buffer bytes.Buffer
	cmd := exec.Command("git", "show", hash, "--", name)
	cmd.Dir = root
	cmd.Stdout = &buffer
	err := cmd.Run()
	return buffer.String(), err
}

func fetchRevisions(root string, modification Modification) (string, string, error) {
	before, err := fetchRevision(root, modification.OldHash, modification.Name)
	if err == nil {
		after, err := fetchRevision(root, modification.NewHash, modification.Name)
		return before, after, err
	}
	return before, "", err
}

func normalize(info Comment, content string) string {
	if info.singleLine == nil && info.multiLine == nil {
		return content
	}
	if info.singleLine != nil {
		content = info.singleLine.ReplaceAllString(content, "${1}")
	}
	if info.multiLine != nil {
		content = info.multiLine.ReplaceAllString(content, "")
	}
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(line)
		}
	}
	return sb.String()
}

func (comments Comments) Filter(modified []Modification) []Modification {
	slog.Debug("Comments filter.")
	slog.Debug("----------------")
	var result []Modification
	for _, modification := range modified {
		if modification.Type == `M` {
			slog.Debug("Considering", slog.String("name", modification.Name))
			index := slices.IndexFunc(comments.comments, search(modification.Name))
			if index >= 0 {
				before, after, err := fetchRevisions(comments.root, modification)
				if err == nil {
					before = normalize(comments.comments[index], before)
					after = normalize(comments.comments[index], after)
					if strings.Compare(before, after) != 0 {
						slog.Debug("Change does not seem to be comment-only, keeping.")
						result = append(result, modification)
					} else {
						slog.Debug("Comment-only change, dropping.")
					}
				}
			} else {
				slog.Debug("Keeping", slog.String("name", modification.Name))
				result = append(result, modification)
			}
		} else {
			result = append(result, modification)
		}
	}
	return result
}

func CreateComments(root string, config json.RawMessage) (Comments, error) {
	var data []CommentData
	err := json.Unmarshal(config, &data)
	var comments = make([]Comment, len(data))
	for idx, comment := range data {
		// of course, this won't work for fancy source code, e.g. var x = "this is a // comment";
		if len(comment.SingleLine) > 0 {
			comments[idx].singleLine = regexp.MustCompile(fmt.Sprintf("(?m)^(.*?)[ \\t]*%s.*$", regexpEscape(comment.SingleLine)))
		}
		if len(comment.MultiLineStart) > 0 {
			comments[idx].multiLine = regexp.MustCompile(fmt.Sprintf("(?sm)%s.+?%s", regexpEscape(comment.MultiLineStart), regexpEscape(comment.MultiLineEnd)))
		}
		comments[idx].patterns = ToPatterns(comment.Patterns)
	}
	return Comments{comments: comments, root: root}, err
}
