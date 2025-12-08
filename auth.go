package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userContextKey contextKey = "user"

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
	GuildID       string `json:"guild_id"`
	jwt.RegisteredClaims
}

type authService struct {
	config          oauth2Config
	jwtSecret       []byte
	requiredGuildID string
}

func newAuthService(cfg authConfig) *authService {
	return &authService{
		config: oauth2Config{
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			redirectURI:  cfg.RedirectURI,
		},
		jwtSecret:       []byte(cfg.JWTSecret),
		requiredGuildID: cfg.RequiredGuildID,
	}
}

func (a *authService) authURL(state string) string {
	return fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		url.QueryEscape(a.config.clientID),
		url.QueryEscape(a.config.redirectURI),
		url.QueryEscape("identify guilds"),
		url.QueryEscape(state),
	)
}

func (a *authService) exchangeCodeForToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", a.config.clientID)
	data.Set("client_secret", a.config.clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", a.config.redirectURI)

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

func (a *authService) fetchDiscordUser(accessToken string) (*discordUser, error) {
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

func (a *authService) isMemberOfRequiredGuild(accessToken string) (bool, error) {
	if a.requiredGuildID == "" {
		return true, nil
	}

	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me/guilds", nil)
	if err != nil {
		return false, fmt.Errorf("error al crear petición: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error en petición: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("código de estado: %d", resp.StatusCode)
	}

	var guilds []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
		return false, fmt.Errorf("error al decodificar guilds: %w", err)
	}

	for _, guild := range guilds {
		if guild.ID == a.requiredGuildID {
			return true, nil
		}
	}

	return false, nil
}

func (a *authService) generateJWT(user *discordUser) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:        user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		Avatar:        user.Avatar,
		GuildID:       a.requiredGuildID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *authService) validateJWT(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("token inválido")
}

func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("error al generar estado aleatorio: %v", err)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)
}

func contextWithUser(ctx context.Context, claims *jwtClaims) context.Context {
	return context.WithValue(ctx, userContextKey, claims)
}

func getUserClaims(ctx context.Context) (*jwtClaims, bool) {
	claims, ok := ctx.Value(userContextKey).(*jwtClaims)
	return claims, ok
}
