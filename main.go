package main

import "log"

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("no se pudo cargar .env: %v", err)
	}

	cfg, err := loadAppConfig()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}

	server, err := newServer(cfg)
	if err != nil {
		log.Fatalf("no se pudo inicializar el servidor: %v", err)
	}

	if err := server.ensureUploadDir(); err != nil {
		log.Fatalf("no se pudo crear la carpeta de subida: %v", err)
	}

	server.listen(cfg.Addr)
}
