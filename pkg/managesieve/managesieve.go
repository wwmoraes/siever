package managesieve

import (
	"io"
)

const (
	Authenticate   WireCommand = `AUTHENTICATE`
	StartTLS       WireCommand = `STARTTLS`
	Logout         WireCommand = `LOGOUT`
	Capability     WireCommand = `CAPABILITY`
	HaveSpace      WireCommand = `HAVESPACE`
	PutScript      WireCommand = `PUTSCRIPT`
	ListScripts    WireCommand = `LISTSCRIPTS`
	SetActive      WireCommand = `SETACTIVE`
	GetScript      WireCommand = `GETSCRIPT`
	DeleteScript   WireCommand = `DELETESCRIPT`
	RenameScript   WireCommand = `RENAMESCRIPT`
	CheckScript    WireCommand = `CHECKSCRIPT`
	NoOp           WireCommand = `NOOP`
	Unauthenticate WireCommand = `UNAUTHENTICATE`
)

type WireCommand = string

// Logger represents the logger functionalities required by the client
type Logger interface {
	SetOutput(w io.Writer)
	SetPrefix(string)

	Printf(string, ...any)
	Println(...any)
}

// TODO create file abstraction to handle scripts on storage level

type Script struct {
	Name   string
	Active bool
}
