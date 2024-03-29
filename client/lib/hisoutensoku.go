package client

import (
	"bytes"
	"compress/zlib"
	"io"
	"math"
	"net"
	"time"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

type type123pkg byte

const (
	HELLO_123             type123pkg = iota + 1 // 0x01
	PUNCH_123                                   // 0x02
	OLLEH_123                                   // 0x03
	CHAIN_123                                   // 0x04
	INIT_REQUEST_123                            // 0x05
	INIT_SUCCESS_123                            // 0x06
	INIT_ERROR_123                              // 0x07
	REDIRECT_123                                // 0x08
	QUIT_123              type123pkg = iota + 3 // 0x0b
	HOST_GAME_123         type123pkg = iota + 4 // 0x0d
	CLIENT_GAME_123                             // 0x0e
	SOKUROLL_TIME         type123pkg = iota + 5 // 0x10
	SOKUROLL_TIME_ACK                           // 0x11
	SOKUROLL_STATE                              // 0x12
	SOKUROLL_SETTINGS                           // 0x13
	SOKUROLL_SETTINGS_ACK                       // 0x14
)

type data123pkg byte

const (
	GAME_LOADED_123         data123pkg = iota + 1 // 0x01
	GAME_LOADED_ACK_123                           // 0x02
	GAME_INPUT_123                                // 0x03
	GAME_MATCH_123                                // 0x04
	GAME_MATCH_ACK_123                            // 0x05
	GAME_MATCH_REQUEST_123  data123pkg = iota + 3 // 0x08
	GAME_REPLAY_123                               // 0x09
	GAME_REPLAY_REQUEST_123 data123pkg = iota + 4 // 0x0b
)

type Status123peer byte

const (
	INACTIVE_123 Status123peer = iota
	SUCCESS_123
	BATTLE_123
	BATTLE_WAIT_ANOTHER_123
)

type spectate123type byte

const (
	NOSPECTATE_123             spectate123type = 0x00
	SPECTATE_123               spectate123type = 0x10
	SPECTATE_FOR_SPECTATOR_123 spectate123type = 0x11
)

type hisoutensokuData struct {
	// PeerAddr    [6]byte  // appear in HELLO
	// TargetAddr  [6]byte  // appear in HELLO
	// GameID      [16]byte // appear in INIT_REQUEST
	InitSuccessPkg [81]byte // copy from INIT_SUCCESS
	ClientProf     string   // parse from INIT_SUCCESS
	HostProf       string   // parse from INIT_SUCCESS
	Spectator      bool     // parse from INIT_SUCCESS
	SwrDisabled    bool     // parse from INIT_SUCCESS

	HostInfo    [45]byte          // parse from HOST_GAME GAME_MATCH
	ClientInfo  [45]byte          // parse from HOST_GAME GAME_MATCH
	StageId     byte              // parse from HOST_GAME GAME_MATCH
	MusicId     byte              // parse from HOST_GAME GAME_MATCH
	RandomSeeds [4]byte           // parse from HOST_GAME GAME_MATCH
	MatchId     byte              // parse from HOST_GAME GAME_MATCH
	ReplayData  map[byte][]uint16 // parse from HOST_GAME GAME_REPLAY
	ReplayEnd   map[byte]bool     // parse from HOST_GAME GAME_REPLAY
}

func newHisoutensokuData() *hisoutensokuData {
	return &hisoutensokuData{
		ReplayData: make(map[byte][]uint16),
		ReplayEnd:  make(map[byte]bool),
	}
}

type status123req byte

const (
	INIT_123 status123req = iota
	SEND_123
	SENT0_123
	SENT1_123
	SEND_AGAIN_123
)

type Hisoutensoku struct {
	peerId         byte              // current host/client peer id (udp mutex id)
	PeerStatus     Status123peer     // current peer status
	peerData       *hisoutensokuData // current peer data record
	gameId         map[byte][16]byte // game id record
	repReqStatus   status123req      // GAME_REPLAY_REQUEST send status
	repReqTime     time.Time         // request send time
	repReqDelay    time.Duration     // delay between GAME_REPLAY_REQUEST and GAME_REPLAY package
	spectatorCount int               // spectator counter

	quitFlag bool // plugin quit flag
}

var logger123 = logrus.WithField("Hisoutensoku", "internal")

// NewHisoutensoku new Hisoutensoku spectating server
func NewHisoutensoku() *Hisoutensoku {
	return &Hisoutensoku{
		PeerStatus:     INACTIVE_123,
		peerData:       newHisoutensokuData(),
		gameId:         make(map[byte][16]byte),
		repReqStatus:   INIT_123,
		repReqDelay:    time.Second,
		spectatorCount: 0,
		quitFlag:       false,
	}
}

// WriteFunc from game host to client
// orig: original data leads with 1 byte of client id
func (h *Hisoutensoku) WriteFunc(orig []byte) (bool, []byte) {

	switch type123pkg(orig[1]) {
	case INIT_SUCCESS_123:
		if len(orig)-1 == 81 {

			switch spectate123type(orig[6]) {
			case NOSPECTATE_123, SPECTATE_123:
				// init success
				h.peerData.Spectator = spectate123type(orig[6]) == SPECTATE_123
				for i := 14; i <= 46; i++ {
					if orig[i] == 0x00 || i == 46 {
						h.peerData.HostProf = string(orig[14:i])
						break
					}
				}
				for i := 46; i <= 78; i++ {
					if orig[i] == 0x00 || i == 78 {
						h.peerData.ClientProf = string(orig[46:i])
						break
					}
				}
				h.peerData.SwrDisabled = (int(orig[78]) | int(orig[79])<<8 | int(80)<<16 | int(81)<<24) != 0

				logger123.Debug("INIT_SUCCESS with host profile ", h.peerData.HostProf, " client profile ",
					h.peerData.ClientProf, " swr disabled ", h.peerData.SwrDisabled)

				h.peerId = orig[0]
				h.PeerStatus = SUCCESS_123
				h.peerData.MatchId = 0

				logger123.Info("Th123 peer init success: spectator=", h.peerData.Spectator)

			case SPECTATE_FOR_SPECTATOR_123:
				logger123.Warn("INIT_SUCCESS type SPECTATE_FOR_SPECTATOR appears here?")
			default:
				logger123.Warn("INIT_SUCCESS spectacle type cannot recognize")
			}

		} else {
			logger123.Warn("INIT_SUCCESS with strange length ", len(orig)-1)
		}

	case QUIT_123:
		if len(orig)-1 == 1 {
			logger123.Info("Th123 peer quit")
			if orig[0] == h.peerId {
				h.PeerStatus = INACTIVE_123
			}
		} else {
			logger123.Warn("QUIT with strange length ", len(orig)-1)
		}

	case CLIENT_GAME_123:
		logger123.Warn("CLIENT_GAME should not appear here right? ", orig[1:])

	case HOST_GAME_123:
		switch data123pkg(orig[2]) {
		case GAME_LOADED_ACK_123:
			if orig[3] == 0x05 {
				logger123.Info("Th123 battle loaded")
				h.PeerStatus = BATTLE_123
			}
		}
	}

	return false, orig
}

// ReadFunc from game client to host
// orig: original data leads with 1 byte of client id
func (h *Hisoutensoku) ReadFunc(orig []byte) (bool, []byte) {

	switch type123pkg(orig[1]) {
	case HELLO_123:
		if len(orig)-1 == 37 {
			if h.PeerStatus > SUCCESS_123 && orig[0] != h.peerId {
				return true, []byte{orig[0], byte(OLLEH_123)}
			}
		} else {
			logger123.Warn("HELLO with strange length ", len(orig)-1)
		}

	case CHAIN_123:
		if len(orig)-1 == 5 {
			if h.PeerStatus > SUCCESS_123 && orig[0] != h.peerId {
				return true, []byte{orig[0], 4, 4, 0, 0, 0}
			}
		} else {
			logger123.Warn("CHAIN with strange length ", len(orig)-1)
		}

	case INIT_REQUEST_123:
		if len(orig)-1 == 65 {

			var gameId [16]byte

			copy(gameId[:], orig[2:18])
			logger123.Debug("INIT_REQUEST with game id ", gameId, " request type is ", orig[26])

			h.gameId[orig[0]] = gameId

			if h.PeerStatus > SUCCESS_123 && orig[0] != h.peerId {
				// from spectator
				if (orig[26] == 0x00 && h.peerData.Spectator && h.peerData.MatchId == 0) || orig[26] == 0x01 {
					// replay request but not start or match request
					logger123.Debug("INIT_REQUEST for game from spectator")
					return true, []byte{orig[0], byte(INIT_ERROR_123), 1, 0, 0, 0}
				} else if orig[26] == 0x00 && !h.peerData.Spectator {
					// not allow spectator
					logger123.Info("Th123 not allow spectator")
					return true, []byte{orig[0], byte(INIT_ERROR_123), 0, 0, 0, 0}
				} else if orig[26] == 0x00 && h.peerData.MatchId > 0 {
					// replay request and match started
					logger123.Info("Th123 spectacle int request from spectator")
					return true, append([]byte{orig[0]}, h.peerData.InitSuccessPkg[:]...)
				}
			}

		} else {
			logger123.Warn("INIT_REQUEST with strange length ", len(orig)-1)
		}

	case INIT_SUCCESS_123:
		if len(orig)-1 == 81 {
			switch spectate123type(orig[6]) {
			case SPECTATE_FOR_SPECTATOR_123:
				copy(h.peerData.InitSuccessPkg[:], orig[1:])
				h.repReqStatus = SEND_123
				logger123.Info("Th123 spectacle INIT_SUCCESS")

				return false, nil
			}
		} else {
			logger123.Warn("INIT_SUCCESS with strange length ", len(orig)-1)
		}

	case QUIT_123:
		if len(orig)-1 == 1 {
			logger123.Info("Th123 peer quit")
			if orig[0] == h.peerId {
				h.PeerStatus = INACTIVE_123
				h.repReqStatus = INIT_123
			} else {
				return false, nil
			}
		} else {
			logger123.Warn("QUIT with strange length ", len(orig)-1)
		}

	case HOST_GAME_123:
		switch data123pkg(orig[2]) {
		case GAME_MATCH_123:
			if len(orig)-1 == 99 {
				if orig[0] == h.peerId {
					// game match data
					copy(h.peerData.HostInfo[:], orig[3:48])
					copy(h.peerData.ClientInfo[:], orig[48:93])
					h.peerData.StageId = orig[93]
					h.peerData.MusicId = orig[94]
					copy(h.peerData.RandomSeeds[:], orig[95:99])
					h.peerData.MatchId = orig[99]
					h.peerData.ReplayData[orig[99]] = make([]uint16, 1) // 填充一个 garbage
					h.peerData.ReplayEnd[orig[99]] = false

					h.repReqStatus = SEND_123

					logger123.Info("Th123 new match ", orig[99])

					return false, nil
				}
			} else if len(orig)-1 != 59 {
				logger123.Warn("HOST_GAME GAME_MATCH with strange length ", len(orig)-1)
			}

		case GAME_REPLAY_123:
			if orig[0] == h.peerId {
				// game replay data
				if len(orig) > 4 && len(orig)-4 == int(orig[3]) {
					timeDelay := time.Now().Sub(h.repReqTime)

					r, err := zlib.NewReader(bytes.NewBuffer(orig[4:]))
					if err != nil {
						logger123.WithError(err).Error("Th123 new zlib reader error")
					} else {
						ans := make([]byte, utils.TransBufSize)
						n, err := r.Read(ans)
						_ = r.Close()

						if err == io.EOF {
							//   game_inputs_count 60 MAX
							if n >= 10 && n-10 == int(ans[9])*2 {
								frameId := int(ans[0]) | int(ans[1])<<8 | int(ans[2])<<16 | int(ans[3])<<24
								endFrameId := int(ans[4]) | int(ans[5])<<8 | int(ans[6])<<16 | int(ans[7])<<24

								data := h.peerData.ReplayData[ans[8]]
								getDataLen := len(data) - 1
								if getDataLen == -1 {
									logger123.Error("Th123 no such match: ", ans[8])
								} else if frameId-getDataLen <= int(ans[9]) {
									newDataLen := frameId - getDataLen

									if newDataLen > 0 {
										newData := make([]uint16, newDataLen)

										for i := 0; i < newDataLen; i++ {
											newData[newDataLen-1-i] = uint16(ans[10+i*2])<<8 | uint16(ans[11+i*2])
										}

										h.peerData.ReplayData[ans[8]] = append(data, newData...)

										if len(h.peerData.ReplayData[ans[8]])-1 != frameId {
											logger123.Error("Th123 replay data not match after append new data")
										}
									}

									if endFrameId != 0 && endFrameId == frameId && !h.peerData.ReplayEnd[ans[8]] {
										logger123.Info("Th123 match end: ", ans[8])
										h.peerData.ReplayEnd[ans[8]] = true
										h.PeerStatus = BATTLE_WAIT_ANOTHER_123
									}

									h.repReqTime = time.Time{}
									h.repReqDelay = timeDelay
									h.repReqStatus = SEND_123
								} else {
									logger123.Warn("Replay data package drop: frame id ", frameId, " length ", ans[9])
								}
							} else {
								logger123.Error("Replay data content invalid")
							}
						} else {
							logger123.WithError(err).Error("Zlib decode error")
						}
						// fmt.Println(err == io.EOF, err, ans[:n])
					}
				} else {
					logger123.Warn("Th123 replay data invalid")
				}
			}

			return false, nil
		}

	case CLIENT_GAME_123:
		switch data123pkg(orig[2]) {
		case GAME_LOADED_ACK_123:
			if orig[3] == 0x05 {
				logger123.Info("Th123 battle loaded")
				h.PeerStatus = BATTLE_123
			}

		case GAME_REPLAY_REQUEST_123:
			if len(orig)-1 == 7 {

				// game replay request from spectator
				frameId := int(orig[3]) | int(orig[4])<<8 | int(orig[5])<<16 | int(orig[6])<<24
				if frameId == 0xffffffff || orig[7] < h.peerData.MatchId {
					data := []byte{orig[0], byte(HOST_GAME_123), byte(GAME_MATCH_123)}
					data = append(data, h.peerData.HostInfo[:]...)
					data = append(data, h.peerData.ClientInfo[:]...)
					data = append(data, h.peerData.StageId)
					data = append(data, h.peerData.MusicId)
					data = append(data, h.peerData.RandomSeeds[:]...)
					data = append(data, h.peerData.MatchId)

					logger123.Debug("GAME_REPLAY_REQUEST reply with GAME_MATCH")
					logger123.Info("New spectator join")
					h.spectatorCount++

					return true, data

				} else if orig[7] == h.peerData.MatchId {

					data := []byte{orig[0], byte(HOST_GAME_123), byte(GAME_REPLAY_123)}

					// replay data
					repData := h.peerData.ReplayData[h.peerData.MatchId]
					endFrameId := len(repData) - 1
					sendFrameId := int(math.Min(float64(endFrameId), float64(frameId+60)))
					var gameInput []byte
					if frameId <= endFrameId {
						// send 60 max
						for i := sendFrameId; i > frameId; i-- {
							gameInput = append(gameInput, []byte{byte(repData[i] >> 8), byte(repData[i])}...)
						}
					}
					if len(gameInput)%4 != 0 {
						logger123.Warn("Th123 game input is not time of 4 ?")
					}

					// append addition data (frameId endFrameId matchId inputCount inputs)
					gameInput = append([]byte{h.peerData.MatchId, byte(len(gameInput) >> 1)}, gameInput...)
					if h.peerData.ReplayEnd[h.peerData.MatchId] {
						if frameId == 0 {
							// when some spectator finish fetching data,
							// it will send 0 frame id, which lead to strange bug
							gameInput = []byte{h.peerData.MatchId, 0}
							sendFrameId = 0
						}
						gameInput = append([]byte{byte(endFrameId), byte(endFrameId >> 8), byte(endFrameId >> 16), byte(endFrameId >> 24)}, gameInput...)
					} else {
						gameInput = append([]byte{0, 0, 0, 0}, gameInput...)
					}
					gameInput = append([]byte{byte(sendFrameId), byte(sendFrameId >> 8), byte(sendFrameId >> 16), byte(sendFrameId >> 24)}, gameInput...)

					// zlib compress
					var zlibData bytes.Buffer
					zlibw := zlib.NewWriter(&zlibData)
					_, err := zlibw.Write(gameInput)
					if err != nil {
						logger123.WithError(err).Error("Th123 zlib compress error")
					}
					_ = zlibw.Close()

					// make data (0x09 size data)
					data = append(data, byte(zlibData.Len()))
					data = append(data, zlibData.Bytes()...)

					if endFrameId == sendFrameId && h.PeerStatus == INACTIVE_123 {
						// let spectator quit
						logger123.Info("Th123 quit spectator")
						return true, []byte{orig[0], byte(QUIT_123)}
					}
					return true, data

				}
			} else {
				logger123.Warn("CLIENT_GAME GAME_REPLAY_REQUEST with strange length ", len(orig)-1)
			}

			return false, nil
		}
	}

	return false, orig
}

func (h *Hisoutensoku) GoroutineFunc(tunnelConn interface{}, _ *net.UDPConn) {
	logger123.Info("Th123 plugin goroutine start")
	defer logger123.Info("Th123 plugin goroutine quit")

bigLoop:
	for {
		if h.PeerStatus == BATTLE_123 {
			switch h.repReqStatus {
			case INIT_123:
				if h.peerData.Spectator {
					gameId := h.gameId[h.peerId]
					requestData := append([]byte{h.peerId, byte(INIT_REQUEST_123)}, gameId[:]...) // INIT_REQUEST and game id
					requestData = append(requestData, make([]byte, 8)...)                         // garbage
					requestData = append(requestData, 0x00)                                       // spectacle request
					requestData = append(requestData, 0x00)                                       //  data length 0
					requestData = append(requestData, make([]byte, 38)...)                        // make it 65 bytes long

					var err error
					switch conn := tunnelConn.(type) {
					case quic.Stream:
						_, err = conn.Write(utils.NewDataFrame(utils.DATA, requestData))
					case *net.TCPConn:
						_, err = conn.Write(utils.NewDataFrame(utils.DATA, requestData))
					}
					if err != nil {
						logger123.WithError(err).Error("Th123 send INIT_REQUEST error")
						break
					}
					logger123.Info("Th123 send spectacle INIT_REQUEST")
				}

			case SEND_123, SEND_AGAIN_123:
				getId := len(h.peerData.ReplayData[h.peerData.MatchId]) - 1

				requestData := []byte{h.peerId, byte(CLIENT_GAME_123), byte(GAME_REPLAY_REQUEST_123),
					byte(getId), byte(getId >> 8), byte(getId >> 16), byte(getId >> 24), h.peerData.MatchId}

				h.repReqTime = time.Now()

				var err error
				switch conn := tunnelConn.(type) {
				case quic.Stream:
					_, err = conn.Write(utils.NewDataFrame(utils.DATA, requestData))
				case *net.TCPConn:
					_, err = conn.Write(utils.NewDataFrame(utils.DATA, requestData))
				}
				if err != nil {
					logger123.WithError(err).Error("Th123 send GAME_REPLAY_REQUEST error")
					break bigLoop
				}

				h.repReqStatus = SENT0_123

			case SENT0_123, SENT1_123:
				h.repReqStatus++
			}
		}

		if h.quitFlag {
			break
		}

		// 15 request per second
		time.Sleep(time.Millisecond * 66)
	}
}

func (h *Hisoutensoku) GetReplayDelay() time.Duration {
	return h.repReqDelay
}

func (h *Hisoutensoku) GetSpectatorCount() int {
	return h.spectatorCount
}

func (h *Hisoutensoku) SetQuitFlag() {
	h.quitFlag = true
}
