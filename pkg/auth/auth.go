package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type DB interface {
	SaveToken(token string, guid string) error
	GetToken(guid string) (string, error)
	GetEmail(guid string) (string, error)
}

type EmailSender interface {
	WarningEmail(email string, message string) error
}

type JWTPayload struct {
	IpUser string
}

func GetTokenHandler(db DB, secretKey string) http.HandlerFunc {
	type Responce struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			if err != nil {
				log.Println("[HANDLER ERROR] Handler 'TokenHandler' ", err.Error())
			}
		}()

		data, ok := r.URL.Query()["guid"]
		if !ok || len(data) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		guid := data[0]

		linkByte := uuid.New().String()

		refreshExpiration := time.Now().Add(time.Hour * 24 * 30)
		rfClaims := Claims{
			Ip: r.RemoteAddr,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: refreshExpiration.Unix(),
				Id:        linkByte,
				Subject:   guid,
			},
		}

		refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, rfClaims)

		refreshToken, err := refresh.SignedString([]byte(secretKey))
		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		refreshHash, err := bcrypt.GenerateFromPassword([]byte(refreshToken)[len(refreshToken)-72:], bcrypt.DefaultCost)
		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := db.SaveToken(string(refreshHash), guid); err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		accessExpiration := time.Now().Add(time.Minute * 15)
		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
			Ip: r.RemoteAddr,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: accessExpiration.Unix(),
				Id:        linkByte,
				Subject:   guid,
			},
		})

		token, err := accessToken.SignedString([]byte(secretKey))
		if err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := Responce{
			AccessToken:  token,
			RefreshToken: refreshToken,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func RefreshHandler(db DB, email EmailSender, secretKey string) http.HandlerFunc {
	type Responce struct {
		AccessToken string `json:"accessToken"`
	}

	type Request struct {
		RefreshToken string `json:"refreshToken"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			if err != nil {
				log.Println("[HANDLER ERROR] Handler 'RefreshHandler' ", err.Error())
			}
		}()
		req := &Request{}
		if err = json.NewDecoder(r.Body).Decode(req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		refreshToken := req.RefreshToken
		accessToken := r.Header.Get("Authorization")

		if accessToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		accessClaims, err := ExtractClaims(accessToken, secretKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		refreshClaims, err := ExtractClaims(refreshToken, secretKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if accessClaims.Subject != refreshClaims.Subject || accessClaims.Id != refreshClaims.Id {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if refreshClaims.ExpiresAt < time.Now().Unix() {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := db.GetToken(refreshClaims.Subject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(token), []byte(refreshToken[len(refreshToken)-72:])); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if refreshClaims.Ip != r.RemoteAddr {
			userEmail, err := db.GetEmail(accessClaims.Subject)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if err := email.WarningEmail(userEmail, "IP address changed"); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		accessExpiration := time.Now().Add(time.Minute * 15)
		newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
			Ip: accessClaims.Ip,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: accessExpiration.Unix(),
				Id:        accessClaims.Id,
				Subject:   accessClaims.Subject,
			},
		})

		token, err = newAccessToken.SignedString([]byte(secretKey))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(Responce{AccessToken: token}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func ExtractClaims(jwtToken string, secretKey string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signature algorithm: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}
