package comm

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func Send(req string) {
	conn, err := net.Dial("unix", filepath.Join(os.TempDir(), "elephant.sock"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\n", req)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		text := scanner.Text()

		if strings.Contains(text, "qid;") {
			if strings.Contains(text, ";done") {
				break
			}

			continue
		}

		fmt.Println(text)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("comm", "request", err)
	}
}
