package client

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"time"
	"unsafe"

	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

/*
#cgo pkg-config: zlib
#include <stdlib.h>
#include <string.h>
#include <zlib.h>

void zlib_encode(size_t len, const unsigned char *orig, size_t *retLen, unsigned char **ret)
{
	unsigned char enc[1024];
	z_stream enc_stream;

	enc_stream.zalloc = Z_NULL;
	enc_stream.zfree = Z_NULL;
	enc_stream.opaque = Z_NULL;

	enc_stream.avail_in = (uInt)len;
	enc_stream.next_in = (Bytef *)orig;
	enc_stream.avail_out = (uInt)sizeof(enc);
	enc_stream.next_out = (Bytef *)enc;

	deflateInit(&enc_stream, Z_DEFAULT_COMPRESSION);
	deflate(&enc_stream, Z_FINISH);
	deflateEnd(&enc_stream);

	*retLen = enc_stream.total_out;
	*ret = (unsigned char *)malloc(sizeof(unsigned char) * enc_stream.total_out);
	memcpy(*ret, enc, enc_stream.total_out);
}
*/
import "C"

var logger155 = logrus.WithField("Hyouibana", "internal")

type type155pkg byte

const (
	CLIENT_T_ACK_155 type155pkg = iota     // 0x00
	HOST_T_ACK_155                         // 0x01
	INIT_ACK_155     type155pkg = iota + 2 // 0x04
	HOST_T_155                             // 0x05
	CLIENT_T_155                           // 0x06
	PUNCH_155                              // 0x07
	INIT_155                               // 0x08
	INIT_REQUEST_155                       // 0x09
	INIT_SUCCESS_155 type155pkg = iota + 3 // 0x0b
	INIT_ERROR_155                         // 0x0c
	HOST_QUIT_155    type155pkg = iota + 5 // 0x0f
	CLIENT_QUIT_155                        // 0x10
	HOST_GAME_155    type155pkg = iota + 6 // 0x12
	CLIENT_GAME_155                        // 0x13
)

type data155pkg byte

const (
	GAME_SELECT_155         data155pkg = iota + 4 // 0x04
	GAME_INPUT_155          data155pkg = iota + 5 // 0x06
	GAME_REPLAY_REQUEST_155 data155pkg = iota + 7 // 0x09
	GAME_REPLAY_MATCH_155                         // 0x0a
	GAME_REPLAY_DATA_155                          // 0x0b
	GAME_REPLAY_END_155                           // 0x0c
)

type match155status byte

const (
	MATCH_WAIT_155   match155status = iota // no game start
	MATCH_ACCEPT_155                       // peer connected
	MATCH_SPECT_ACK_155
	MATCH_SPECT_INIT_155
	MATCH_SPECT_SUCCESS_155 // fetching replay data
	MATCH_SPECT_ERROR_155   // cannot get replay data
)

var th155id = [19]byte{0x57, 0x09, 0xf6, 0x67, 0xf0, 0xfd, 0x4b, 0xd0, 0xb9, 0x9a, 0x74, 0xf8, 0x38, 0x33, 0x81, 0x88, 0x00, 0x00, 0x00}

// magic [6] [8] th155 0x71 th155_beta 0x72 for spectacle
var th155ConfMagic = [12]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0xff, 0x00, 0x00, 0x01}

