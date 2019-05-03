package daemon

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bitspark/slang/pkg/env"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var SlangVersion string

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// allow every host/origin to make a connection e.g. :8080 -> 5149
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	Root             = &UserID{0}
	ConnectedClients = make(map[*UserID][]*ConnectedClient, 0)
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*ConnectedClient]bool

	// Inbound messages from the clients.
	broadcast chan *AddressableMessage

	// Register requests from the clients.
	register chan *ConnectedClient

	// Unregister requests from clients.
	unregister chan *ConnectedClient
}

type AddressableMessage struct {
	u *UserID
	m []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *AddressableMessage),
		register:   make(chan *ConnectedClient),
		unregister: make(chan *ConnectedClient),
		clients:    make(map[*ConnectedClient]bool),
	}
}

func (h *Hub) broadCastTo(u *UserID, m string) {
	h.broadcast <- &AddressableMessage{u, []byte(m)}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if client.userID != message.u {
					return
				}
				select {
				case client.send <- message.m:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

type Server struct {
	Host   string
	Port   int
	router *mux.Router
	ctx    *context.Context
}

type ConnectedClient struct {
	hub       *Hub
	websocket *websocket.Conn
	userID    *UserID
	send      chan []byte
}

type UserID struct {
	id int
}

func addContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewServer(ctx *context.Context, env *env.Environment) *Server {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)
	srv := &Server{env.HTTP.Address, env.HTTP.Port, r, ctx}
	srv.mountWebServices()
	return srv
}

func (s *Server) AddWebsocket(path string) {
	r := s.router.Path(path)
	r.HandlerFunc(serveWs)
	hub := newHub()
	go hub.run()
	newCtx := context.WithValue(*s.ctx, "hub", hub)
	s.ctx = &newCtx
}

func simpleReader(ws *websocket.Conn) {
	newline := []byte{'\n'}
	space := []byte{' '}
	defer ws.Close()
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		log.Printf("Got Message: %v", message)
	}
}

func writer(c *ConnectedClient) {
	ticker := time.NewTicker(pingPeriod)
	ws := c.websocket
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
	}()
	for {
		select {
		case msg := <-c.send:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			ws.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	hub := r.Context().Value("hub").(*Hub)
	user := Root
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	// if the user is connected append the websocket
	backChannel := make(chan []byte)
	client := &ConnectedClient{hub, ws, user, backChannel}
	hub.register <- client
	go writer(client)
	// we keep this around for now as it serves the websocket ping<>pong
	go simpleReader(ws)
}

func (s *Server) mountWebServices() {
	s.AddService("/operator", DefinitionService)
	s.AddService("/run", RunnerService)
	s.AddService("/share", SharingService)
	s.AddOperatorProxy("/instance")
	s.AddWebsocket("/ws")
}

func (s *Server) AddService(pathPrefix string, services *Service) {
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		(func(endpoint *Endpoint) {
			r.HandleFunc(path, endpoint.Handle)
		})(endpoint)
	}
}

func (s *Server) AddStaticServer(pathPrefix string, directory http.Dir) {
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix, http.FileServer(http.Dir(directory))))
}

func (s *Server) AddOperatorProxy(pathPrefix string) {
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix,
		r.HandlerFunc(proxyRequestToOperator).GetHandler()))
}

func (s *Server) AddRedirect(path string, redirectTo string) {
	r := s.router.Path(path)
	r.Handler(http.RedirectHandler(redirectTo, http.StatusSeeOther))
}

func (s *Server) Run() error {
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), addContext(*s.ctx, handler))
}
