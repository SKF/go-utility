package client

type Client interface {
	PostLogEntry(request interface{}) (err error)
}
