package tui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/aymanbagabas/go-osc52/v2"
)

// writeClipboard copies text to the system clipboard. It prefers the platform's
// native clipboard utility (which reports real errors and is honored by every
// terminal, including GNOME Terminal/VTE), and falls back to an OSC 52 terminal
// escape when no native tool is available or it fails — eg. on a Wayland session
// without wl-clipboard, or over SSH. OSC 52 works in terminals like Ghostty but
// is fire-and-forget (success can't be confirmed) and is not honored by some
// terminals, so it is the last resort rather than the default.
//
// It returns the name of the backend used ("wl-copy", "xclip", "osc52", ...).
func writeClipboard(text string) (string, error) {
	if tool, args := nativeClipboardCmd(); tool != "" {
		cmd := exec.Command(tool, args...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return tool, nil
		}
		// Native tool present but failed (eg. xclip can't reach the X server
		// on a Wayland session). Fall through to OSC 52.
	}

	// OSC 52 fallback. A bare sequence is emitted: tmux with `set-clipboard`
	// on/external relays it to the outer terminal, so TmuxMode (DCS
	// passthrough, which requires `allow-passthrough on`) is intentionally
	// not used. Written to stderr to stay out of Bubble Tea's stdout renderer;
	// both share the pty, so tmux/the terminal still see the sequence.
	if _, err := osc52.New(text).WriteTo(os.Stderr); err != nil {
		return "", err
	}
	return "osc52", nil
}

// nativeClipboardCmd returns the clipboard command and args appropriate to the
// current OS and (on Linux/BSD) display server, or "" if none is found.
func nativeClipboardCmd() (string, []string) {
	switch runtime.GOOS {
	case "darwin":
		if p, err := exec.LookPath("pbcopy"); err == nil {
			return p, nil
		}
	case "windows":
		if p, err := exec.LookPath("clip"); err == nil {
			return p, nil
		}
	default: // linux, *bsd
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			if p, err := exec.LookPath("wl-copy"); err == nil {
				return p, nil
			}
		}
		if os.Getenv("DISPLAY") != "" {
			if p, err := exec.LookPath("xclip"); err == nil {
				return p, []string{"-selection", "clipboard"}
			}
			if p, err := exec.LookPath("xsel"); err == nil {
				return p, []string{"--clipboard", "--input"}
			}
		}
	}
	return "", nil
}
