package client

import (
	"github.com/sirupsen/logrus"
)

type type123pkg byte

const (
	HELLO type123pkg = iota + 1
	PUNCH
	OLLEH
	CHAIN
	INIT_REQUEST
	INIT_SUCCESS
	INIT_ERROR
	REDIRECT
	QUIT      = iota + 3
	HOST_GAME = iota + 4
	CLIENT_GAME
	SOKUROLL_TIME = iota + 5
	SOKUROLL_TIME_ACK
	SOKUROLL_STATE
	SOKUROLL_SETTINGS
	SOKUROLL_SETTINGS_ACK
)

type status123peer byte

const (
	INACTIVE status123peer = iota
	SUCCESS
)

type spectate123type byte

const (
	NOSPECTATE             spectate123type = 0x00
	SPECTATE               spectate123type = 0x10
	SPECTATE_FOR_SPECTATOR spectate123type = 0x11
)

type hisoutensokuData struct {
	PeerAddr    [6]byte  // appear in HELLO
	TargetAddr  [6]byte  // appear in HELLO
	GameID      [16]byte // appear in INIT_REQUEST
	ClientProf  string   // appear in INIT_REQUEST and INIT_SUCCESS
	HostProf    string   // appear in INIT_SUCCESS
	Spectator   bool     // appear in INIT_SUCCESS
	SwrDisabled bool     // appear in INIT_SUCCESS
}

type Hisoutensoku struct {
	peerId     byte
	peerStatus status123peer

	peerData map[byte]hisoutensokuData // data record with client id
}

var logger123 = logrus.WithField("Hisoutensoku", "internal")

// NewHisoutensoku new Hisoutensoku spectating server
func NewHisoutensoku() *Hisoutensoku {
	return &Hisoutensoku{
		peerStatus: INACTIVE,
		peerData:   make(map[byte]hisoutensokuData),
	}
}

// WriteFunc from game host to client
// orig: original data leads with 1 byte of client id
func (h *Hisoutensoku) WriteFunc(orig []byte) (bool, []byte) {

	switch type123pkg(orig[1]) {
	case INIT_SUCCESS:
		if len(orig)-1 == 81 {
			var v hisoutensokuData
			var ok bool
			var hprof, cprof string
			var swrDisabled int

			logger123.Debug("INIT_SUCCESS with spectacle type ", orig[7])
			if v, ok = h.peerData[orig[0]]; !ok {
				logger123.Warn("INIT_SUCCESS without HELLO ahead?")
				v = hisoutensokuData{}
			}

			switch spectate123type(orig[7]) {
			case NOSPECTATE:
				v.Spectator = false
			case SPECTATE:
				v.Spectator = true
			case SPECTATE_FOR_SPECTATOR:
				logger123.Warn("INIT_SUCCESS type SPECTATE_FOR_SPECTATOR appears here?")
			default:
				logger123.Warn("INIT_SUCCESS spectacle type cannot recognize")
			}

			for i := 14; i <= 46; i++ {
				if orig[i] == 0x00 || i == 46 {
					hprof = string(orig[14:i])
					break
				}
			}
			for i := 46; i <= 78; i++ {
				if orig[i] == 0x00 || i == 78 {
					cprof = string(orig[46:i])
					break
				}
			}
			swrDisabled = int(orig[78]) | int(orig[79])<<8 | int(80)<<16 | int(81)<<24
			logger123.Debug("INIT_SUCCESS with host profile ", hprof, " client profile ", cprof, " swr disabled ", swrDisabled)

			v.HostProf, v.ClientProf, v.SwrDisabled = hprof, cprof, swrDisabled != 0

			h.peerData[orig[0]] = v

			h.peerId = orig[0]
			h.peerStatus = SUCCESS

		} else {
			logger123.Debug("INIT_SUCCESS with strange length ", len(orig)-1)
		}

	case QUIT:
		if len(orig)-1 == 1 {
			logger123.Debug("QUIT")
			if orig[0] == h.peerId {
				h.peerStatus = INACTIVE
			}
		} else {
			logger123.Debug("QUIT with strange length ", len(orig)-1)
		}

	case CLIENT_GAME:
		logger123.Warn("CLIENT_GAME should not appear here right? ", orig[1:])

	case HOST_GAME:

	}

	return false, orig
}

// ReadFunc from game client to host
// orig: original data leads with 1 byte of client id
func (h *Hisoutensoku) ReadFunc(orig []byte) (bool, []byte) {

	switch type123pkg(orig[1]) {
	case HELLO:
		if len(orig)-1 == 37 {
			if v, ok := h.peerData[orig[0]]; !ok {
				logger123.Debug("HELLO with data ", orig[4:10], orig[20:26])
				v = hisoutensokuData{}
				copy(v.PeerAddr[:], orig[4:10])
				copy(v.TargetAddr[:], orig[20:26])
				h.peerData[orig[0]] = v
			} else {
				copy(v.PeerAddr[:], orig[4:10])
				copy(v.TargetAddr[:], orig[20:26])
				h.peerData[orig[0]] = v
			}
		} else {
			logger123.Debug("HELLO with strange length ", len(orig)-1)
		}

	case INIT_REQUEST:
		if len(orig)-1 == 65 {
			var prof string
			var ok bool
			var v hisoutensokuData

			logger123.Debug("INIT_REQUEST with game id ", orig[2:18], " request type is ", orig[26])
			if v, ok = h.peerData[orig[0]]; !ok {
				logger123.Warn("INIT_REQUEST without HELLO ahead?")
				v = hisoutensokuData{}
			}

			if orig[26] == 0x01 {
				if orig[27] > 32 {
					logger123.Warn("INIT_REQUEST profile name too long!")
				} else {
					prof = string(orig[28 : 28+orig[27]])
					logger123.Debug("INIT_REQUEST client profile name is ", prof)
				}
			}

			copy(v.GameID[:], orig[2:18])
			if len(prof) > 0 {
				v.ClientProf = prof
			}
			h.peerData[orig[0]] = v
		} else {
			logger123.Debug("INIT_REQUEST with strange length ", len(orig)-1)
		}

	case QUIT:
		if len(orig)-1 == 1 {
			logger123.Debug("QUIT")
			if orig[0] == h.peerId {
				h.peerStatus = INACTIVE
			}
		} else {
			logger123.Debug("QUIT with strange length ", len(orig)-1)
		}

	case HOST_GAME:
		logger123.Warn("HOST_GAME should not appear here right? ", orig[1:])

	case CLIENT_GAME:

	}

	return false, orig
}
