package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var uploadDir = "uploads"

type config struct {
	addr      string
	uploadDir string
}

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

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("no se pudo crear la carpeta de subida: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/files", listHandler)
	mux.HandleFunc("/files/", fileHandler)

	log.Printf("servidor escuchando en %s, carpeta de subidas: %s", cfg.addr, uploadDir)
	if err := http.ListenAndServe(cfg.addr, logRequest(mux)); err != nil {
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
