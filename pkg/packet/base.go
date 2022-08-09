package packet

/**
  @author: baisongyuan
  @since: 2022/7/25
**/

const (
	REQUEST_FAIL = iota + 10000
)

// client packet code

const (
	Handshake = iota + 20000
	RequestConfigPacket
)

const (
	HandshakeSuccess = iota + 30000
	ConfigChangePacket
	ResponseConfig
)
