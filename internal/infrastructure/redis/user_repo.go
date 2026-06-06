package redis

import (
	"context"
	"encoding/json"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
	"github.com/trace/trace/internal/domain"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/pkg/crypto"
)

// UserRepository implements ports.UserProfileRepository.
// Transparently encrypts/decrypts profile data using AES-256.
// Health and injury data is sensitive — encryption happens at this layer,
// never inside agent or application code.
type UserRepository struct {
	client *Client
	cipher *crypto.AESCipher
}

func NewUserRepository(client *Client, cipher *crypto.AESCipher) *UserRepository {
	return &UserRepository{client: client, cipher: cipher}
}

func (r *UserRepository) Get(ctx context.Context, userID string) (*user.Profile, error) {
	key := ProfileKey(userID)
	ciphertext, err := r.client.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	plaintext, err := r.cipher.Decrypt(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("profile decrypt: %w", err)
	}
	var p user.Profile
	if err := json.Unmarshal(plaintext, &p); err != nil {
		return nil, fmt.Errorf("profile unmarshal: %w", err)
	}
	return &p, nil
}

func (r *UserRepository) Save(ctx context.Context, profile *user.Profile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("profile marshal: %w", err)
	}
	ciphertext, err := r.cipher.Encrypt(data)
	if err != nil {
		return fmt.Errorf("profile encrypt: %w", err)
	}
	return r.client.rdb.Set(ctx, ProfileKey(profile.UserID), ciphertext, 0).Err()
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	n, err := r.client.rdb.Exists(ctx, ProfileKey(userID)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
