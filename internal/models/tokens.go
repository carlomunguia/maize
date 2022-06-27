package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"log"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token represents a token used to authenticate a user.
type Token struct {
	PlainText  string    `json:"token"`
	UserID     int64     `json:"-"`
	Hash       []byte    `json:"-"`
	Expiration time.Time `json:"expiration"`
	Scope      string    `json:"-"`
}

// GenerateToken generates a new token for the given user.
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID:     int64(userID),
		Expiration: time.Now().Add(ttl),
		Scope:      scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `delete from tokens where user_id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	stmt = `
	INSERT INTO tokens
	(user_id, name, email, token_hash, expiration, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = m.DB.ExecContext(ctx, stmt,
		u.ID,
		u.LastName,
		u.Email,
		t.Hash,
		t.Expiration,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetUserByToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	stmt := `
	SELECT u.id, u.first_name, u.last_name, u.email
		FROM Users u
		INNER JOIN tokens t on (u.id = t.user_id)
	WHERE t.token_hash = ? and t.expiration > ?`

	err := m.DB.QueryRowContext(ctx, stmt, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user, nil
}
