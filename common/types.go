package common

// ClientHeader is the header sent to the server at the beginning of the connection
type ClientHeader struct {
	FormatVersion int
	Protocol      string
	HTTPHost      string
}
