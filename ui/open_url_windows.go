//go:build windows

package ui

import "os/exec"

func openURL(url string) {
	_ = exec.Command("cmd", "/c", "start", url).Start()
}
