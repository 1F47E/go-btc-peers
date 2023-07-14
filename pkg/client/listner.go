package client

import (
	"fmt"
	"io"

	"github.com/btcsuite/btcd/wire"
)

func (n *Node) connListen() {
	a := fmt.Sprintf("︎ ◀︎ %s:", n.Endpoint())
	defer func() {
		n.Conn = nil
		log.Warnf("%s closed\n", a)
	}()
	// buf := make([]byte, 65536)
	// bufReader := bufio.NewReader(n.Conn)
	for {
		if n.Conn == nil {
			return
		}
		fmt.Println()
		// cnt, err := bufReader.Read(buf)
		// if err != nil {
		// 	log.Fatalf("failed to read from conn: %v", err)
		// 	return
		// }
		// log.Debugf("[listner]: raw bytes read: %v\n", cnt)
		// data := buf[:cnt]
		// log.Debugf("[listner]: raw buf: %v\n", data)
		// log.Debugf("[listner]: raw buf: %v\n", string(data))
		cnt, msg, rawPayload, err := wire.ReadMessageN(n.Conn, cfg.Pver, cfg.Btcnet)
		// cnt, msg, rawPayload, err := wire.ReadMessageWithEncodingN(n.Conn, cfg.Pver, cfg.Btcnet, wire.BaseEncoding)
		if err != nil {
			if err == io.EOF {
				log.Warnf("%s EOF, exit\n", a)
				return
			}
			// Since the protocol version is 70016 but we don't
			// implement compact blocks, we have to ignore unknown
			// messages after the version-verack handshake. This
			// matches bitcoind's behavior and is necessary since
			// compact blocks negotiation occurs after the
			// handshake.
			if err == wire.ErrUnknownMessage {
				log.Warnf("%s ERR: unknown message, ignoring\n", a)
				continue
			}

			// log.Fatalf("Cant read buffer, error: %v\n", err)
			log.Warnf("%s ERR: Cant read buffer, error: %v\n", a, err)
			log.Warnf("%s ERR: bytes read: %v\n", a, cnt)
			log.Warnf("%s ERR: msg: %v\n", a, msg)
			log.Warnf("%s ERR: rawPayload: %v\n", a, rawPayload)
			continue
		}
		log.Debugf("%s Got message: %d bytes, cmd: %s rawPayload len: %d\n", a, cnt, msg.Command(), len(rawPayload))
		switch m := msg.(type) {
		case *wire.MsgVersion:
			log.Infof("%s MsgVersion received\n", a)
			log.Debugf("%s version: %v\n", a, m.ProtocolVersion)
			log.Debugf("%s msg: %+v\n", a, m)
		case *wire.MsgVerAck:
			log.Infof("%s MsgVerAck received\n", a)
			log.Debugf("%s msg: %+v\n", a, m)
		case *wire.MsgPing:
			log.Infof("%s MsgPing received\n", a)
			log.Debugf("%s nonce: %v\n", a, m.Nonce)
			log.Debugf("%s msg: %+v\n", a, m)
		case *wire.MsgPong:
			log.Infof("%s MsgPong received\n", a)
			if m.Nonce == n.PingNonce {
				log.Debugf("%s pong OK\n", a)
				n.PingCount++
				n.PingNonce = 0
			} else {
				log.Warnf("%s pong nonce mismatch, expected %v, got %v\n", a, n.PingNonce, m.Nonce)
			}
		case *wire.MsgAddr:
			log.Infof("%s MsgAddr received\n", a)
			log.Debugf("%s got %d addresses\n", a, len(m.AddrList))
			batch := make([]string, len(m.AddrList))
			for i, a := range m.AddrList {
				batch[i] = fmt.Sprintf("[%s]:%d", a.IP.String(), a.Port)
			}
			newNodesCh <- batch
		case *wire.MsgAddrV2:
			log.Infof("%s MsgAddrV2 received\n", a)
			log.Debugf("%s got %d addresses\n", a, len(m.AddrList))
			batch := make([]string, len(m.AddrList))
			for i, a := range m.AddrList {
				batch[i] = fmt.Sprintf("[%s]:%d", a.Addr.String(), a.Port)
			}
			newNodesCh <- batch

		case *wire.MsgInv:
			log.Infof("%s MsgInv received\n", a)
			log.Debugf("%s data: %d\n", a, len(m.InvList))
			// TODO: answer on inv

		case *wire.MsgFeeFilter:
			log.Infof("%s MsgFeeFilter received\n", a)
			log.Debugf("%s fee: %v\n", a, m.MinFee)
		case *wire.MsgGetHeaders:
			log.Infof("%s MsgGetHeaders received\n", a)
			log.Debugf("%s headers: %d\n", a, len(m.BlockLocatorHashes))

		default:
			log.Infof("%s unknown message received\n", a)
			log.Debugf("%s msg: %+v\n", a, m)
		}
	}
}
