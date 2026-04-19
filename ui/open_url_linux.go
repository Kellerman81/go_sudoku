//go:build linux && !android

package ui

import "os/exec"

func openURL(url string) {
	_ = exec.Command("xdg-open", url).Start()
}
