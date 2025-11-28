package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var uploadDir = "uploads"
var jwtSecret []byte
var discordConfig oauth2Config

type config struct {
	addr      string
	uploadDir string
}

type oauth2Config struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

type discordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

type jwtClaims struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	jwt.RegisteredClaims
}

type contextKey string

const userContextKey contextKey = "user"

type fileEntry struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type renameRequest struct {
	NewName string `json:"newName"`
}

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("no se pudo cargar .env: %v", err)
	}

	cfg := readConfig()
	uploadDir = cfg.uploadDir

	if err := readAuthConfig(); err != nil {
		log.Fatalf("configuración de autenticación inválida: %v", err)
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("no se pudo crear la carpeta de subida: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/auth/discord", authDiscordHandler)
	mux.HandleFunc("/auth/discord/callback", authCallbackHandler)
	mux.HandleFunc("/auth/logout", logoutHandler)
	mux.HandleFunc("/auth/me", authRequired(meHandler))
	mux.HandleFunc("/upload", authRequired(uploadHandler))
	mux.HandleFunc("/files", authRequired(listHandler))
	mux.HandleFunc("/files/", authRequired(fileHandler))

	log.Printf("servidor escuchando en %s, carpeta de subidas: %s", cfg.addr, uploadDir)
	if err := http.ListenAndServe(cfg.addr, corsMiddleware(logRequest(mux))); err != nil {
		log.Fatalf("servidor detenido: %v", err)
	}
}

func readConfig() config {
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

	return config{
		addr:      addr,
		uploadDir: upload,
	}
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "solo se permite POST", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "no se pudo procesar el formulario", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "archivo requerido con campo 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName := header.Filename
	if custom := strings.TrimSpace(r.FormValue("filename")); custom != "" {
		fileName = custom
	}

	safeName, err := sanitizeName(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dstPath := filepath.Join(uploadDir, safeName)
	if _, err := os.Stat(dstPath); err == nil {
		http.Error(w, "ya existe un archivo con ese nombre", http.StatusConflict)
		return
	}

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		log.Printf("error al abrir destino: %v", err)
		http.Error(w, "no se pudo guardar el archivo", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("error al copiar archivo: %v", err)
		http.Error(w, "no se pudo escribir el archivo", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "archivo subido", "name": safeName})
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		log.Printf("error al leer carpeta: %v", err)
		http.Error(w, "no se pudo listar archivos", http.StatusInternalServerError)
		return
	}

	var files []fileEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			log.Printf("error al leer info de archivo %s: %v", e.Name(), err)
			continue
		}
		files = append(files, fileEntry{
			Name:     e.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, files)
}

func sanitizeName(name string) (string, error) {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == "" {
		return "", fmt.Errorf("nombre de archivo requerido")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("nombre de archivo inválido")
	}
	return name, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error al serializar respuesta: %v", err)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/files/")
	if name == "" || strings.Contains(name, "/") {
		http.Error(w, "ruta inválida", http.StatusBadRequest)
		return
	}

	currentName, err := sanitizeName(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		serveFile(w, r, currentName)
	case http.MethodDelete:
		deleteFile(w, r, currentName)
	case http.MethodPut:
		renameFile(w, r, currentName)
	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(uploadDir, name)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "archivo no encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error al acceder a archivo: %v", err)
		http.Error(w, "no se pudo abrir el archivo", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.Error(w, "nombre inválido", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, path)
}

func deleteFile(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(uploadDir, name)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "archivo no encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error al acceder a archivo: %v", err)
		http.Error(w, "no se pudo eliminar el archivo", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.Error(w, "nombre inválido", http.StatusBadRequest)
		return
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "archivo no encontrado", http.StatusNotFound)
			return
		}
		log.Printf("error al eliminar archivo: %v", err)
		http.Error(w, "no se pudo eliminar el archivo", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "archivo eliminado", "name": name})
}

