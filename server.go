package main

import (
	"log"
	"net/http"
	"os"
)

type server struct {
	uploadDir      string
	auth           *authService
	frontendOrigin string
}

func newServer(cfg appConfig) *server {
	return &server{
		uploadDir:      cfg.UploadDir,
		auth:           newAuthService(cfg.Auth),
		frontendOrigin: cfg.FrontendOrigin,
	}
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
