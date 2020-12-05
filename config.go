package main

const (
	// ShoutDownTimeoutS waiting before timeput
	ShoutDownTimeoutS int = 20
	// ReadTimeoutS waiting for read
	ReadTimeoutS = 20
	// WriteTimeoutS wait for write
	WriteTimeoutS = 20
	// MaxIdleConc is max number of idle connections
	MaxIdleConn = 10
	// IdleConnTimeout is max time to keep idle conn alive
	IdleConnTimeout = 30
)