// version [123:131]
var th155ConfOrig = [156]C.uchar{0x10, 0x00, 0x00, 0x08, 0x08, 0x00, 0x00, 0x00, 0x69, 0x73, 0x5f, 0x77, 0x61, 0x74, 0x63, 0x68,
	0x08, 0x00, 0x00, 0x01, 0x00, 0x10, 0x00, 0x00, 0x08, 0x05, 0x00, 0x00, 0x00, 0x65, 0x78, 0x74, 0x72, 0x61, 0x01, 0x00,
	0x00, 0x01, 0x10, 0x00, 0x00, 0x08, 0x0b, 0x00, 0x00, 0x00, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x5f, 0x77, 0x61, 0x74, 0x63,
	0x68, 0x08, 0x00, 0x00, 0x01, 0x01, 0x10, 0x00, 0x00, 0x08, 0x04, 0x00, 0x00, 0x00, 0x6e, 0x61, 0x6d, 0x65, 0x10, 0x00,
	0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x08, 0x0a, 0x00, 0x00, 0x00, 0x62, 0x61, 0x74, 0x74, 0x6c, 0x65,
	0x5f, 0x6e, 0x75, 0x6d, 0x02, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x08, 0x07, 0x00, 0x00, 0x00,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x10, 0x00, 0x00, 0x08, 0x05,
	0x00, 0x00, 0x00, 0x63, 0x6f, 0x6c, 0x6f, 0x72, 0x02, 0x00, 0x00, 0x05, 0x0a, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01}

func zlibDataDecodeError(l int, d []byte) string {
	if len(d) < 3 || d[0] != 0x78 || d[1] != 0x9c {
		return "NOT_ZLIB_DATA_ERROR"
	}

	b := bytes.NewBuffer(d)
	r, err := zlib.NewReader(b)
	if err != nil {
		return err.Error()
	}

	ans := make([]byte, l*2)
	n, err := r.Read(ans)
	if err != io.EOF {
		return err.Error()
	}
	_ = r.Close()

	if l != n {
		return "ZLIB_LENGTH_NOT_MATCH_ERROR"
	}

	dataStr := ""

	i, j, s := 0, 0, 0
	for j < n {
		switch s {
		case 0:
			if ans[j] == 0x10 {
				s++
			} else {
				s = 0
			}
		case 1, 2:
			if ans[j] == 0x00 {
				s++
			} else {
				s = 0
			}
		case 3:
			if ans[j] == 0x08 {
				s++
			} else {
				s = 0
			}
		case 4:
			nl := utils.LittleIndia2Int(ans[j : j+4])
			i = j + 4 + nl
			dataStr += string(ans[j+4:i]) + " "

			s = 0
		}

		j++
	}

	return dataStr
}

func zlibDataDecodeSignConfVersion(l int, d []byte) error {
	if len(d) < 3 || d[0] != 0x78 || d[1] != 0x9c {
		return errors.New("NOT_ZLIB_DATA_ERROR")
	}

	b := bytes.NewBuffer(d)
	r, err := zlib.NewReader(b)
	if err != nil {
		return err
	}

	ans := make([]byte, l*2)
	n, err := r.Read(ans)
	if err != io.EOF {
		return err
	}
	_ = r.Close()

	if l != n {
		return errors.New("ZLIB_LENGTH_NOT_MATCH_ERROR")
	}

	i, j, s := 0, 0, 0
	for j < n {
		switch s {
		case 0:
			if ans[j] == 0x10 {
				s++
			} else {
				s = 0
			}
		case 1, 2:
			if ans[j] == 0x00 {
				s++
			} else {
				s = 0
			}
		case 3:
			if ans[j] == 0x08 {
				s++
			} else {
				s = 0
			}
		case 4:
			nl := utils.LittleIndia2Int(ans[j : j+4])
			i = j + 4 + nl
			if string(ans[j+4:i]) == "version" {
				for k, l := i, 123; l < 131; {
					th155ConfOrig[l] = (C.uchar)(ans[k])

					l++
					k++
				}
				switch th155ConfOrig[128] {
				case 0xaa: // th155
					th155ConfMagic[6] = 0x71
					th155ConfMagic[6] = 0x71
				case 0xab: // th155_beta
					th155ConfMagic[6] = 0x72
					th155ConfMagic[6] = 0x72
				default:
					logger155.Warn("Th155 plugin find unsupported version code ", th155ConfOrig[123:131])
				}
				return nil
			}

			s = 0
		}

		j++
	}

	return errors.New("VERSION_NOT_FOUND_ERROR")
}

