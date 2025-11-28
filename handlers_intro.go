package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type introRequest struct {
	SoundName string `json:"soundName"`
}

func (s *server) introHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "solo se permite POST", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := getUserClaims(r.Context())
	if !ok {
		http.Error(w, "no se pudo obtener usuario", http.StatusInternalServerError)
		return
	}

	var payload introRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "cuerpo JSON inválido", http.StatusBadRequest)
		return
	}

	soundName, err := sanitizeName(payload.SoundName)
	if err != nil || soundName == "" {
		http.Error(w, "nombre de sonido requerido", http.StatusBadRequest)
		return
	}

	path := filepath.Join(s.uploadDir, soundName)
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		http.Error(w, "el sonido no existe en uploads", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error al validar sonido: %v", err)
		http.Error(w, "no se pudo procesar la solicitud", http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.Error(w, "nombre de sonido inválido", http.StatusBadRequest)
		return
	}

	log.Printf("solicitud de intro recibida: user_id=%s sound=%s", claims.UserID, soundName)
	writeJSON(w, http.StatusOK, map[string]string{
		"message":   "solicitud de intro registrada",
		"soundName": soundName,
	})
}
