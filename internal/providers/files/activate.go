package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/abenz1267/elephant/internal/common"
)

const (
	ActionOpen     = "open"
	ActionOpenDir  = "opendir"
	ActionCopyPath = "copypath"
	ActionCopyFile = "copyfile"
)

func Activate(qid uint32, identifier, action string, arguments string) {
	path := ""

	if action == "" {
		action = ActionOpen
	}

	for k := range paths {
		md5 := md5.Sum([]byte(k))
		md5str := hex.EncodeToString(md5[:])

		if identifier == md5str {
			path = k
			break
		}
	}

	switch action {
	case ActionOpen, ActionOpenDir:
		if action == ActionOpenDir {
			path = filepath.Dir(path)
		}

		run := strings.TrimSpace(fmt.Sprintf("%s xdg-open '%s'", common.LaunchPrefix(config.LaunchPrefix), path))

		if forceTerminalForFile(path) {
			run = common.WrapWithTerminal(run)
		}

		cmd := exec.Command("sh", "-c", run)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actionopen", err)
		} else {
			go func() {
				cmd.Wait()
			}()
		}
	case ActionCopyPath:
		cmd := exec.Command("wl-copy", path)

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actioncopypath", err)
		} else {
			go func() {
				cmd.Wait()
			}()
		}

	case ActionCopyFile:
		cmd := exec.Command("wl-copy", "-t", "text/uri-list", fmt.Sprintf("file://%s", path))

		err := cmd.Start()
		if err != nil {
			slog.Error(Name, "actioncopyfile", err)
		} else {
			go func() {
				cmd.Wait()
			}()
		}
	default:
		slog.Error(Name, "nosuchaction", action)
	}
}

func forceTerminalForFile(file string) bool {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("xdg-mime query default $(xdg-mime query filetype %s)", file))

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	cmd.Dir = homedir

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return false
	}

	if _, ok := terminalApps[strings.TrimSpace(string(out))]; ok {
		return true
	}

	return false
}
