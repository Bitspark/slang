package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Bitspark/slang/pkg/env"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

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
	SlangVersion string
	// currently the `gorilla/websocket` library feels rather verbose (ping<>pong)- which is ok for ATM
	// as long as we keep the implementation small enough.
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

// Envelop functions as an addressable pack of messages that is only sent to
// specific user and their connections. We use an array here, so we can send multiple message
// to a single user at once. This comes in handy when implementing burst protection.
// Imagine we where to send 10k message per second over a single websocket - not such a good idea.
// What we want instead is collect most of them inside a timeframe and send them together.
// Receiving end must of course know that message can hold 1+N message and dispatch accordingly.
type envelop struct {
	receiver *UserID
	messages []*message
}

// In the end we need to represent a message we want to send to a connected client.
// The only sensible way we can achive that without loosing too much information is to encode the entire
// array of messages as json array.
func (e *envelop) Bytes() []byte {
	body, _ := json.Marshal(e.messages)
	return body
}

// An Envelop is able to hold an unlimited amount of messages intended for one
// receiver e.g. UserID, this is a variadic method that allows appending *message(s).
func (e *envelop) append(m ...*message) {
	e.messages = append(e.messages, m...)
}

// A message is holding data and a topic - we do this so the interface can listen on different topics
// and discard recieved data easier or better decide where to route the information.
type message struct {
	Topic   Topic       `json:"topic"`
	Payload interface{} `json:"payload"`
}

type Server struct {
	Host   string
	Port   int
	router *mux.Router
	ctx    *context.Context
}

// ConnectedClient holds everything we need to know about a connection that
// was made with a websocket
type ConnectedClient struct {
	hub       *Hub
	websocket *websocket.Conn
	userID    *UserID
	// Send data through this channel in order to get it send through the websocket
	// currently this is what the `hub` uses to send it's recieved message through a websocket.
	send chan []byte
}

// UserID represents an Identifier for a user of the system
// This is also intended to deliver `envelops` to the correct `ConnectedClients`
// instead of sending a message to all connected clients. Currently we do not have real users.
// The variable acts more as a placeholder and thinking vehicle to reminde oneself that the system
// might have more than one user in the future - I know, YAGNI.
type UserID struct {
	id int
}

// We want types around the topic as this makes it easier to work with.
// A Topic in this case is meant for easier distinction of what type of message is being send.
// Basically we attach type information for the other end of the connected client.
type Topic int

const (
	Port     Topic = iota
	Operator       // currently unused but displays the intended usage
)

// Since we can't send proper type information over the wire, we send a string
// representation instead.
func (t Topic) String() string {
	return [...]string{"Port", "Operator"}[t]
}

// This encodes a `Topic` to Json using it's string representation
func (t Topic) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func addContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *envelop),
		register:   make(chan *ConnectedClient),
		unregister: make(chan *ConnectedClient),
		clients:    make(map[*ConnectedClient]bool),
	}
}

// Send a message to single user on all his connections
// This API is probably a little volatile so use with caution and don't reach deep into it.
func (h *Hub) broadCastTo(u *UserID, topic Topic, data interface{}) {
	var messages []*message
	messages = append(messages, &message{topic, data})
	h.broadcast <- &envelop{u, messages}
}

func (h *Hub) run() {
	var (
		incomingEnvelop *envelop
		ok              bool
		minTimer        <-chan time.Time
		maxTimer        <-chan time.Time
	)
	min := 100 * time.Millisecond
	max := 500 * time.Millisecond
	postBox := make(map[*UserID]*envelop)

	actualSend := func(letter *envelop) {
		for client := range h.clients {
			// this might become PINA as iterating all clients to find only those which we want to address
			// could get expensive - maybe look up the clients by `userID` in the first place.
			if client.userID != letter.receiver {
				continue
			}
			// wrapping `<-` with a `select` and `default` makes it non-blocking if there is no reciever on the other reading off the channel.
			// The channel is buffered meaning that we can sucessfully write into it as long as a reciever is pulling data from the other end.
			select {
			case client.send <- letter.Bytes():
				// message written
			default:
				// buffer of channel full - no one reading? Let's disconnect them.
				close(client.send)
				delete(h.clients, client)
			}
		}
	}

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok = h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case incomingEnvelop = <-h.broadcast:
			// Keep adding messages from other envelop for at least `min` amount of time to the postbox.
			minTimer = time.After(min)
			// We only want to set the how long we are going to keep adding messages once.
			if maxTimer == nil {
				maxTimer = time.After(max)
			}
			// If we already have a envelop for that user, appened the contents of the current envelop
			// to the one that is already scheduled to be sent.
			if _, ok := postBox[incomingEnvelop.receiver]; ok {
				// What we do here is to grow the single envelop`s messages.
				postBox[incomingEnvelop.receiver].append(incomingEnvelop.messages...)
			} else {
				// If do not yet have a letter to be sent create one for the current receiver.
				postBox[incomingEnvelop.receiver] = incomingEnvelop
			}

		case <-minTimer:
			// We have not seen new letters since `min` time which means we can now send
			// send the letters to all users.
			minTimer, maxTimer = nil, nil
			for _, e := range postBox {
				actualSend(e)
			}
			// we also need to clear reset our letters since we have sent them all
			postBox = make(map[*UserID]*envelop)
		case <-maxTimer:
			// Enough time has now passed and we should now send what we have already collected.
			// This useful in scenarios where we send just enough data to
			minTimer, maxTimer = nil, nil
			for _, e := range postBox {
				actualSend(e)
			}
			postBox = make(map[*UserID]*envelop)
		}
	}
}

