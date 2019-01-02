package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kubeplay/gameserver/pkg/types"
)

type HandlerFn func(w http.ResponseWriter, r *http.Request)

var (
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	dbConfig  = struct {
		file   string
		bucket string
	}{
		"/tmp/kubeplay.db",
		"/registry/v1",
	}
)

func NewResponse(w http.ResponseWriter) *HttpResponse {
	return &HttpResponse{
		statusCode: 200,
		response:   w,
	}
}

type HttpResponse struct {
	statusCode int
	response   http.ResponseWriter
}

func (r *HttpResponse) Status(statusCode int) *HttpResponse {
	r.statusCode = statusCode
	return r
}

func (r *HttpResponse) WriteRawJSON(rawObj []byte) error {
	r.response.Header().Set("Content-Type", "application/json")
	r.response.WriteHeader(r.statusCode)
	_, err := r.response.Write(rawObj)
	return err
}

func (r *HttpResponse) WriteJSON(obj types.Object) error {
	r.response.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	r.response.WriteHeader(r.statusCode)
	_, err = r.response.Write(append(data, '\n'))
	return err
}

// GenerateNewJwtToken creates a new user token to allow machine-to-machine interaction
func GenerateNewJwtToken(key []byte, p *types.PlayerClaims, exp time.Time) error {
	token := jwt.New(jwt.SigningMethodHS256)
	p.ExpiresAt = exp.UTC().Unix()
	p.IssuedAt = time.Now().UTC().Unix()
	token.Claims = p
	// Sign and get the complete encoded token as a string
	var err error
	p.AccessToken, err = token.SignedString(key)
	return err
}

// DecodeUserToken decodes a jwtToken (HS256 and RS256)
func DecodeUserToken(jwtTokenString, jwtSecret string) (*types.PlayerClaims, error) {
	player := &types.PlayerClaims{}
	token, err := jwt.ParseWithClaims(jwtTokenString, player, func(token *jwt.Token) (interface{}, error) {
		switch t := token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			return []byte(jwtSecret), nil
		default:
			return nil, fmt.Errorf("unknown sign method [%v]", t)
		}
	})
	if err == nil && token.Valid {
		return player, nil
	}
	switch t := err.(type) {
	case *jwt.ValidationError:
		if t.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, fmt.Errorf("it's not a valid token")
		} else if t.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return nil, fmt.Errorf("the token is expired or not valid yet")
		} else {
			return nil, fmt.Errorf("failed decoding token [%v]", err)
		}
	default:
		return nil, fmt.Errorf("unknown error, failed decoding token [%v]", err)
	}
}
