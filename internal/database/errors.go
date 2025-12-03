package database

import "errors"

var (
	// ErrDatabaseURLNotSet is returned when DATABASE_URL is not found in environment.
	ErrDatabaseURLNotSet = errors.New("DATABASE_URL environment variable not set")

	// ErrUnsupportedDriver is returned when the database driver is not supported.
	ErrUnsupportedDriver = errors.New("unsupported database driver")

	// ErrEnvNotLoaded is returned when Connect is called before LoadEnv.
	ErrEnvNotLoaded = errors.New("environment not loaded: call LoadEnv first")

	// ErrAlreadyConnected is returned when Connect is called while already connected.
	ErrAlreadyConnected = errors.New("database already connected")

	// ErrNotConnected is returned when Close is called without an active connection.
	ErrNotConnected = errors.New("database not connected")
)
