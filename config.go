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
	AllowedOrigins []string
	Auth           authConfig
	Mongo          mongoConfig
}

type authConfig struct {
	ClientID        string
	ClientSecret    string
	RedirectURI     string
	JWTSecret       string
	RequiredGuildID string
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

	frontendOrigins := splitOrigins(os.Getenv("FRONTEND_ORIGIN"))
	if len(frontendOrigins) == 0 {
		frontendOrigins = []string{"http://localhost:5173"}
	}
	frontend := frontendOrigins[0]

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
		AllowedOrigins: mergeOrigins(frontendOrigins, splitOrigins(os.Getenv("ALLOWED_ORIGINS"))),
		Auth:           authCfg,
		Mongo:          mongoCfg,
	}, nil
}

func splitOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := normalizeOrigin(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}
	return origins
}

func mergeOrigins(groups ...[]string) []string {
	seen := make(map[string]struct{})
	var merged []string

	for _, group := range groups {
		for _, origin := range group {
			origin = normalizeOrigin(origin)
			if _, ok := seen[origin]; ok {
				continue
			}
			seen[origin] = struct{}{}
			merged = append(merged, origin)
		}
	}

	return merged
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	for strings.HasSuffix(origin, "/") {
		origin = strings.TrimSuffix(origin, "/")
	}
	return origin
}

func readAuthConfig() (authConfig, error) {
	clientID := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_SECRET"))
	redirectURI := strings.TrimSpace(os.Getenv("DISCORD_REDIRECT_URI"))
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	requiredGuildID := strings.TrimSpace(os.Getenv("DISCORD_REQUIRED_GUILD_ID"))

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
	if requiredGuildID == "" {
		return authConfig{}, fmt.Errorf("DISCORD_REQUIRED_GUILD_ID es requerido")
	}

	return authConfig{
		ClientID:        clientID,
		ClientSecret:    clientSecret,
		RedirectURI:     redirectURI,
		JWTSecret:       secret,
		RequiredGuildID: requiredGuildID,
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
