package common

// ClientHeader is the header sent to the server at the beginning of the connection
type ClientHeader struct {
	FormatVersion int    `json:"format_version"`
	Protocol      string `json:"protocol"`
	HTTPHost      string `json:"http_host"`
}
