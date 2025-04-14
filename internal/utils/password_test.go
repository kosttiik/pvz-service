package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "secretpassword123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: strings.Repeat("a", 72),
			wantErr:  false,
		},
		{
			name:     "Very long password",
			password: strings.Repeat("a", 100),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash for valid password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "password123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password for test: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "Invalid password",
			password: "wrongpassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalid_hash",
			wantErr:  true,
		},
		{
			name:     "Very long password",
			password: strings.Repeat("a", 100),
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "Empty hash",
			password: "password123",
			hash:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid hash format",
			password: "password123",
			hash:     "not-a-bcrypt-hash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPasswordHashingAndChecking(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Normal password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Complex password",
			password: "P@ssw0rd!123$",
			wantErr:  false,
		},
		{
			name:     "Unicode password",
			password: "пароль123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if hash == tt.password {
					t.Error("HashPassword() returned plain password")
				}

				if err := CheckPassword(tt.password, hash); err != nil {
					t.Errorf("CheckPassword() error = %v", err)
				}

				if err := CheckPassword("wrong"+tt.password, hash); err == nil {
					t.Error("CheckPassword() should fail with wrong password")
				}
			}
		})
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Invalid hash format",
			password: "password123",
			hash:     "invalidhash",
			wantErr:  true,
		},
		{
			name:     "Empty hash",
			password: "password123",
			hash:     "",
			wantErr:  true,
		},
		{
			name:     "Very long hash",
			password: "password123",
			hash:     "$2a$10$" + strings.Repeat("x", 100),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
