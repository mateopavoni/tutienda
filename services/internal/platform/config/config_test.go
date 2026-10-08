package config

import "testing"

func TestJWTSecretUnsafe(t *testing.T) {
	cases := []struct {
		name string
		c    Common
		want bool
	}{
		{"dev default in dev", Common{Env: "dev", JWTSecret: DevJWTSecret}, false},
		{"dev default in prod", Common{Env: "prod", JWTSecret: DevJWTSecret}, true},
		{"env.example placeholder in prod", Common{Env: "prod", JWTSecret: "change-me-to-a-long-random-secret"}, true},
		{"real secret in prod", Common{Env: "prod", JWTSecret: "k3J9xQ2mVw8LpT5aZr1Yb7NcHd4Fe6Sg"}, false},
	}
	for _, tc := range cases {
		if got := tc.c.jwtSecretUnsafe(); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
