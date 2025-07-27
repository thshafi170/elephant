// Package comm provides functionallity to communitate with elephant
package comm

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/abenz1267/elephant/internal/providers"
)

const (
	ActionQuery    = "query"
	ActionCleanup  = "cleanup"
	ActionActivate = "activate"
)

func StartListen() {
	file := filepath.Join(os.TempDir(), "elephant.sock")
	os.Remove(file)

	l, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: file,
	})
	if err != nil {
		slog.Error("comm", "socket", err)
	}
	defer l.Close()

	slog.Info("comm", "listen", "starting")

	for {
		conn, err := l.AcceptUnix()
		if err != nil {
			slog.Error("comm", "accept", err)
		}

		slog.Info("comm", "connection", "new")

		conn.Write(fmt.Appendf(nil, "session:%d\n", providers.GetSID()))

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		message := scanner.Text()
		slog.Info("comm", "request", message)

		request := strings.Split(message, ";")

		switch request[0] {
		case ActionQuery:
			if len(request) != 4 {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid query request '%s'\n", message))
				continue
			}

			sid, err := strconv.ParseUint(request[1], 10, 32)
			if err != nil {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid query request '%s'\n", message))
				continue
			}

			providers.Query(uint32(sid), request[2], strings.Fields(request[3]), conn)
		case ActionCleanup:
			if len(request) != 2 {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid cleanup request '%s'\n", message))
				continue
			}

			sid, err := strconv.ParseUint(request[1], 10, 32)
			if err != nil {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid cleanup request '%s'\n", message))
				continue
			}

			providers.Cleanup(uint32(sid))
		case ActionActivate:
			if len(request) != 5 {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid activate request '%s'\n", message))
				continue
			}

			sid, err := strconv.ParseUint(request[1], 10, 32)
			if err != nil {
				slog.Error("comm", "requestinvalid", message)
				conn.Write(fmt.Appendf(nil, "error: invalid activate request '%s'\n", message))
				continue
			}

			providers.Activate(uint32(sid), request[2], request[3], request[4])
		default:
			slog.Error("comm", "requestinvalid", request[0])
			conn.Write(fmt.Appendf(nil, "error: invalid action '%s'\n", request[0]))
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("comm", "request", err)
	}
}
