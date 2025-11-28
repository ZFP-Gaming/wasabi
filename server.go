package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type server struct {
	uploadDir        string
	auth             *authService
	frontendOrigin   string
	mongoClient      *mongo.Client
	introsCollection *mongo.Collection
}

func newServer(cfg appConfig) (*server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear cliente de mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("no se pudo conectar a mongo: %w", err)
	}

	return &server{
		uploadDir:        cfg.UploadDir,
		auth:             newAuthService(cfg.Auth),
		frontendOrigin:   cfg.FrontendOrigin,
		mongoClient:      client,
		introsCollection: client.Database(cfg.Mongo.Database).Collection(cfg.Mongo.Collection),
	}, nil
}

func (s *server) ensureUploadDir() error {
	return os.MkdirAll(s.uploadDir, 0o755)
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/discord", s.authDiscordHandler)
	mux.HandleFunc("/auth/discord/callback", s.authCallbackHandler)
	mux.HandleFunc("/auth/logout", s.logoutHandler)
	mux.HandleFunc("/auth/me", s.authRequired(s.meHandler))
	mux.HandleFunc("/upload", s.authRequired(s.uploadHandler))
	mux.HandleFunc("/files", s.authRequired(s.listHandler))
	mux.HandleFunc("/files/", s.authRequired(s.fileHandler))
	mux.HandleFunc("/intro", s.authRequired(s.introHandler))

	return corsMiddleware(s.frontendOrigin, logRequest(mux))
}

func (s *server) listen(addr string) {
	log.Printf("servidor escuchando en %s, carpeta de subidas: %s", addr, s.uploadDir)
	if err := http.ListenAndServe(addr, s.routes()); err != nil {
		log.Fatalf("servidor detenido: %v", err)
	}
}
