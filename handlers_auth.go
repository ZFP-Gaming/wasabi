package main

import (
	"fmt"
	"log"
	"net/http"
)

func (s *server) authDiscordHandler(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, s.auth.authURL(state), http.StatusTemporaryRedirect)
}

func (s *server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		http.Redirect(w, r, fmt.Sprintf("%s?error=access_denied", s.frontendOrigin), http.StatusTemporaryRedirect)
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

	accessToken, err := s.auth.exchangeCodeForToken(code)
	if err != nil {
		log.Printf("error al intercambiar código: %v", err)
		http.Error(w, "error al autenticar con Discord", http.StatusUnauthorized)
		return
	}

	user, err := s.auth.fetchDiscordUser(accessToken)
	if err != nil {
		log.Printf("error al obtener usuario: %v", err)
		http.Error(w, "error al obtener perfil de usuario", http.StatusInternalServerError)
		return
	}

	isMember, err := s.auth.isMemberOfRequiredGuild(accessToken)
	if err != nil {
		log.Printf("error al verificar membresía: %v", err)
		http.Error(w, "no se pudo verificar membresía del servidor", http.StatusUnauthorized)
		return
	}
	if !isMember {
		http.Error(w, "debes ser miembro del servidor de Discord para usar Wasabi", http.StatusForbidden)
		return
	}

	jwtToken, err := s.auth.generateJWT(user)
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

	http.Redirect(w, r, s.frontendOrigin, http.StatusTemporaryRedirect)
}

func (s *server) logoutHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *server) meHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "solo se permite GET", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := getUserClaims(r.Context())
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
