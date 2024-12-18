package auth_test

import (
	"app/pkg/auth"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type MockDB struct {
	guid  string
	token string
	email string
}

func (db *MockDB) GetToken(guid string) (string, error) {
	if guid == db.guid {
		return db.token, nil
	} else {
		return "", sql.ErrNoRows
	}
}

func (t *MockDB) SaveToken(token string, guid string) error {
	if guid == t.guid {
		t.token = token
		return nil
	} else {
		return sql.ErrNoRows
	}
}

func (t *MockDB) GetEmail(guid string) (string, error) {
	if guid == t.guid {
		return t.email, nil
	} else {
		return "", sql.ErrNoRows
	}
}

func TestAuthHandlers(t *testing.T) {
	db := &MockDB{
		guid:  "a41f0a51-3015-4ce8-b3c2-38bb684f1f00",
		token: "",
		email: "example@gmail.com",
	}

	tokenHandler := auth.GetTokenHandler(db, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	q := url.Values{}
	q.Add("guid", "a41f0a51-3015-4ce8-b3c2-38bb684f1f00")
	req.URL.RawQuery = q.Encode()

	tokenHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	body := struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{}

	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if body.AccessToken == "" || body.RefreshToken == "" {
		t.Errorf("Expected access token, got empty")
	}

	w = httptest.NewRecorder()
	rBody := bytes.Buffer{}
	err := json.NewEncoder(&rBody).Encode(body)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	req = httptest.NewRequest("GET", "/", &rBody)
	req.Header.Set("Authorization", body.AccessToken)

	refreshHandler := auth.RefreshHandler(db, nil, "secret")

	refreshHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d | body: %s", http.StatusOK, w.Code, body.RefreshToken)
	}

	if w.Body.String() == "" {
		t.Errorf("Expected refresh token, got %s", w.Body.String())
	}
}
