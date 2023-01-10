package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	client "github.com/weilinfox/youmu-thlink/client/lib"
	"github.com/weilinfox/youmu-thlink/glg-go"
	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client-gui", "internal")
var icon *gdk.Pixbuf

const appName = "白玉楼製作所 ThLink"

type status struct {
	localPort  int
	serverHost string
	tunnelType string

	userConfigChange bool

	client    *client.Client
	plugin    interface{}
	pluginNum int

	brokerTVersion byte
	brokerVersion  string

	delay           [40]time.Duration
	delayPos        int
	delayLen        int
	pluginDelay     [40]time.Duration
	pluginDelayPos  int
	pluginDelayLen  int
	pluginDelayShow bool
}

// clientStatus all of this client-gui
var clientStatus = status{
	localPort:        client.DefaultLocalPort,
	serverHost:       client.DefaultServerHost,
	tunnelType:       client.DefaultTunnelType,
	client:           client.NewWithDefault(),
	plugin:           nil,
	pluginNum:        0,
	userConfigChange: false,
	delayPos:         0,
	delayLen:         0,
	pluginDelayPos:   0,
	pluginDelayLen:   0,
	pluginDelayShow:  false,
}

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	const appID = "love.inuyasha.thlink"
	app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		logrus.WithError(err).Fatal("Could not create app")
	}

	icon, err = getIcon()
	if err != nil {
		logger.WithError(err).Error("Get icon error")
	}

	app.Connect("activate", onAppActivate)

	app.Run(os.Args)

}

var appWindow *gtk.ApplicationWindow

