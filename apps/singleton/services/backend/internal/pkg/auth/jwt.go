package auth

import jwtv5 "github.com/golang-jwt/jwt/v5"

// SigningMethod is the JWT signing method for both token generation and verification.
var SigningMethod = jwtv5.SigningMethodHS256
