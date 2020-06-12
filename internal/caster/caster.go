package caster

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Caster could be described as a collection of Mountpoints.
// HTTP(S) server implementing the semantics of the NTRIPv2 protocol.
// Sources POST (publish) streaming data to unique Mountpoints (URL Paths)
// on the Caster.
// Clients subscribe to streams via GET requests to Mountpoints.
type Caster struct {
	sync.RWMutex
	// A Collection of URL paths to which data is being streamed
	Mounts  map[string]*Mountpoint
	Timeout time.Duration
}

// ListenHTTP starts a HTTP server given a port in the format of the net/http library
func (caster *Caster) ListenHTTP(port string) error {
	server := &http.Server{
		Addr:    port,
		Handler: http.HandlerFunc(caster.RequestHandler),
	}
	return server.ListenAndServe()
}

// RequestHandler function for all incoming HTTP(S) requests
func (caster *Caster) RequestHandler(w http.ResponseWriter, r *http.Request) {
	conn := NewConnection(w, r)
	defer conn.Request.Body.Close()

	logger := log.WithFields(log.Fields{
		"request_id": conn.ID(),
		"path":       conn.Request.URL.Path,
		"method":     conn.Request.Method,
		"source_ip":  conn.Request.RemoteAddr,
	})

	conn.Writer.Header().Set("X-Request-Id", conn.ID())
	conn.Writer.Header().Set("Ntrip-Version", "Ntrip/2.0")
	conn.Writer.Header().Set("Server", "NTRIP GoCaster")

	switch conn.Request.Method {
	case http.MethodPost:
		mount := NewMountpoint(conn)
		if err := caster.AddMountpoint(mount); err != nil {
			conn.Writer.WriteHeader(http.StatusConflict)
			conn.Writer.(http.Flusher).Flush()
			logger.Error(err.Error())
			return
		}
		conn.Writer.(http.Flusher).Flush()

		logger.Info("Mountpoint Connected - " + mount.Source.Request.URL.Path)
		err := mount.Broadcast(caster.Timeout)
		logger.Info("Mountpoint Disconnected - " + err.Error())

		caster.DeleteMountpoint(mount.Source.Request.URL.Path)
		return

	case http.MethodGet:
		mount := caster.GetMountpoint(conn.Request.URL.Path)
		if mount == nil {
			conn.Writer.WriteHeader(http.StatusNotFound)
			logger.Error("No Existing Mountpoint") // Should probably reserve logger.Error for server errors
			return
		}

		conn.Writer.Header().Set("Content-Type", "application/octet-stream")
		logger.Info("Accepted Client Connection")
		mount.RegisterSubscriber(conn)
		for { // TODO: Come up with a Connection struct method name which makes sense for this
			select {
			case data, _ := <-conn.channel:
				fmt.Fprintf(conn.Writer, "%s", data)
				conn.Writer.(http.Flusher).Flush() // TODO: Add timeout on write
			case <-conn.Request.Context().Done():
				mount.DeregisterSubscriber(conn)
				logger.Info("Client Disconnected - client closed connection")
				return
			case <-mount.Source.Request.Context().Done():
				logger.Info("Client Disconnected - mountpoint closed connection")
				return
			case <-time.After(caster.Timeout):
				logger.Info("Client Disconnected - timout writing to client")
				return
			}
		}

	default:
		conn.Writer.WriteHeader(http.StatusNotImplemented)
		logger.Error("Request Method Not Implemented")
	}
}

// AddMountpoint adds a Mounpoint object to a Casters collection of Mounpoints
func (caster *Caster) AddMountpoint(mount *Mountpoint) (err error) {
	caster.Lock()
	defer caster.Unlock()
	if _, ok := caster.Mounts[mount.Source.Request.URL.Path]; ok {
		return errors.New("Mountpoint in use")
	}

	caster.Mounts[mount.Source.Request.URL.Path] = mount
	return nil
}

// DeleteMountpoint removes a Mounpoint object from a Casters collection of Mounpoints
func (caster *Caster) DeleteMountpoint(id string) {
	caster.Lock()
	defer caster.Unlock()
	delete(caster.Mounts, id)
}

// GetMountpoint returns a mount object from the a Casters collection of Mountpoints
// given it's ID as a string
func (caster *Caster) GetMountpoint(id string) (mount *Mountpoint) {
	caster.RLock()
	defer caster.RUnlock()
	return caster.Mounts[id]
}
