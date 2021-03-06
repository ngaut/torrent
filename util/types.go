package util

import (
	"encoding"
	"encoding/binary"
	"errors"
	"net"

	"github.com/bradfitz/iter"

	"github.com/anacrolix/torrent/bencode"
)

// Concatenated 6-byte peer addresses.
type CompactIPv4Peers []CompactPeer

var (
	// This allows bencode.Unmarshal to do better than a string or []byte.
	_ bencode.Unmarshaler      = &CompactIPv4Peers{}
	_ encoding.BinaryMarshaler = CompactIPv4Peers{}
)

// This allows bencode.Unmarshal to do better than a string or []byte.
func (me *CompactIPv4Peers) UnmarshalBencode(b []byte) (err error) {
	var bb []byte
	err = bencode.Unmarshal(b, &bb)
	if err != nil {
		return
	}
	*me, err = UnmarshalIPv4CompactPeers(bb)
	return
}

func (me CompactIPv4Peers) MarshalBinary() (ret []byte, err error) {
	ret = make([]byte, len(me)*6)
	for i, cp := range me {
		copy(ret[6*i:], cp.IP.To4())
		binary.BigEndian.PutUint16(ret[6*i+4:], uint16(cp.Port))
	}
	return
}

// Represents peer address in either IPv6 or IPv4 form.
type CompactPeer struct {
	IP   net.IP
	Port int
}

func (me *CompactPeer) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case 18:
		me.IP = make([]byte, 16)
	case 6:
		me.IP = make([]byte, 4)
	default:
		return errors.New("bad length")
	}
	copy(me.IP, b)
	b = b[len(me.IP):]
	me.Port = int(binary.BigEndian.Uint16(b))
	return nil
}

func UnmarshalIPv4CompactPeers(b []byte) (ret []CompactPeer, err error) {
	if len(b)%6 != 0 {
		err = errors.New("bad length")
		return
	}
	num := len(b) / 6
	ret = make([]CompactPeer, num)
	for i := range iter.N(num) {
		off := i * 6
		err = ret[i].UnmarshalBinary(b[off : off+6])
		if err != nil {
			return
		}
	}
	return
}