// onAppActivate setup Main window
func onAppActivate(app *gtk.Application) {

	var err error

	if appWindow != nil {
		if !appWindow.IsVisible() {
			appWindow.Show()
		}
		logger.Debug("Already running")
		return
	}

	clientStatus.brokerTVersion, clientStatus.brokerVersion = clientStatus.client.BrokerVersion()

	appWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		logger.WithError(err).Fatal("Could not create app window.")
	}

	// action
	appWindow.Connect("destroy", onAppDestroy)

	// simple actions
	aQuit := glib.SimpleActionNew("quit", nil)
	aQuit.Connect("activate", func() {
		onAppDestroy()
		app.Quit()
	})
	app.AddAction(aQuit)
	aAbout := glib.SimpleActionNew("about", nil)
	aAbout.Connect("activate", showAboutDialog)
	app.AddAction(aAbout)

	appWindow.Connect("destroy", onAppDestroy)

	// header bar
	menuBtn, err := gtk.MenuButtonNew()
	if err != nil {
		logger.WithError(err).Fatal("Could not create menu button.")
	}
	header, err := gtk.HeaderBarNew()
	if err != nil {
		logger.WithError(err).Fatal("Could not create header bar.")
	}
	menu := glib.MenuNew()
	menu.Append("Reset config", "app.reset")
	menu.Append("Network discovery", "app.net-disc")
	menu.Append("Tunnel status", "app.t-status")
	menu.Append("About thlink", "app.about")
	menu.Append("Quit", "app.quit")
	menuBtn.SetMenuModel(&menu.MenuModel)
	header.PackStart(menuBtn)
	header.SetShowCloseButton(true)
	header.SetTitle(appName)
	header.SetSubtitle("v" + utils.Version + "-" + strconv.Itoa(int(utils.TunnelVersion)))

	// grid
	mainGrid, err := gtk.GridNew()
	if err != nil {
		logger.WithError(err).Fatal("Could not create grid.")
	}
	mainGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	mainGrid.SetBorderWidth(10)
	mainGrid.SetRowSpacing(10)

	mainGrid.SetHAlign(gtk.ALIGN_FILL)

	// server label
	serverLabel, err := gtk.LabelNew("Server")
	if err != nil {
		logger.WithError(err).Fatal("Could not create server label.")
	}
	serverLabel.SetHAlign(gtk.ALIGN_START)

	// broker address box
	serverEntryBuf, err := gtk.EntryBufferNew(clientStatus.serverHost, len(clientStatus.serverHost))
	if err != nil {
		logger.WithError(err).Fatal("Could not create server entry buffer.")
	}
	serverEntry, err := gtk.EntryNewWithBuffer(serverEntryBuf)
	if err != nil {
		logger.WithError(err).Fatal("Could not create server entry.")
	}
	serverEntry.SetHExpand(true)
	serverEntry.Connect("changed", func() {
		clientStatus.userConfigChange = true
		host, err := serverEntry.GetText()
		if err != nil {
			logger.WithError(err).Error("Could not get server entry text.")
		}
		clientStatus.serverHost = host
		logger.Debug("Server host change to ", clientStatus.serverHost)
	})

	// ping button
	pingBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	pingLabel, err := gtk.LabelNew("Null delay")
	if err != nil {
		logger.WithError(err).Fatal("Could not create delay label.")
	}
	pingLabel.SetHExpand(true)
	pingBtn, err := gtk.ButtonNewWithLabel("Ping")
	if err != nil {
		logger.WithError(err).Fatal("Could not create ping button.")
	}
	pingBtn.SetHExpand(true)

	setPingLabel := func(delay time.Duration) {
		logger.Debugf("Display new delay %.2f ms", float64(delay.Nanoseconds())/1000000)
		pingLabel.SetText(fmt.Sprintf("%.2f ms", float64(delay.Nanoseconds())/1000000))

		clientStatus.delay[clientStatus.delayPos] = delay
		clientStatus.delayPos = (clientStatus.delayPos + 1) % 40
		if clientStatus.delayLen < 40 {
			clientStatus.delayLen++
		}

		switch p := clientStatus.plugin.(type) {
		case *client.Hisoutensoku:
			clientStatus.pluginDelayShow = true
			if p.PeerStatus == client.BATTLE {
				clientStatus.pluginDelay[clientStatus.pluginDelayPos] = p.GetReplayDelay()
				clientStatus.pluginDelayPos = (clientStatus.pluginDelayPos + 1) % 40
				if clientStatus.pluginDelayLen < 40 {
					clientStatus.pluginDelayLen++
				}
			}
		}
	}
	pingBtn.Connect("clicked", func() {

		if !clientStatus.client.Serving() {

			err := onConfigUpdate()
			if err != nil {
				logger.WithError(err).Error("Update ping failed")
				showErrorDialog(appWindow, "Update ping error", err)
				return
			}

		}

		setPingLabel(onUpdatePing())
	})
	pingBox.Add(pingLabel)
	pingBox.Add(pingBtn)

	// setup label
	setupLabel, err := gtk.LabelNew("Tunnel info")
	if err != nil {
		logger.WithError(err).Fatal("Could not create setup label.")
	}
	setupLabel.SetHAlign(gtk.ALIGN_START)

	// local port
	localPortBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		logger.WithError(err).Fatal("Could not create local port box.")
	}
	localPortLabel, err := gtk.LabelNew("Local port")
	if err != nil {
		logger.WithError(err).Fatal("Could not create local port label.")
	}
	localPortBox.Add(localPortLabel)
	localPortDefault := strconv.Itoa(client.DefaultLocalPort)
	localEntryBuf, err := gtk.EntryBufferNew(localPortDefault, len(localPortDefault))
	if err != nil {
		logger.WithError(err).Fatal("Could not create local entry buffer.")
	}
	localPortEntry, err := gtk.EntryNewWithBuffer(localEntryBuf)
	if err != nil {
		logger.WithError(err).Fatal("Could not create local entry.")
	}
	localPortEntry.SetMaxLength(5)
	localPortEntry.Connect("insert-text", func(_ *gtk.Entry, in []byte, length int) {
		for i := 0; i < length; i++ {
			if in[i] < '0' || in[i] > '9' {
				localPortEntry.StopEmission("insert-text")
				break
			}
		}
	})
	localPortEntry.Connect("changed", func() {
		clientStatus.userConfigChange = true
		port, err := localPortEntry.GetText()
		if err != nil {
			logger.WithError(err).Error("Update local port failed")
		}
		port64, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			logger.WithError(err).Error("Update local port failed")
		}
		clientStatus.localPort = int(port64)
		logger.Debug("Local port change to ", clientStatus.localPort)
	})
	localPortBox.Add(localPortEntry)
	localPortBox.SetHExpand(true)

	// protocol choose
	protoRadioBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 50)
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio box.")
	}
	protoRadioTcp, err := gtk.RadioButtonNewWithLabelFromWidget(nil, "TCP")
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio button TCP.")
	}
	protoRadioTcp.Connect("toggled", func(r *gtk.RadioButton) {
		if r.GetActive() {
			clientStatus.tunnelType = "tcp"
		} else {
			clientStatus.tunnelType = "quic"
		}
		clientStatus.userConfigChange = true
		logger.Debug("Protocol change to ", clientStatus.tunnelType)
	})
	protoRadioBox.Add(protoRadioTcp)
	protoRadioQuic, err := gtk.RadioButtonNewWithLabelFromWidget(protoRadioTcp, "QUIC")
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio button QUIC.")
	}
	protoRadioBox.Add(protoRadioQuic)
	protoRadioBox.SetHAlign(gtk.ALIGN_CENTER)

	// plugin choose
	pluginLabel, err := gtk.LabelNew("Plugin")
	if err != nil {
		logger.WithError(err).Fatal("Could not create plugin label.")
	}
	pluginLabel.SetHAlign(gtk.ALIGN_START)

	pluginRadioBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 50)
	if err != nil {
		logger.WithError(err).Fatal("Could not create plugin radio box.")
	}
	pluginRadioOff, err := gtk.RadioButtonNewWithLabelFromWidget(nil, "Off")
	if err != nil {
		logger.WithError(err).Fatal("Could not create plugin radio button Off.")
	}
	pluginRadioOff.Connect("toggled", func(r *gtk.RadioButton) {
		if r.GetActive() {
			clientStatus.pluginNum = 0
			clientStatus.userConfigChange = true
			logger.Debug("Plugin change to 0")
		}
	})
	pluginRadioBox.Add(pluginRadioOff)
	pluginRadio123, err := gtk.RadioButtonNewWithLabelFromWidget(pluginRadioOff, "th123")
	if err != nil {
		logger.WithError(err).Fatal("Could not create plugin radio button 123.")
	}
	pluginRadio123.Connect("toggled", func(r *gtk.RadioButton) {
		if r.GetActive() {
			clientStatus.pluginNum = 123
			clientStatus.userConfigChange = true
			logger.Debug("Plugin change to 123")
		}
	})
	pluginRadioBox.Add(pluginRadio123)
	pluginRadioBox.SetHAlign(gtk.ALIGN_CENTER)

	// peer address label
	peerLabel, err := gtk.LabelNew("Peer IP")
	if err != nil {
		logger.WithError(err).Fatal("Could not create peer address label.")
	}
	peerLabel.SetMarginTop(10)
	peerLabel.SetHAlign(gtk.ALIGN_START)

	// peer ip label
	addrLabel, err := gtk.LabelNew("No tunnel established")
	if err != nil {
		logger.WithError(err).Fatal("Could not create peer ip label.")
	}
	addrLabel.SetHExpand(true)

	// peer control button
	ctlBtnBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		logger.WithError(err).Fatal("Could not create peer control button box.")
	}
	connBtn, err := gtk.ButtonNewWithLabel("Connect")
	if err != nil {
		logger.WithError(err).Fatal("Could not create connect button.")
	}
	connBtn.SetHExpand(true)
	connBtn.Connect("clicked", func() {

		if !clientStatus.userConfigChange && clientStatus.client.Serving() {
			logger.Debug("Already serving")
			return
		}

		err := onConfigUpdate()
		if err != nil {
			logger.WithError(err).Error("Connect failed")
			showErrorDialog(appWindow, "Connect failed", err)
			return
		}

		err = clientStatus.client.Connect()
		if err != nil {
			logger.WithError(err).Error("Connect failed")
			showErrorDialog(appWindow, "Connect failed", err)
			return
		}

		addrLabel.SetText(clientStatus.client.PeerHost())

		go func() {
			var err error

			switch clientStatus.pluginNum {
			case 123:
				logger.Info("Append th12.3 hisoutensoku plugin")
				h := client.NewHisoutensoku()
				clientStatus.plugin = h
				err = clientStatus.client.Serve(h.ReadFunc, h.WriteFunc, h.GoroutineFunc)

			default:
				clientStatus.plugin = nil
				err = clientStatus.client.Serve(nil, nil, nil)

			}
			if err != nil {
				logger.WithError(err).Error("Connect failed")
				glib.IdleAdd(func() bool {
					showErrorDialog(appWindow, "Connect failed", err)
					return false
				})
			}

		}()

		logger.Debug("Connect")
	})
	refreshBtn, err := gtk.ButtonNewWithLabel("Refresh")
	if err != nil {
		logger.WithError(err).Fatal("Could not create refresh button.")
	}
	refreshBtn.SetHExpand(true)
	refreshBtn.Connect("clicked", func() {

		clientStatus.client.Close()

		connBtn.Emit("clicked")

		logger.Debug("Refresh")
	})
	copyBtn, err := gtk.ButtonNewWithLabel("  Copy  ")
	if err != nil {
		logger.WithError(err).Fatal("Could not create copy button.")
	}
	copyBtn.SetHExpand(true)
	copyBtn.Connect("clicked", func() {

		if clientStatus.client.Serving() {

			clipBoard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			if err != nil {
				showErrorDialog(appWindow, "Get clipboard error", err)
				return
			}
			addr, err := addrLabel.GetText()
			if err != nil {
				showErrorDialog(appWindow, "Get address text error", err)
				return
			}

			clipBoard.SetText(addr)

			logger.Debug("Copy")

		} else {

			logger.Debug("Nothing to copy")

		}

	})
	ctlBtnBox.Add(connBtn)
	ctlBtnBox.Add(refreshBtn)
	ctlBtnBox.Add(copyBtn)
	ctlBtnBox.SetHAlign(gtk.ALIGN_FILL)
	ctlBtnBox.SetHExpand(true)
	ctlBtnBox.SetMarginTop(10)

	// client status label
	statusLabel, err := gtk.LabelNew("Initializing")
	if err != nil {
		logger.WithError(err).Fatal("Could not create status label.")
	}
	statusLabel.SetMarginTop(10)
	statusLabel.SetHAlign(gtk.ALIGN_START)

	// reset action
	aReset := glib.SimpleActionNew("reset", nil)
	aReset.Connect("activate", func() {
		serverEntry.SetText(client.DefaultServerHost)
		localPortEntry.SetText(strconv.Itoa(client.DefaultLocalPort))
		protoRadioTcp.SetActive(true)
		pluginRadioOff.SetActive(true)
	})
	app.AddAction(aReset)

	// net discover action
	aNetDisc := glib.SimpleActionNew("net-disc", nil)
	aNetDisc.Connect("activate", func() {

		// showNetInfoDialog show network delay info dialog
		showNetInfoDialog := func(infoMap map[int]string) error {
			// prepare data
			sortDelay := make([]int, len(infoMap))
			i := 0
			for k := range infoMap {
				sortDelay[i] = k
				i++
			}
			sort.Ints(sortDelay)

			// setup dialog with button
			dialog, err := gtk.DialogNew()
			if err != nil {
				return err
			}
			dialog.SetIcon(icon)
			dialog.SetTitle("Network discovery")
			btn, err := dialog.AddButton("Close", gtk.RESPONSE_CLOSE)
			if err != nil {
				return err
			}
			btn.Connect("clicked", func() {
				dialog.Destroy()
			})

			infoTreeView, err := gtk.TreeViewNew()
			if err != nil {
				return err
			}

			// setup dialog with TreeView
			dialogBox, err := dialog.GetContentArea()
			if err != nil {
				return err
			}
			dialogBox.Add(infoTreeView)

			// setup TreeView
			cellRenderer, err := gtk.CellRendererTextNew()
			if err != nil {
				return err
			}
			serverColumn, err := gtk.TreeViewColumnNewWithAttribute("Server", cellRenderer, "text", 0)
			if err != nil {
				return err
			}
			delayColumn, err := gtk.TreeViewColumnNewWithAttribute("Delay", cellRenderer, "text", 1)
			if err != nil {
				return err
			}
			infoListStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
			if err != nil {
				return err
			}
			infoTreeView.AppendColumn(serverColumn)
			infoTreeView.AppendColumn(delayColumn)
			infoTreeView.SetModel(infoListStore)
			infoTreeView.Connect("row-activated", func(_ *gtk.TreeView, p *gtk.TreePath, _ *gtk.TreeViewColumn) {

				i := p.GetIndices()[0]
				logger.Debug("Net server selected ", i)

				serverEntry.SetText(infoMap[sortDelay[i]])

				dialog.Destroy()

			})

			// append data
			for _, k := range sortDelay {
				logger.Debug("Append server info ", infoMap[k], " delay ", k)
				iter := infoListStore.Append()
				err = infoListStore.Set(iter, []int{0, 1}, []interface{}{infoMap[k], fmt.Sprintf("%.3fms", float64(k)/1000000)})
				if err != nil {
					return err
				}
			}

			// single selection
			infoSel, err := infoTreeView.GetSelection()
			if err != nil {
				return err
			}
			infoSel.SetMode(gtk.SELECTION_SINGLE)

			dialog.ShowAll()

			return nil
		}

		go func() {
			infoMap, err := client.NetBrokerDelay(client.DefaultServerHost)
			if err != nil {
				logger.WithError(err).Warn("Get network broker delay error")

				userServer, err := serverEntry.GetText()
				if err != nil {
					logger.WithError(err).Warn("Get server entry text error")
					return
				}

				infoMap, err = client.NetBrokerDelay(userServer)

				if err != nil {
					glib.IdleAdd(func() bool {
						showErrorDialog(appWindow, "Net discovery Failed", err)
						return false
					})
				}
			}

			// show net discovery dialog
			infoMapCov := make(map[int]string)
			for k, v := range infoMap {
				infoMapCov[v] = k
			}
			if err == nil {
				logger.Debug("Show net discovery dialog")

				glib.IdleAdd(func() bool {
					err = showNetInfoDialog(infoMapCov)
					if err != nil {
						showErrorDialog(appWindow, "Show info discovery dialog error", err)
					}
					return false
				})

			}
		}()

	})
	app.AddAction(aNetDisc)

	// tunnel status
	aTStatus := glib.SimpleActionNew("t-status", nil)
	aTStatus.Connect("activate", func() {

		showTStatusDialog := func() error {

			// setup dialog with button
			dialog, err := gtk.DialogNew()
			if err != nil {
				return err
			}
			dialog.SetIcon(icon)
			dialog.SetTitle("Tunnel status")
			btn, err := dialog.AddButton("Close", gtk.RESPONSE_CLOSE)
			if err != nil {
				return err
			}
			btn.Connect("clicked", func() {
				dialog.Destroy()
			})

			glg, err := glgo.GlgLineGraphNew()
			if err != nil {
				return err
			}

			// setup dialog with glgLineGraph
			dialogBox, err := dialog.GetContentArea()
			if err != nil {
				return err
			}
			glg.SetHExpand(true)
			glg.SetVExpand(true)
			dialogBox.Add(glg)

			source := glib.TimeoutAdd(1000, func() bool {

				pos := (clientStatus.delayPos + 39) % 40
				glg.GlgLineGraphDataSeriesAddValue(0,
					float64(clientStatus.delay[pos].Nanoseconds())/1000000)

				if clientStatus.pluginDelayShow {
					switch p := clientStatus.plugin.(type) {
					case *client.Hisoutensoku:
						if p.PeerStatus == client.BATTLE {
							glg.GlgLineGraphDataSeriesAddValue(1, float64(p.GetReplayDelay().Nanoseconds())/1000000)
						}
					}
				}

				return true
			})

			dialog.Connect("destroy", func() {
				glib.SourceRemove(source)
			})

			dialog.SetDefaultSize(500, 300)
			dialog.ShowAll()

			glg.GlgLineGraphDataSeriesAdd("Tunnel Delay", "blue")
			glg.GlgLineGraphDataSeriesAdd("Peer Delay", "red")
			pos, l := clientStatus.delayPos, clientStatus.delayLen
			if l == 40 {
				for i := pos; i < 40; i++ {
					glg.GlgLineGraphDataSeriesAddValue(0, float64(clientStatus.delay[i].Nanoseconds())/1000000)
				}
			}
			for i := 0; i < pos; i++ {
				glg.GlgLineGraphDataSeriesAddValue(0, float64(clientStatus.delay[i].Nanoseconds())/1000000)
			}
			if clientStatus.pluginDelayShow {
				pos, l = clientStatus.pluginDelayPos, clientStatus.pluginDelayLen
				if l == 40 {
					for i := pos; i < 40; i++ {
						glg.GlgLineGraphDataSeriesAddValue(1, float64(clientStatus.pluginDelay[i].Nanoseconds())/1000000)
					}
				}
				for i := 0; i < pos; i++ {
					glg.GlgLineGraphDataSeriesAddValue(1, float64(clientStatus.pluginDelay[i].Nanoseconds())/1000000)
				}
			}

			return nil
		}

		err = showTStatusDialog()
		if err != nil {
			showErrorDialog(appWindow, "Show tunnel status dialog error", err)
		}
	})
	app.AddAction(aTStatus)

	// add items to grid
	mainGrid.Add(serverLabel)
	mainGrid.Add(serverEntry)
	mainGrid.Add(pingBox)
	mainGrid.Add(setupLabel)
	mainGrid.Add(localPortBox)
	mainGrid.Add(protoRadioBox)
	mainGrid.Add(pluginLabel)
	mainGrid.Add(pluginRadioBox)
	mainGrid.Add(peerLabel)
	mainGrid.Add(addrLabel)
	mainGrid.Add(ctlBtnBox)
	mainGrid.Add(statusLabel)

	// tray icon
	onStatusIconSetup(appWindow)

	appWindow.SetTitlebar(header)
	if icon != nil {
		appWindow.SetIcon(icon)
	}
	appWindow.Add(mainGrid)
	appWindow.SetResizable(false)
	appWindow.SetDefaultSize(320, 450)
	appWindow.SetShowMenubar(true)
	appWindow.ShowAll()

	// auto update ping
	go func() {
		for {
			time.Sleep(time.Second * 2)

			glib.IdleAdd(func() bool {
				if clientStatus.client.Serving() {
					setPingLabel(clientStatus.client.TunnelDelay())
				} else {
					setPingLabel(clientStatus.client.Ping())
				}
				return false
			})
		}
	}()

	// auto update status label
	go func() {
		for {

			if clientStatus.client.Serving() {
				if clientStatus.plugin != nil {
					switch p := clientStatus.plugin.(type) {
					case *client.Hisoutensoku:
						switch p.PeerStatus {
						case client.SUCCESS:
							glib.IdleAdd(func() bool {
								statusLabel.SetText("th12.3 game loaded")
								return false
							})
						case client.BATTLE:
							glib.IdleAdd(func() bool {
								delay := float64(p.GetReplayDelay().Nanoseconds()) / 1000000
								if delay > 9999 {
									delay = 9999
								}
								statusLabel.SetText("th12.3 game ongoing | Delay " +
									fmt.Sprintf("%.2f ms", delay))
								return false
							})
						case client.BATTLE_WAIT_ANOTHER:
							glib.IdleAdd(func() bool {
								statusLabel.SetText(fmt.Sprintf("th12.3 game waiting | %d spectator(s)", p.GetSpectatorCount()))
								return false
							})
						default:
							glib.IdleAdd(func() bool {
								tv, v := clientStatus.client.Version()
								if tv == clientStatus.brokerTVersion && v == clientStatus.brokerVersion {
									statusLabel.SetText("th12.3 game not started")
								} else {
									statusLabel.SetText("Plugin alert! Server is v" + clientStatus.brokerVersion + "-" + strconv.Itoa(int(clientStatus.brokerTVersion)))
								}
								return false
							})
						}
					}
				} else {
					glib.IdleAdd(func() bool {
						tv, v := clientStatus.client.Version()
						if tv == clientStatus.brokerTVersion && v == clientStatus.brokerVersion {
							statusLabel.SetText("Connected")
						} else {
							statusLabel.SetText("Alert! Server is v" + clientStatus.brokerVersion + "-" + strconv.Itoa(int(clientStatus.brokerTVersion)))
						}
						return false
					})
				}
			} else {
				glib.IdleAdd(func() bool {
					tv, v := clientStatus.client.Version()
					if tv == clientStatus.brokerTVersion && v == clientStatus.brokerVersion {
						statusLabel.SetText("Not connected")
					} else {
						statusLabel.SetText("Alert! Server is v" + clientStatus.brokerVersion + "-" + strconv.Itoa(int(clientStatus.brokerTVersion)))
					}
					return false
				})
			}

			time.Sleep(time.Millisecond * 66)

		}
	}()

	// refresh delay
	_, _ = pingBtn.Emit("clicked")

}

