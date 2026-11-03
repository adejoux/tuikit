// Package editor opens a file in the user's external editor ($VISUAL, then
// $EDITOR) from a Bubble Tea program, suspending the TUI while it runs.
package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ErrNoEditor is returned when neither VISUAL nor EDITOR is set.
var ErrNoEditor = errors.New("no editor configured: set VISUAL or EDITOR")

// FinishedMsg is sent when the editor exits. Path is the edited file.
type FinishedMsg struct {
	Path string
	Err  error
}

// Command returns the editor command line from VISUAL, then EDITOR.
func Command() string {
	if editor := strings.TrimSpace(os.Getenv("VISUAL")); editor != "" {
		return editor
	}
	return strings.TrimSpace(os.Getenv("EDITOR"))
}

// Open opens path in the external editor. A line above zero positions the
// cursor there when the editor is known to support it.
func Open(path string, line int) tea.Cmd {
	editor := Command()
	if editor == "" {
		return func() tea.Msg { return FinishedMsg{Path: path, Err: ErrNoEditor} }
	}
	cmd := exec.Command("sh", "-c", editor+` "$@"`, "tuikit-editor")
	cmd.Args = append(cmd.Args, Args(editor, path, line)...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			err = fmt.Errorf("editor: %w", err)
		}
		return FinishedMsg{Path: path, Err: err}
	})
}

// Args returns the arguments that open path at line for editor, the editor
// command line as found in VISUAL or EDITOR. Editors that are not known to
// support positioning, and lines below 1, just get the path.
func Args(editor, path string, line int) []string {
	if line < 1 {
		return []string{path}
	}
	fields := strings.Fields(editor)
	if len(fields) == 0 {
		return []string{path}
	}
	switch filepath.Base(fields[0]) {
	case "vi", "vim", "nvim", "gvim", "mvim", "nano", "emacs", "emacsclient", "hx", "helix", "micro", "kak":
		return []string{fmt.Sprintf("+%d", line), path}
	case "code", "code-insiders", "codium", "cursor", "zed":
		return []string{"-g", fmt.Sprintf("%s:%d", path, line)}
	case "subl", "sublime_text":
		return []string{fmt.Sprintf("%s:%d", path, line)}
	default:
		return []string{path}
	}
}
