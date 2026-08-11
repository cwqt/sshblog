package ui

import (
	"errors"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// feedOpenedMsg reports the result of launching the system opener so the model
// can surface any failure in the footer.
type feedOpenedMsg struct{ err error }

// openFeed returns a command that hands target — the hosted feed URL, or a
// local file path in development — to the OS's default handler (a browser,
// usually) via `open` / `xdg-open` / `start`.
//
// Caveat: the opener runs on the machine hosting the SSH server, not on the
// client. It only reaches a browser when the server shares a desktop with the
// person connecting (e.g. `ssh localhost` during local `make dev`). On a
// headless remote host there is nothing to open, and the command reports the
// launch error rather than doing anything — remote readers instead point their
// feed reader at the public URL directly.
func openFeed(target string) tea.Cmd {
	return func() tea.Msg {
		if target == "" {
			return feedOpenedMsg{err: errors.New("no RSS feed available")}
		}
		name, args := openerCommand(target)
		// Start (not Run) so a slow or long-lived opener never blocks the UI;
		// this still reports launch failures such as a missing xdg-open.
		if err := exec.Command(name, args...).Start(); err != nil {
			return feedOpenedMsg{err: err}
		}
		return feedOpenedMsg{}
	}
}

// openerCommand picks the platform's default-application launcher.
func openerCommand(target string) (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		return "open", []string{target}
	case "windows":
		// The empty "" is start's window-title argument, so a quoted target
		// isn't mistaken for one.
		return "cmd", []string{"/c", "start", "", target}
	default: // linux, *bsd, etc.
		return "xdg-open", []string{target}
	}
}
