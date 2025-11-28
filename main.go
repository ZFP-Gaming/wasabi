package main

import (
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

const uploadDir = "uploads"

type fileEntry struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type renameRequest struct {
	NewName string `json:"newName"`
}

func main() {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("no se pudo crear la carpeta de subida: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/files", listHandler)
	mux.HandleFunc("/files/", renameHandler)

	addr := ":8080"
	log.Printf("servidor escuchando en %s, carpeta de subidas: %s", addr, uploadDir)
	if err := http.ListenAndServe(addr, logRequest(mux)); err != nil {
		log.Fatalf("servidor detenido: %v", err)
	}
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

func renameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "solo se permite PUT", http.StatusMethodNotAllowed)
		return
	}

	// expected path: /files/{nombre}
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

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
