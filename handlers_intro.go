package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	effect := strings.TrimSuffix(soundName, filepath.Ext(soundName))
	if effect == "" {
		effect = soundName
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	_, err = s.introsCollection.UpdateOne(
		ctx,
		bson.M{"id": claims.UserID},
		bson.M{"$set": bson.M{"effect": effect}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Printf("error al guardar intro en mongo: %v", err)
		http.Error(w, "no se pudo guardar la intro", http.StatusInternalServerError)
		return
	}

	log.Printf("solicitud de intro registrada: user_id=%s sound=%s effect=%s", claims.UserID, soundName, effect)
	writeJSON(w, http.StatusOK, map[string]string{
		"message":   "solicitud de intro registrada",
		"soundName": soundName,
		"effect":    effect,
	})
}
