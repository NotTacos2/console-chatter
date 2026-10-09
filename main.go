package main

import (
	//"bufio"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "23324"
)

var clients = map[ssh.Session]bool{}

func send(msg string) {
	for s := range clients {
		fmt.Fprintln(s, msg)
	}
}

func main() {
	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			logging.Middleware(),
			activeterm.Middleware(),

			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					username := sess.User()
					clients[sess] = true
					fmt.Fprintln(sess, "This doesn't show your draft messages so make sure you're typing slowly")
					send(fmt.Sprintf("%s has joined! Say hi!", username))
					//next(sess)
					buf := make([]byte, 1024) // I hate go so much
					linebuf := ""
					for {
						mes, err := sess.Read(buf)
						if err != nil {
							break
						}
						line := string(buf[:mes])
						linebuf += line

						if strings.Contains(line, "\n") || strings.Contains(line, "\r") { // this is making me crazy aaaa
							realmsg := strings.TrimSpace(linebuf)
							if realmsg != "" {
								//fmt.Fprintln(sess, fmt.Sprintf("[%s] %s", username, realmsg))

								send(fmt.Sprintf("[%s] %s", username, realmsg))
							}
							linebuf = ""
						}
					}

					// scanner doubles my messages
					/*scanner := bufio.NewScanner(sess)
					for scanner.Scan() {
						realmsg := scanner.Text()
						if realmsg != "" {
							send(fmt.Sprintf("[%s] %s", username, realmsg))
						}
					}*/
					delete(clients, sess)
					send(fmt.Sprintf("%s has left us... ", username))
				}
			},
		),
	)
	if err != nil {
		log.Error("Could not start server.", "error", err)
	}
	log.Info("Starting SSH server", "host", host, "port", port)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not start server", "error", err)
	}
}
