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
	// Root defines a single instance of our user as we currently do not have
	// I use that to get the rest of the code to think about multi tanentcy
	Root = &UserID{0}
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*ConnectedClient]bool

	// Message that should be send to specific connections
	broadcast chan *envelop

	// Register requests from the clients.
	register chan *ConnectedClient

	// Unregister requests from clients.
	unregister chan *ConnectedClient
}

// Envelop functions as an addressable message that is only sent to
// specific users and their connections
type envelop struct {
	reciever *UserID
	message  []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *envelop),
		register:   make(chan *ConnectedClient),
		unregister: make(chan *ConnectedClient),
		clients:    make(map[*ConnectedClient]bool),
	}
}

// send a message to single user on all his connections
func (h *Hub) broadCastTo(u *UserID, m string) {
	h.broadcast <- &envelop{u, []byte(m)}
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
				// this might become PINA as iterating all clients to find only those which we want to address
				// could get expensive
				if client.userID != message.reciever {
					return
				}
				// wrapping `<-` with a `select` and `default` makes it non-blocking if there is no reciever on the other end
				// What happens if there is reciever? We can assume this connection has been dropped/closed.
				select {
				case client.send <- message.message:
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

func simpleReader(c *ConnectedClient) {
	ws := c.websocket
	newline := []byte{'\n'}
	space := []byte{' '}
	defer func() {
		c.hub.unregister <- c
	}()

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
	// create a new client for each connection we recieve
	// attaching the user makes it possible to send message to multiple
	// open browsers that are accociated with the user.
	client := &ConnectedClient{hub, ws, user, make(chan []byte)}
	hub.register <- client

	// Part of the RFC is this Ping<>Pong thing which we need to have both in the writer and reader of the
	// socket connection.
	// see -> https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers#Pings_and_Pongs_The_Heartbeat_of_WebSockets
	// Apart from just playing ping pong this go routine waits on messages from the `hub` that it can forward to the connected client
	go writer(client)

	// we keep this around for now as it serves the websocket ping<>pong
	// if we want to allow the UI to control the daemon via websockets, this is the place.
	go simpleReader(client)

	// so basically only returns if the ping pong fails or there is another error.
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