func (c *ConnectedClient) waitOnIncoming() {
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

func (c *ConnectedClient) waitOnOutgoing() {
	ticker := time.NewTicker(pingPeriod)
	ws := c.websocket
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// This happens if there was a call to `close(c.send)`,
				// which means the client was disconnect from the hub via `unregister <- c` or otherwise closed.
				//
				// And actually we should never reach this in normal execution, as the client should no
				// longer be in the list of registered client, otherwise we would be reading an endless amount of <nil>.
				return
			}
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			ws.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // websocket likly closed
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	hub := GetHub(r)
	user := Root
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}
	// Create a new client for each connection we recieve
	// attaching the user makes it possible to send message to multiple
	// open browsers that are accociated with the user.
	//
	// Then channel is a bytes channel with the size of 256, I am unsure of
	// wether this is enough and what happends if we have message greater than that.
	// Might be a better idea to make channel of a type with a smaller size the get the same effect.
	// Without the uncertainty.
	client := &ConnectedClient{hub, ws, user, make(chan []byte, 256)}
	hub.register <- client

	// Part of the RFC is this Ping<>Pong thing which we need to have both in the writer and reader of the
	// socket connection. see -> https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API/Writing_WebSocket_servers#Pings_and_Pongs_The_Heartbeat_of_WebSockets
	//
	// Apart from just playing ping pong this go routine
	// waits on messages from the `hub` that it can forward outwards to the connected client
	go client.waitOnOutgoing()

	// we keep this around for now as it serves the websocket ping<>pong
	// if we want to allow the UI to control the daemon via websockets, this is the place.
	go client.waitOnIncoming()

	// so basically only returns if the ping pong fails or there is another error.
}

func NewServer(ctx *context.Context, env *env.Environment) *Server {
	r := mux.NewRouter().StrictSlash(true)
	srv := &Server{env.HTTP.Address, env.HTTP.Port, r, ctx}
	srv.mountWebServices()
	return srv
}

func (s *Server) Handler() http.Handler {
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(s.router)
	return addContext(*s.ctx, handler)
}

func (s *Server) AddWebsocket(path string) {
	r := s.router.Path(path)
	r.HandlerFunc(serveWs)
	hub := newHub()

	// Don't know yet if that is good idea
	// Maybe should make the `hub` a singleton instead of shoving it
	// into the context where it needs more typing to be safe
	// What we want is for the hub to be available in everywhare we might want to send messages
	// to a connected client.
	newCtx := SetHub(*s.ctx, hub)
	s.ctx = &newCtx
	go hub.run()
}

func (s *Server) mountWebServices() {
	s.AddService("/operator", DefinitionService)
	s.AddService("/run", RunnerService)
	s.AddService("/share", SharingService)
	s.AddService("/instances", InstanceService)
	s.AddService("/instance", RunningInstanceService)
	s.AddWebsocket("/ws")
}

func (s *Server) AddService(pathPrefix string, services *Service) {
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		path := path
		(func(endpoint *Endpoint) {
			r.HandleFunc(path, endpoint.Handle)
		})(endpoint)
	}
}

func (s *Server) AddStaticServer(pathPrefix string, directory http.Dir) {
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix, http.FileServer(directory)))
}

func (s *Server) AddRedirect(path string, redirectTo string) {
	r := s.router.Path(path)
	r.Handler(http.RedirectHandler(redirectTo, http.StatusSeeOther))
}

func (s *Server) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Handler())
}
