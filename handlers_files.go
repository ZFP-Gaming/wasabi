package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type fileEntry struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type renameRequest struct {
	NewName string `json:"newName"`
}

var allowedUploadExts = map[string]bool{
	".mp3": true,
	".ogg": true,
	".wav": true,
	".m4a": true,
}

func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
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

	originalExt := strings.ToLower(filepath.Ext(header.Filename))
	if originalExt == "" {
		originalExt = strings.ToLower(filepath.Ext(fileName))
	}
	if !allowedUploadExts[originalExt] {
		http.Error(w, "formato no soportado, usa mp3, ogg, wav o m4a", http.StatusBadRequest)
		return
	}

	safeName, err := sanitizeName(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	finalName := ensureMP3Name(safeName)
	dstPath := filepath.Join(s.uploadDir, finalName)
	if _, err := os.Stat(dstPath); err == nil {
		http.Error(w, "ya existe un archivo con ese nombre", http.StatusConflict)
		return
	}

	if originalExt == ".mp3" {
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
	} else {
		if err := s.convertAndSaveAsMP3(file, originalExt, dstPath); err != nil {
			log.Printf("error al convertir a mp3: %v", err)
			http.Error(w, "no se pudo convertir el archivo a mp3", http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "archivo subido", "name": finalName})
}

func (s *server) listHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(s.uploadDir)
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

func (s *server) fileHandler(w http.ResponseWriter, r *http.Request) {
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
		s.serveFile(w, r, currentName)
	case http.MethodDelete:
		s.deleteFile(w, r, currentName)
	case http.MethodPut:
		s.renameFile(w, r, currentName)
	default:
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
	}
}

func (s *server) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(s.uploadDir, name)
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

func (s *server) deleteFile(w http.ResponseWriter, r *http.Request, name string) {
	path := filepath.Join(s.uploadDir, name)
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

func (s *server) renameFile(w http.ResponseWriter, r *http.Request, currentName string) {
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

	oldPath := filepath.Join(s.uploadDir, currentName)
	if _, err := os.Stat(oldPath); errors.Is(err, os.ErrNotExist) {
		http.Error(w, "archivo no encontrado", http.StatusNotFound)
		return
	}

	newPath := filepath.Join(s.uploadDir, newName)
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

func (s *server) convertAndSaveAsMP3(src io.Reader, sourceExt, dstPath string) error {
	tmpIn, err := os.CreateTemp("", "upload-*"+sourceExt)
	if err != nil {
		return fmt.Errorf("no se pudo preparar el archivo temporal: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if _, err := io.Copy(tmpIn, src); err != nil {
		tmpIn.Close()
		return fmt.Errorf("no se pudo guardar el archivo temporal: %w", err)
	}
	if err := tmpIn.Close(); err != nil {
		return fmt.Errorf("no se pudo cerrar el archivo temporal: %w", err)
	}

	tmpOut, err := os.CreateTemp(s.uploadDir, ".tmp-convert-*.mp3")
	if err != nil {
		return fmt.Errorf("no se pudo preparar el destino temporal: %w", err)
	}
	tmpOutPath := tmpOut.Name()
	tmpOut.Close()
	defer os.Remove(tmpOutPath)

	if err := convertToMP3(tmpIn.Name(), tmpOutPath); err != nil {
		return err
	}

	if err := os.Rename(tmpOutPath, dstPath); err != nil {
		return fmt.Errorf("no se pudo mover el mp3 convertido: %w", err)
	}

	if err := os.Chmod(dstPath, 0o644); err != nil {
		return fmt.Errorf("no se pudieron ajustar permisos: %w", err)
	}

	return nil
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

func ensureMP3Name(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if base == "" {
		base = name
	}
	if base == "" {
		base = "audio"
	}
	return base + ".mp3"
}

func convertToMP3(inputPath, outputPath string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-vn", "-codec:a", "libmp3lame", "-qscale:a", "2", outputPath)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg no pudo convertir el archivo: %v: %s", err, stderr.String())
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error al serializar respuesta: %v", err)
	}
}