func renameFile(w http.ResponseWriter, r *http.Request, currentName string) {
	var payload renameRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "cuerpo JSON inválido", http.StatusBadRequest)
		return
	}

	newName, err := sanitizeName(payload.NewName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if newName == currentName {
		writeJSON(w, http.StatusOK, map[string]string{"message": "archivo sin cambios", "name": newName})
		return
	}

	oldPath := filepath.Join(uploadDir, currentName)
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "archivo no encontrado", http.StatusNotFound)
		return
	}

	newPath := filepath.Join(uploadDir, newName)
	if _, err := os.Stat(newPath); err == nil {
		http.Error(w, "ya existe un archivo con el nuevo nombre", http.StatusConflict)
		return
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		log.Printf("error al renombrar archivo: %v", err)
		http.Error(w, "no se pudo renombrar el archivo", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "archivo renombrado", "name": newName})
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func readAuthConfig() error {
	clientID := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("DISCORD_CLIENT_SECRET"))
	redirectURI := strings.TrimSpace(os.Getenv("DISCORD_REDIRECT_URI"))
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))

	if clientID == "" {
		return fmt.Errorf("DISCORD_CLIENT_ID es requerido")
	}
	if clientSecret == "" {
		return fmt.Errorf("DISCORD_CLIENT_SECRET es requerido")
	}
	if redirectURI == "" {
		return fmt.Errorf("DISCORD_REDIRECT_URI es requerido")
	}
	if secret == "" {
		return fmt.Errorf("JWT_SECRET es requerido")
	}

	discordConfig = oauth2Config{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
	jwtSecret = []byte(secret)

	log.Printf("configuración de autenticación cargada correctamente")
	return nil
}

func authDiscordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	state := generateRandomState()

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
		Path:     "/",
	})

	authURL := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify&state=%s",
		url.QueryEscape(discordConfig.clientID),
		url.QueryEscape(discordConfig.redirectURI),
		url.QueryEscape(state),
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		http.Redirect(w, r, "http://localhost:5173?error=access_denied", http.StatusTemporaryRedirect)
		return
	}

	if code == "" {
		http.Error(w, "código de autorización requerido", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		http.Error(w, "estado inválido - posible ataque CSRF", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		log.Printf("error al intercambiar código: %v", err)
		http.Error(w, "error al autenticar con Discord", http.StatusUnauthorized)
		return
	}

	user, err := fetchDiscordUser(accessToken)
	if err != nil {
		log.Printf("error al obtener usuario: %v", err)
		http.Error(w, "error al obtener perfil de usuario", http.StatusInternalServerError)
		return
	}

	jwtToken, err := generateJWT(user)
	if err != nil {
		log.Printf("error al generar JWT: %v", err)
		http.Error(w, "error al crear sesión", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
		Path:     "/",
	})

	http.Redirect(w, r, "http://localhost:5173", http.StatusTemporaryRedirect)
}

func exchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", discordConfig.clientID)
	data.Set("client_secret", discordConfig.clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", discordConfig.redirectURI)

	resp, err := http.PostForm("https://discord.com/api/oauth2/token", data)
	if err != nil {
		return "", fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("código de estado: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return result.AccessToken, nil
}

func fetchDiscordUser(accessToken string) (*discordUser, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear petición: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("código de estado: %d", resp.StatusCode)
	}

	var user discordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error al decodificar usuario: %w", err)
	}

	return &user, nil
}

func generateJWT(user *discordUser) (string, error) {
	claims := jwtClaims{
		UserID:        user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		Avatar:        user.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

func authRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "no autenticado", http.StatusUnauthorized)
			return
		}

		claims, err := validateJWT(cookie.Value)
		if err != nil {
			http.Error(w, "token inválido o expirado", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "solo se permite POST", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "sesión cerrada"})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(userContextKey).(*jwtClaims)
	if !ok {
		http.Error(w, "no se pudo obtener usuario", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       claims.UserID,
		"username":      claims.Username,
		"discriminator": claims.Discriminator,
		"avatar":        claims.Avatar,
	})
}

func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("error al generar estado aleatorio: %v", err)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}
