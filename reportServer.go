package tsbWriter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

var b *Broker

// A single Broker will be created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
//
type Broker struct {

	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients.  (The values
	// are just booleans and are meaningless.)
	//
	clients map[chan string]bool

	// Channel into which new clients can be pushed
	//
	newClients chan chan string

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan string

	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	//
	messages chan string
}

// This Broker method starts a new goroutine.  It handles
// the addition & removal of clients, as well as the broadcasting
// of messages out to clients that are currently attached.
//
func (b *Broker) Start() {

	// Start a goroutine
	//
	go func() {

		// Loop endlessly
		//
		for {

			// Block until we receive from one of the
			// three following channels.
			select {

			case s := <-b.newClients:

				// There is a new client attached and we
				// want to start sending them messages.
				b.clients[s] = true

			case s := <-b.defunctClients:

				// A client has dettached and we want to
				// stop sending them messages.
				delete(b.clients, s)
				close(s)

			case msg := <-b.messages:
				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s := range b.clients {
					s <- msg
				}

			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
//
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan string)

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- messageChan
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly.
	for {

		// Read from our messageChan.
		msg, open := <-messageChan

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		m := make(map[string]string)

		m["data"] = msg

		byteS, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintf(w, "data: %s\n\n", byteS)
		if err != nil {
			panic(err)
		}

		// Flush the response.  This is only possible if
		// the repsonse supports streaming.
		f.Flush()
	}
}

// Handler for the main page, which we wire up to the
// route at "/" below in `main`.
//
func handler(w http.ResponseWriter, r *http.Request) {

	// Did you know Golang's ServeMux matches only the
	// prefix of the request URL?  It's true.  Here we
	// insist the path is just "/".
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	t := getIndex()
	// Render the template, writing to `w`.
	t.Execute(w, "friend")
}
func getIndex() *template.Template {
	return template.Must(template.New("Buffer Logs").Parse(indexHtml))
}

// Main routine
func startServer() {
	// Make a new Broker instance
	b = &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}

	// Start processing events
	b.Start()

	// Make b the HTTP handler for "/events/".  It can do
	// this because it has a ServeHTTP method.  That method
	// is called in a separate goroutine for each
	// request to "/events/".
	http.Handle("/events/", b)

	// Generate a constant stream of events that get pushed
	// into the Broker's messages channel and are then broadcast
	// out to any clients that are attached.
	// go func() {
	// 	for i := 0; ; i++ {

	// 		// Create a little message to send to clients,
	// 		// including the current time.
	// 		b.messages <- fmt.Sprintf("%d - the time is %v", i, time.Now())

	// 		// Print a nice log message and sleep for 5s.
	// 		log.Printf("Sent message %d ", i)
	// 		time.Sleep(5e9)

	// 	}
	// }()

	// When we get a request at "/", call `handler`
	// in a new goroutine.
	http.Handle("/", http.HandlerFunc(handler))
	fmt.Println("Serving buffer logs on localhost:8999")

	// Start the server and listen forever on port 8000.
	if err := http.ListenAndServe(":8999", nil); err != nil {
		panic(err)
	}
}

const indexHtml = `<!DOCTYPE html>
<html>
<head>
	<title>Buffer Report</title>
</head>
<body>
	<script type="text/javascript">
	    // Create a new HTML5 EventSource
	    var source = new EventSource('/events/');
		var test = 1;
	    // Create a callback for when a new message is received.
	    source.onmessage = function(e) {
	        js = JSON.parse(e.data)
	        
	        document.body.innerHTML = '<span style="white-space: pre-line; font-family: monospace"><tt>' + js.data + '</tt></span>';
	    };
	</script>
</body>
</html>`
