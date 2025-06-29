package session

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"potatossh/internal/database"
	"potatossh/internal/terminal"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

type Session struct {
	Id          string
	ws_conn     *websocket.Conn
	ssh_client  *ssh.Client
	ssh_session *ssh.Session
	stdin       io.WriteCloser
	stdout      io.Reader
	Server      database.Server
	new_data    chan rune
	ws_done     chan struct{}
	term        *terminal.Terminal
	template    *template.Template
}

var upgrader = websocket.Upgrader{}

func connectToHost(user, host, pass string) (*ssh.Client, *ssh.Session, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	client, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}

func NewSession(server database.Server, template *template.Template) (*Session, error) {
	client, session, err := connectToHost(server.User, fmt.Sprintf("%s:%d", server.Address, server.Port), server.Password)
	if err != nil {
		return nil, err
	}

	// session input and output
	serverIn, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	serverOut, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return &Session{
		Id:          uuid.New().String(),
		ws_conn:     nil,
		ssh_client:  client,
		ssh_session: session,
		stdin:       serverIn,
		stdout:      serverOut,
		Server:      server,
		new_data:    make(chan rune),
		ws_done:     nil,
		term:        terminal.NewTerminal(serverIn, server.Name),
		template:    template,
	}, nil
}

func (s *Session) Start() error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // enable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := s.ssh_session.RequestPty("xterm-256color", 40, 80, modes); err != nil {
		return err
	}

	if err := s.ssh_session.Shell(); err != nil {
		return err
	}

	fmt.Printf("Staring session %s with %s (%s).\n", s.Id, s.Server.Name, s.Server.Address)

	go s.collectStdOut()

	return nil
}

func (s *Session) AttachWebSocket(w http.ResponseWriter, r *http.Request) error {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	if s.ws_conn != nil {
		// disconnect
		fmt.Println("Previous WebSocket connection exist. Closing.")
		s.ws_conn.Close()
	}
	s.ws_conn = ws
	s.ws_done = make(chan struct{})
	fmt.Printf("The client %s attached to session %s (%s - %s).\n", s.ws_conn.RemoteAddr().String(), s.Id, s.Server.Name, s.Server.Address)
	go s.wsSender()
	go s.wsPinger()
	s.sendStdin()
	fmt.Printf("The client %s dettached from session %s  (%s - %s).\n", s.ws_conn.RemoteAddr().String(), s.Id, s.Server.Name, s.Server.Address)
	s.ws_conn = nil
	return nil
}

func (s *Session) Disconnect() {
	s.ssh_client.Close()
	s.ssh_session.Close()
	s.ws_conn.Close()
}

func (s *Session) collectStdOut() {
	// producer
	reader := bufio.NewReader(s.stdout)
	for {
		if c, _, err := reader.ReadRune(); err != nil {
			fmt.Println(c, err)
			if err == io.EOF {
				return
			} else {
				log.Fatal(err)
			}
		} else {
			s.new_data <- c
		}
	}
}

func (s *Session) Terminal() *terminal.Terminal {
	return s.term
}

func (s *Session) renderHTML() []byte {
	var buffer bytes.Buffer
	s.template.ExecuteTemplate(&buffer, "codeblock", s)
	if s.term.TitleUpdate {
		s.term.TitleUpdate = false
		s.template.ExecuteTemplate(&buffer, "title_oob", s)
	}
	return buffer.Bytes()
}

func (s *Session) wsSender() {
	// consumer
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	doSend := false
	for {
		// fmt.Println("wsSender!")
		select {
		case <-ticker.C:
			if doSend {
				s.ws_conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := s.ws_conn.WriteMessage(websocket.TextMessage, s.renderHTML()); err != nil {
					return
				}
				doSend = false
			}
		case r := <-s.new_data:
			s.term.ProcessCharacter(r)
			doSend = true
		case <-s.ws_done:
			return
		}
	}
}

func (s *Session) wsPinger() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.ws_conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				log.Println("ping:", err)
			}
		case <-s.ws_done:
			return
		}
	}
}

type KeyMessage struct {
	Keys string `json:"keys"`
}

type SizeMessage struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

type BrowserMessage struct {
	Type string `json:"type"`
	*KeyMessage
	*SizeMessage
}

func (s *Session) sendStdin() {
	s.ws_conn.SetReadLimit(8192)
	s.ws_conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.ws_conn.SetPongHandler(func(string) error { s.ws_conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		_, message, err := s.ws_conn.ReadMessage()
		if err != nil {
			fmt.Println("ReadMessage error:", err)
			break
		}
		var msg BrowserMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Println("Unmarshal error:", err.Error())
			continue
		}

		if msg.Type == "keyboard" {
			if _, err := s.stdin.Write([]byte(msg.Keys)); err != nil {
				fmt.Println("sendStdin Write")
				break
			}
		} else if msg.Type == "size" {
			err := s.ssh_session.WindowChange(msg.Rows, msg.Columns)
			if err != nil {
				fmt.Println("WindowChange error:", err)
				continue
			}
			s.term.SetSize(msg.Rows, msg.Columns)
		}
	}
	close(s.ws_done)
}

func (s *Session) InjectStdin(bytes []byte) error {
	_, err := s.stdin.Write(bytes)
	return err
}
