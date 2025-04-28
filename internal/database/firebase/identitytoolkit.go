package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var apiKey = os.Getenv("WEB_API_KEY")

type signInResp struct {
	IDToken string `json:"idToken"`
	LocalID string `json:"localId"` // == uid
	Email   string `json:"email"`
}

func SignInWithEmail(ctx context.Context, email, password string) (signInResp, error) {
	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}
	b, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=%s", apiKey)
	res, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return signInResp{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return signInResp{}, fmt.Errorf("firebase signIn status=%d", res.StatusCode)
	}

	var out signInResp
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return signInResp{}, err
	}
	return out, nil
}
