package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	server   http.Server
	listener net.Listener
	done     chan struct{}
	log      *log.Logger
	db       *sql.DB
}

func NewServer(listen string, db *sql.DB, log *log.Logger) (*Server, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	addrToListen, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addrToListen)
	if err != nil {
		return nil, err
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	serv := &Server{
		server: http.Server{
			Addr:              listen,
			Handler:           router,
			ReadTimeout:       3 * time.Second,
			ReadHeaderTimeout: 3 * time.Second,
		},
		listener: listener,
		log:      log,
		done:     make(chan struct{}),
		db:       db,
	}
	serv.makeRouter(router)
	return serv, nil
}

func (s *Server) Serve(ctx context.Context) {
	errC := make(chan error)
	go func() {
		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errC <- err
		}
		close(errC)
	}()
	stop := ctx.Done()
	for {
		select {
		case <-stop:
			stop = nil
			s.shutdown()
		case err := <-errC:
			if err != nil {
				s.log.Printf("Error serving HTTP: %s", err.Error())
			}
			close(s.done)
			return
		}
	}
}

func (s *Server) shutdown() {
	ctx, canc := context.WithTimeout(context.Background(), 30*time.Second)
	if err := s.server.Shutdown(ctx); err != nil {
		s.log.Printf("error shutting down: %s", err)
	}
	canc()
}

func (s *Server) Done() <-chan struct{} {
	return s.done
}

func (s *Server) makeRouter(r *gin.Engine) {
	r.GET("/promotions/:id", s.GetPromotion)
}
