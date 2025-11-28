package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type appConfig struct {
	Addr           string
	UploadDir      string
	FrontendOrigin string
	Auth           authConfig
	Mongo          mongoConfig
}

type authConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	JWTSecret    string
}

type mongoConfig struct {
	URI        string
	Database   string
	Collection string
}

func loadAppConfig() (appConfig, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	addr := port
	if !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	upload := strings.TrimSpace(os.Getenv("UPLOAD_DIR"))
	if upload == "" {
		upload = "uploads"
	}

	frontend := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
	if frontend == "" {
		frontend = "http://localhost:5173"
	}

	authCfg, err := readAuthConfig()
	if err != nil {
		return appConfig{}, err
	}

	mongoCfg, err := readMongoConfig()
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		Addr:           addr,
		UploadDir:      upload,
		FrontendOrigin: frontend,
		Auth:           authCfg,
		Mongo:          mongoCfg,
	}, nil
}

func readAuthConfig() (authConfig, error) {
	clientID := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_SECRET"))
	redirectURI := strings.TrimSpace(os.Getenv("DISCORD_REDIRECT_URI"))
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if clientID == "" {
		return authConfig{}, fmt.Errorf("DISCORD_CLIENT_ID es requerido")
	}
	if clientSecret == "" {
		return authConfig{}, fmt.Errorf("DISCORD_CLIENT_SECRET es requerido")
	}
	if redirectURI == "" {
		return authConfig{}, fmt.Errorf("DISCORD_REDIRECT_URI es requerido")
	}
	if secret == "" {
		return authConfig{}, fmt.Errorf("JWT_SECRET es requerido")
	}

	return authConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		JWTSecret:    secret,
	}, nil
}

func readMongoConfig() (mongoConfig, error) {
	uri := strings.TrimSpace(os.Getenv("MONGO_URL"))
	if uri == "" {
		return mongoConfig{}, fmt.Errorf("MONGO_URL es requerido")
	}

	db := strings.TrimSpace(os.Getenv("MONGO_DB"))
	if db == "" {
		db = "bot"
	}

	collection := strings.TrimSpace(os.Getenv("MONGO_COLLECTION"))
	if collection == "" {
		collection = "intros"
	}

	return mongoConfig{
		URI:        uri,
		Database:   db,
		Collection: collection,
	}, nil
}

func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("no se pudo abrir %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, val, ok := strings.Cut(text, "=")
		if !ok {
			return fmt.Errorf("linea %d invalida en %s", line, path)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("linea %d sin clave en %s", line, path)
		}
		val = strings.TrimSpace(val)
		if n := len(val); n >= 2 && ((val[0] == '"' && val[n-1] == '"') || (val[0] == '\'' && val[n-1] == '\'')) {
			val = val[1 : n-1]
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("no se pudo definir %s: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error leyendo %s: %w", path, err)
	}
	return nil
}