func zlibDataEncodeConf() (int, []byte) {
	var cLen C.size_t = 0
	var cAns *C.uchar
	var ans []byte

	C.zlib_encode(156, &th155ConfOrig[0], &cLen, &cAns)
	cIAns := (*[utils.TransBufSize]C.uchar)(unsafe.Pointer(cAns))
	for i := 0; i < int(cLen); i++ {
		ans = append(ans, byte(cIAns[i]))
	}

	return int(cLen), ans
}

type Hyouibana struct {
	peerId byte // current host/client peer id (udp mutex id)

	// peerHPTime      time.Time      // most resent current host/client peer HOST_T package met time
	// peerQuit        bool           // current match met HOST_QUIT or CLIENT_QUIT
	peerHostT       []byte         // spectator peer id need to send HOST_T
	MatchStatus     match155status // current match status
	matchEnd        bool           // current match end
	matchId         int            // current match id
	matchInfo       []byte         // current match info
	matchRandId     int            // current match random id
	initSuccessInfo []byte         // replay init success info
	initErrorInfo   []byte         // replay init error info
	frameId         [2]int         // replay frame record id
	frameRec        [2][]byte      // replay frame record
	timeId          int64          // th155 protocol client start time in ms
	randId          int32          // th155 protocol random id

	spectatorCount int  // spectator counter
	quitFlag       bool // plugin quit flag
}

// NewHyouibana new Hyouibana spectating server
func NewHyouibana() *Hyouibana {
	return &Hyouibana{
		// peerHPTime:  time.Time{},
		// peerQuit:    false,
		MatchStatus:    MATCH_WAIT_155,
		matchEnd:       false,
		matchId:        0,
		frameId:        [2]int{0, 0},
		frameRec:       [2][]byte{},
		timeId:         time.Now().UnixMilli(),
		randId:         rand.Int31(),
		spectatorCount: 0,
		quitFlag:       false,
	}
}

// WriteFunc from game host to client
// orig: original data leads with 1 byte of client id
func (h *Hyouibana) WriteFunc(orig []byte) (bool, []byte) {

	switch type155pkg(orig[1]) {

	case HOST_T_155:
		/*if len(orig)-1 == 12 {
			if h.MatchStatus != MATCH_WAIT_155 && orig[0] == h.peerId {
				h.peerHPTime = time.Now()
			}
		} else {
			logger155.Warn("HOST_T with strange length ", len(orig)-1)
		}*/

	case PUNCH_155:
		if len(orig)-1 == 32 {
			if h.MatchStatus != MATCH_WAIT_155 && orig[0] == h.peerId {
				orig[2] = 0x02
				orig[3], orig[4] = 0x01, 0x00
				return true, orig
			}
		} else {
			logger155.Warn("PUNCH with strange length ", len(orig)-1)
		}

	case INIT_SUCCESS_155:
		if len(orig)-1 > 52 {
			// h.peerHPTime = time.Now()
			// h.peerQuit = false
			h.matchEnd = false
			h.matchId = 0
			// h.matchRandId = utils.LittleIndia2Int(orig[5:9])
			h.frameId[0], h.frameId[1] = 0, 0
			h.frameRec[0], h.frameRec[1] = []byte{}, []byte{}

			h.peerId = orig[0]
			h.MatchStatus = MATCH_ACCEPT_155

			logger155.Info("Met INIT_SUCCESS")
		} else {
			logger155.Warn("INIT_SUCCESS with strange length ", len(orig)-1)
		}

	case HOST_QUIT_155:
		if len(orig)-1 == 12 {
			if h.matchId == 0 || h.MatchStatus == MATCH_SPECT_ERROR_155 {
				h.MatchStatus = MATCH_WAIT_155
				logger155.Info("Met HOST_QUIT and reset")
			}
		} else {
			logger155.Warn("HOST_QUIT　", orig[1], " with strange length ", len(orig)-1)
		}

	case CLIENT_QUIT_155:
		logger155.Warn("CLIENT_QUIT should not appear here")

	}

	return false, orig
}