// onAppDestroy close client
func onAppDestroy() {
	setStatusIconHide()
	clientStatus.client.Close()
}

// onConfigUpdate update config and setup new client
func onConfigUpdate() error {

	if clientStatus.client.Serving() {
		clientStatus.client.Close()
	}

	if clientStatus.userConfigChange {
		newClient, err := client.New(clientStatus.localPort, clientStatus.serverHost, clientStatus.tunnelType)
		if err != nil {
			return err
		}
		clientStatus.client = newClient
		clientStatus.brokerTVersion, clientStatus.brokerVersion = clientStatus.client.BrokerVersion()
		logger.Debugf("New client %d %s %s", clientStatus.localPort, clientStatus.serverHost, clientStatus.tunnelType)
		clientStatus.userConfigChange = false
	}

	return nil
}

// onUpdatePing send ping via client
func onUpdatePing() time.Duration {
	return clientStatus.client.Ping()
}

// showErrorDialog show error dialog
func showErrorDialog(appWin *gtk.ApplicationWindow, msg string, err error) {
	dialog := gtk.MessageDialogNew(appWin, gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "%s", err)
	dialog.SetTitle(msg)
	dialog.Show()
	dialog.Connect("response", func() {
		dialog.Destroy()
	})
}

// showAboutDialog show about dialog
func showAboutDialog() {

	about, err := gtk.AboutDialogNew()
	if err != nil {
		logger.WithError(err).Error("Show about dialog error")
	}
	about.SetProgramName(appName)
	about.SetVersion("Client Version " + utils.Version + " Tunnel Version " + strconv.Itoa(int(utils.TunnelVersion)))
	about.SetAuthors([]string{"桜風の狐"})
	about.SetCopyright("https://github.com/gotk3/gotk3 ISC License\n" +
		"https://github.com/lucas-clemente/quic-go MIT License\n" +
		"https://github.com/sirupsen/logrus MIT License\n" +
		"https://github.com/weilinfox/youmu-thlink/glg-go GPL-3.0 License\n" +
		"https://github.com/weilinfox/youmu-thlink AGPL-3.0 License\n" +
		"\n2022-2023 weilinfox")
	about.SetTitle("About ThLink")
	if icon != nil {
		about.SetIcon(icon)
		about.SetLogo(icon)
	}

	about.Show()
	about.Connect("response", func() {
		about.Destroy()
	})

}
