package server

type HTTPClient struct {
	conn   *common.ClientConn
}

func NewHTTPClient(conn net.Conn) (*HTTPClient, error) {
	c := &HTTPClient{conn: common.NewClientConn(conn)}
	return c, nil
}