// ReadFunc from game client to host
// orig: original data leads with 1 byte of client id
func (h *Hyouibana) ReadFunc(orig []byte) (bool, []byte) {

	if h.MatchStatus == MATCH_WAIT_155 {
		switch type155pkg(orig[1]) {

		case INIT_REQUEST_155:
			if len(orig)-1 > 40 {
				err := zlibDataDecodeSignConfVersion(utils.LittleIndia2Int(orig[37:41]), orig[41:])
				if err != nil {
					logger155.WithError(err).Warn("INIT_REQUEST decode version failed")
				}
			} else {
				logger155.Warn("INIT_REQUEST here　with strange length ", len(orig)-1)
			}

		}
	} else if orig[0] != h.peerId {
		switch type155pkg(orig[1]) {

		case HOST_T_ACK_155:
			if len(orig)-1 != 12 {
				logger155.Warn("HOST_T_ACK with strange length ", len(orig)-1)
			}

		case CLIENT_T_155:
			if len(orig)-1 == 20 {
				repData := []byte{orig[0], byte(CLIENT_T_ACK_155), 0x00, 0x00, 0x00}
				repData = append(repData, orig[5:9]...)
				repData = append(repData, orig[17:21]...)

				h.peerHostT = append(h.peerHostT, orig[0])

				return true, repData
			} else {
				logger155.Warn("CLIENT_T　with strange length ", len(orig)-1)
			}

		case INIT_155:
			if len(orig)-1 == 24 {
				return true, []byte{orig[0], byte(INIT_ACK_155)}
			} else {
				logger155.Warn("INIT　with strange length ", len(orig)-1)
			}

		case INIT_REQUEST_155:
			if len(orig)-1 > 40 {
				switch orig[31] {

				case 0x6b: // battle
					// message busy
					errData := []byte{orig[0], byte(INIT_ERROR_155), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x28, 0x00, 0x28, 0x00, 0x00, 0x01,
						0x1f, 0x00, 0x00, 0x00, 0x78, 0x9c, 0x13, 0x60, 0x60, 0xe0, 0x60, 0x67, 0x60, 0x60, 0xc8, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c,
						0x4f, 0x15, 0x00, 0x72, 0x59, 0x80, 0xdc, 0xa4, 0xd2, 0xe2, 0x4a, 0x46, 0x06, 0x06, 0x46, 0x00, 0x4a, 0xa5, 0x04, 0xe6}
					logger155.Info("New spector connect error with message busy")

					return true, errData

				case 0x70, 0x71: // spectacle
					var repData []byte
					switch h.MatchStatus {

					case MATCH_SPECT_SUCCESS_155:
						repData = []byte{orig[0], byte(INIT_SUCCESS_155), 0x00, 0x00, 0x00, byte(h.matchRandId), byte(h.matchRandId >> 8), byte(h.matchRandId >> 16), byte(h.matchRandId >> 24)}
						repData = append(repData, h.initSuccessInfo...)

						h.spectatorCount++
						logger155.Info("New spector connected")

						return true, repData

					case MATCH_SPECT_ERROR_155:
						repData = []byte{orig[0], byte(INIT_ERROR_155)}
						repData = append(repData, h.initErrorInfo...)
						logger155.Info("New spector connect error with host error message")

						return true, repData

					default:
						// message ready (will not reach in fact)
						repData = []byte{byte(INIT_ERROR_155), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x29, 0x00, 0x29, 0x00, 0x00, 0x01,
							0x20, 0x00, 0x00, 0x00, 0x78, 0x9c, 0x13, 0x60, 0x60, 0xe0, 0x60, 0x67, 0x60, 0x60, 0xc8, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c,
							0x4f, 0x15, 0x00, 0x72, 0x59, 0x81, 0xdc, 0xa2, 0xd4, 0xc4, 0x94, 0x4a, 0x46, 0x06, 0x06, 0x46, 0x00, 0x51, 0x07, 0x05, 0x39}
						logger155.Info("New spector connect error with message ready")

						return true, repData
					}

				default:
					logger155.Warn("Th155 plugin spectator server get unknown INIT_REQUEST ", orig[1:])
				}
			} else {
				logger155.Warn("INIT_REQUEST　with strange length ", len(orig)-1)
			}

		case HOST_QUIT_155:
			logger155.Warn("HOST_QUIT should not appear here")

		case CLIENT_QUIT_155:
			if len(orig)-1 == 12 {
				if h.matchId == 0 || h.MatchStatus == MATCH_SPECT_ERROR_155 {
					h.MatchStatus = MATCH_WAIT_155
					logger155.Info("Met CLIENT_QUIT and reset")
				}
			} else {
				logger155.Warn("CLIENT_QUIT　", orig[1], " with strange length ", len(orig)-1)
			}

		case CLIENT_GAME_155:
			switch data155pkg(orig[3]) {

			case GAME_REPLAY_REQUEST_155:
				if len(orig)-1 == 22 || len(orig)-1 == 14 {
					mid, fid0, fid1 := utils.LittleIndia2Int(orig[7:11]), 0, 0
					if len(orig)-1 == 22 {
						fid0, fid1 = utils.LittleIndia2Int(orig[15:19])<<1, utils.LittleIndia2Int(orig[19:23])<<1
					}
					repData := []byte{orig[0]}

					if mid == 0 {
						if h.matchId == 0 {
							break
						} else if h.MatchStatus == MATCH_WAIT_155 { // TODO: impossible to reach this code block
							// HOST_QUIT
							logger155.Info("Quit spectator")
							repData = append(repData, []byte{byte(HOST_QUIT_155), 0x00, 0x00, 0x00,
								byte(h.matchRandId), byte(h.matchRandId >> 8), byte(h.matchRandId >> 16), byte(h.matchRandId >> 24),
								0x00, 0x00, 0x00, 0x00}...)
						} else {
							// GAME_REPLAY_MATCH
							repData = append(repData, h.matchInfo...)
						}
					} else if mid != h.matchId || h.matchEnd {
						// GAME_REPLAY_END
						repData = append(repData, []byte{byte(HOST_GAME_155), byte(GAME_REPLAY_END_155), 0x00, 0x00, 0x00,
							byte(mid), byte(mid >> 8), byte(mid >> 16), byte(mid >> 24)}...)
					} else {
						fidE0, fidE1 := int(math.Min(float64(h.frameId[0]*2), float64(fid0+48))), int(math.Min(float64(h.frameId[1]*2), float64(fid1+48))) // max=24*2
						replay := [2][]byte{h.frameRec[0][fid0:fidE0], h.frameRec[1][fid1:fidE1]}
						fid0 >>= 1
						fid1 >>= 1
						fidE0 >>= 1
						fidE1 >>= 1
						repData = append(repData, []byte{byte(HOST_GAME_155), byte(GAME_REPLAY_DATA_155), 0x02, 0x00, 0x00, byte(mid), byte(mid >> 8), byte(mid >> 16), byte(mid >> 24),
							byte(fid0), byte(fid0 >> 8), byte(fid0 >> 16), byte(fid0 >> 24), byte(fidE0), byte(fidE0 >> 8), byte(fidE0 >> 16), byte(fidE0 >> 24)}...)
						repData = append(repData, replay[0]...)
						repData = append(repData, byte(fid1), byte(fid1>>8), byte(fid1>>16), byte(fid1>>24), byte(fidE1), byte(fidE1>>8), byte(fidE1>>16), byte(fidE1>>24))
						repData = append(repData, replay[1]...)
					}

					return true, repData
				} else {
					logger155.Warn("CLIENT_GAME GAME_REPLAY_REQUEST　with strange length ", len(orig)-1)
				}

			default:
				logger155.Warn("Th155 plugin spectator server get unexpected CLIENT_GAME ", orig[1:])
			}

		}

		return false, nil
	}

	return false, orig
}

