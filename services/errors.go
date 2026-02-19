package services

import "errors"

var (
	ErrEmailAlreadyExists    = errors.New("ya existe una cuenta con este email")
	ErrUsernameAlreadyExists = errors.New("ya existe una cuenta con este nombre de usuario")
	ErrInvalidCredentials    = errors.New("email o contraseña incorrectos")
	ErrUserNotFound          = errors.New("el usuario no existe")
	ErrWrongPassword         = errors.New("la contraseña actual es incorrecta")
	ErrInvalidToken          = errors.New("token inválido o expirado")
)