func (h *Hyouibana) GoroutineFunc(tunnelConn interface{}, conn *net.UDPConn) {
	logger155.Info("Th155 plugin goroutine start")
	defer logger155.Info("Th155 plugin goroutine quit")

	hostAddr, err := net.ResolveUDPAddr("udp", conn.RemoteAddr().String())
	if err != nil {
		logger155.WithError(err).Error("Th155 plugin goroutine cannot resolve host udp address")
		return
	}
	hostConn, err := net.DialUDP("udp", nil, hostAddr)
	if err != nil {
		logger155.WithError(err).Error("Th155 plugin goroutine cannot get udp connection to host")
		return
	}
	hostConnClosed := false
	defer func() {
		if !hostConnClosed {
			_ = hostConn.Close()
		}
	}()

	ch := make(chan int, 2)

	// replay request
	go func() {
		defer func() {
			ch <- 1
		}()

		timeWait := 0
		connSucc := time.Now()
	bigLoop:
		for {

			if h.quitFlag {
				break
			}

			time.Sleep(time.Millisecond * 33)
			timeWait += 1
			timeWait %= 30 // 1s

			if timeWait == 0 {

				switch h.MatchStatus {

				case MATCH_WAIT_155, MATCH_SPECT_ERROR_155:
					connSucc = time.Now()

				default:
					timeDiff := time.Now().UnixMilli() - h.timeId
					specData := append([]byte{byte(CLIENT_T_155)}, []byte{0x01, 0x71, 0x00, byte(h.randId), byte(h.randId >> 8), byte(h.randId >> 16), byte(h.randId >> 24)}...)
					specData = append(specData, []byte{0x03, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00}...)
					specData = append(specData, []byte{byte(timeDiff), byte(timeDiff >> 8), byte(timeDiff >> 16), byte(timeDiff >> 24)}...)

					_, err := hostConn.Write(specData)
					if err == nil {
						connSucc = time.Now()
					} else {
						logger155.WithError(err).Warn("Th155 plugin host conn write CLIENT_T error")
					}

					// spectator HOST_T
					var ids []byte
					ids, h.peerHostT = h.peerHostT, []byte{}
					for _, id := range ids {
						repData := []byte{id, byte(HOST_T_155), 0x00, 0x00, 0x00, byte(h.matchRandId), byte(h.matchRandId >> 8), byte(h.matchRandId >> 16), byte(h.matchRandId >> 24),
							byte(timeDiff), byte(timeDiff >> 8), byte(timeDiff >> 16), byte(timeDiff >> 24)}

						switch conn := tunnelConn.(type) {
						case quic.Stream:
							_, err = conn.Write(utils.NewDataFrame(utils.DATA, repData))
						case *net.TCPConn:
							_, err = conn.Write(utils.NewDataFrame(utils.DATA, repData))
						}

						if err != nil {
							logger155.WithError(err).Warn("Th155 realize tunnel disconnected")
							break bigLoop
						}
					}
				}

			} else if timeWait%2 == 0 {

				switch h.MatchStatus {

				case MATCH_WAIT_155:
					connSucc = time.Now()

				case MATCH_ACCEPT_155:
					specData := append([]byte{byte(INIT_155)}, th155id[:]...)
					specData = append(specData, []byte{byte(h.randId), byte(h.randId >> 8), byte(h.randId >> 16), byte(h.randId >> 24)}...)

					_, err := hostConn.Write(specData)
					if err == nil {
						connSucc = time.Now()
					} else {
						logger155.WithError(err).Error("Th155 plugin host conn write INIT error")
					}

				case MATCH_SPECT_ACK_155:
					l, th155SpecConf := zlibDataEncodeConf()
					if l == 0 || th155SpecConf == nil {
						logger155.WithError(err).Error("Th155 plugin INIT_REQUEST zlib compression error")
						h.MatchStatus = MATCH_SPECT_ERROR_155
						break
					}
					specData := append([]byte{byte(INIT_REQUEST_155)}, th155id[:]...)
					specData = append(specData, byte(h.randId), byte(h.randId>>8), byte(h.randId>>16), byte(h.randId>>24))
					specData = append(specData, th155ConfMagic[:]...)   // spectacle
					specData = append(specData, 0x9c, 0x00, 0x00, 0x00) // 156
					specData = append(specData, th155SpecConf[:l]...)

					_, err = hostConn.Write(specData)
					if err == nil {
						connSucc = time.Now()
					} else {
						logger155.WithError(err).Error("Th155 plugin host conn write INIT_REQUEST error")
					}
					h.MatchStatus = MATCH_SPECT_INIT_155

				case MATCH_SPECT_SUCCESS_155:
					var specData []byte
					if h.matchEnd {
						specData = []byte{byte(CLIENT_GAME_155), 0x01, byte(GAME_REPLAY_REQUEST_155), 0x00, 0x00, 0x00,
							0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
							0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
					} else {
						specData = []byte{byte(CLIENT_GAME_155), 0x01, byte(GAME_REPLAY_REQUEST_155), 0x00, 0x00, 0x00,
							byte(h.matchId), byte(h.matchId >> 8), byte(h.matchId >> 16), byte(h.matchId >> 24),
							0x00, 0x00, 0x00, 0x00,
							byte(h.frameId[0]), byte(h.frameId[0] >> 8), byte(h.frameId[0] >> 16), byte(h.frameId[0] >> 24),
							byte(h.frameId[1]), byte(h.frameId[1] >> 8), byte(h.frameId[1] >> 16), byte(h.frameId[1] >> 24)}
					}

					_, err := hostConn.Write(specData)
					if err == nil {
						connSucc = time.Now()
					} else {
						logger155.WithError(err).Error("Th155 plugin host conn write GAME_REPLAY_REQUEST error")
					}
				}

			}

			if h.MatchStatus != MATCH_WAIT_155 && time.Now().Sub(connSucc) > time.Second { // lost host connection
				logger155.Error("Th155 plugin host connection lost")
				h.MatchStatus = MATCH_WAIT_155
				h.matchEnd = true
			}

		}

		hostConnClosed = true
		_ = hostConn.Close()

	}()

	// replay record
	go func() {
		defer func() {
			ch <- 1
		}()

		buf := make([]byte, utils.TransBufSize)

		for {
			time.Sleep(time.Millisecond * 33)

			n, err := hostConn.Read(buf)

			if err != nil {
				logger155.WithError(err).Error("Th155 plugin host conn read error")
				break
			}
			switch type155pkg(buf[0]) {

			case CLIENT_T_ACK_155:

			case INIT_ACK_155:
				if n == 1 {
					h.MatchStatus = MATCH_SPECT_ACK_155
				} else {
					logger155.Warn("INIT_ACK with strange length ", n)
				}

			case PUNCH_155:
				if n == 32 {
					buf[1] = 0x02
					buf[2], buf[3] = 0x01, 0x00
					_, err = hostConn.Write(buf[:n])
					if err != nil {
						logger155.WithError(err).Warn("Th155 plugin host punch reply write error")
					}
				} else {
					logger155.Warn("PUNCH with strange length ", n)
				}

			case HOST_T_155:
				if n == 12 {
					buf[0] = byte(HOST_T_ACK_155)
					_, err = hostConn.Write(buf[:n])
					if err != nil {
						logger155.WithError(err).Warn("Th155 plugin host host_t reply write error")
					}
				} else {
					logger155.Warn("HOST_T with strange length ", n)
				}

			case INIT_SUCCESS_155:
				if n > 52 {
					h.MatchStatus = MATCH_SPECT_SUCCESS_155
					h.matchRandId = utils.LittleIndia2Int(buf[4:8])
					h.initSuccessInfo = make([]byte, n-8)
					copy(h.initSuccessInfo, buf[8:n])
					logger155.Info("Th155 plugin spectator get INIT_SUCCESS")
				} else {
					logger155.Warn("INIT_SUCCESS with strange length ", n)
				}

			case INIT_ERROR_155:
				if n > 20 {
					h.MatchStatus = MATCH_SPECT_ERROR_155
					h.initErrorInfo = make([]byte, n-1)
					copy(h.initErrorInfo, buf[1:n])
					logger155.Info("Th155 plugin spectator get INIT_ERROR with ", zlibDataDecodeError(utils.LittleIndia2Int(buf[16:20]), buf[20:n]))
				} else {
					logger155.Warn("INIT_ERROR with strange length ", n)
				}

			case HOST_GAME_155:
				switch data155pkg(buf[1]) {

				case GAME_REPLAY_MATCH_155:
					if n > 21 {
						mid := utils.LittleIndia2Int(buf[5:9])
						if mid != h.matchId {
							h.matchId = mid
							h.matchEnd = false
							h.matchInfo = make([]byte, n)
							copy(h.matchInfo, buf[:n])
							h.frameId[0], h.frameId[1] = 0, 0
							h.frameRec[0], h.frameRec[1] = []byte{}, []byte{}
							logger155.Info("Th155 plugin spectator get new match id ", mid)
						}
					} else {
						logger155.Warn("HOST_GAME GAME_REPLAY_MATCH with strange length ", n)
					}

				case GAME_REPLAY_DATA_155:
					if n >= 24 {
						mid := utils.LittleIndia2Int(buf[5:9])
						if mid != h.matchId {
							logger155.Warn("Th155 plugin spectator get invalid match id ", mid, " expect ", h.matchId)
						} else {
							fidS, fidE := utils.LittleIndia2Int(buf[9:13]), utils.LittleIndia2Int(buf[13:17])
							fidL := fidE - fidS
							if fidS == h.frameId[0] {
								h.frameId[0] = fidE
								h.frameRec[0] = append(h.frameRec[0], buf[17:17+fidL*2]...)
								if len(h.frameRec[0]) != fidE*2 {
									logger155.Warn("Th155 plugin spectator get wrong record0 length after append new data ", len(h.frameRec[0]), " expect ", fidE*2)
								}
							} else {
								logger155.Warn("Th155 plugin spectator get invalid start frame id ", fidS, " expect ", h.frameId[0])
							}
							fidS, fidE = utils.LittleIndia2Int(buf[17+fidL*2:21+fidL*2]), utils.LittleIndia2Int(buf[21+fidL*2:25+fidL*2])
							if fidS == h.frameId[1] {
								h.frameId[1] = fidE
								h.frameRec[1] = append(h.frameRec[1], buf[25+fidL*2:n]...)
								if len(h.frameRec[1]) != fidE*2 {
									logger155.Warn("Th155 plugin spectator get wrong record1 length after append new data ", len(h.frameRec[1]), " expect ", fidE*2)
								}
							} else {
								logger155.Warn("Th155 plugin spectator get invalid start frame id ", fidS, " expect ", h.frameId[1])
							}

							// logger155.Debug("Th155 plugin spectator get HOST_GAME GAME_REPLAY_DATA match id ", h.matchId, " frame id ", h.frameId)
						}
					} else {
						logger155.Warn("HOST_GAME GAME_REPLAY_DATA with strange length ", n)
					}

				case GAME_REPLAY_END_155:
					if n == 9 {
						mid := utils.LittleIndia2Int(buf[5:9])
						if mid != h.matchId {
							logger155.Warn("Th155 plugin spectator get invalid match id ", mid, " expect ", h.matchId)
						} else {
							logger155.Info("Th155 plugin spectator get HOST_GAME GAME_REPLAY_END match id ", h.matchId)
							h.matchEnd = true
						}
					} else {
						logger155.Warn("HOST_GAME GAME_REPLAY_END with strange length ", n)
					}

				default:
					logger155.Warn("Th155 plugin spectator get invalid package ", buf[:n])
				}

			case HOST_QUIT_155:
				logger155.Info("Th155 plugin spectator get HOST_QUIT")
				h.MatchStatus = MATCH_WAIT_155 // the end

			default:
				logger155.Warn("Th155 plugin spectator get invalid package ", buf[:n])

			}
		}
	}()

	<-ch
}

func (h *Hyouibana) SetQuitFlag() {
	h.quitFlag = true
}

func (h *Hyouibana) GetSpectatorCount() int {
	return h.spectatorCount
}
