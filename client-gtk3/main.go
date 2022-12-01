package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	client "github.com/weilinfox/youmu-thlink/client/lib"
	"github.com/weilinfox/youmu-thlink/utils"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client-gui", "internal")
var icon *gdk.Pixbuf

const appName = "白玉楼製作所 ThLink"

type Status struct {
	localPort  int
	serverHost string
	tunnelType string

	userConfigChange bool

	client *client.Client
}

// clientStatus all of this client-gui
var clientStatus = Status{
	localPort:        client.DefaultLocalPort,
	serverHost:       client.DefaultServerHost,
	tunnelType:       client.DefaultTunnelType,
	client:           client.NewWithDefault(),
	userConfigChange: false,
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

// onAppActivate setup Main window
func onAppActivate(app *gtk.Application) {

	appWindow, err := gtk.ApplicationWindowNew(app)
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
	aInfo := glib.SimpleActionNew("info", nil)
	aInfo.Connect("activate", func() {

	})
	app.AddAction(aInfo)

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
	menu.Append("Tunnel info", "app.info")
	menu.Append("About thlink", "app.about")
	menu.Append("Quit", "app.quit")
	menuBtn.SetMenuModel(&menu.MenuModel)
	header.PackStart(menuBtn)
	header.SetShowCloseButton(true)
	header.SetTitle(appName)
	header.SetSubtitle("v" + utils.Version)

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

	// peer address label
	peerLabel, err := gtk.LabelNew("Peer IP")
	if err != nil {
		logger.WithError(err).Fatal("Could not create button.")
	}
	peerLabel.SetMarginTop(10)
	peerLabel.SetHAlign(gtk.ALIGN_START)

	// peer ip label
	addrLabel, err := gtk.LabelNew("No tunnel established")
	if err != nil {
		logger.WithError(err).Fatal("Could not create button.")
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
			err := clientStatus.client.Serve()
			if err != nil {
				logger.WithError(err).Error("Connect failed")
				showErrorDialog(appWindow, "Connect failed", err)
				return
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

	// reset action
	aReset := glib.SimpleActionNew("reset", nil)
	aReset.Connect("activate", func() {
		serverEntry.SetText(client.DefaultServerHost)
		localPortEntry.SetText(strconv.Itoa(client.DefaultLocalPort))
		protoRadioTcp.SetActive(true)
	})
	app.AddAction(aReset)

	// add items to grid
	mainGrid.Add(serverLabel)
	mainGrid.Add(serverEntry)
	mainGrid.Add(pingBox)
	mainGrid.Add(setupLabel)
	mainGrid.Add(localPortBox)
	mainGrid.Add(protoRadioBox)
	mainGrid.Add(peerLabel)
	mainGrid.Add(addrLabel)
	mainGrid.Add(ctlBtnBox)

	appWindow.SetTitlebar(header)
	if icon != nil {
		appWindow.SetIcon(icon)
	}
	appWindow.Add(mainGrid)
	appWindow.SetResizable(false)
	appWindow.SetDefaultSize(240, 320)
	appWindow.SetShowMenubar(true)
	appWindow.ShowAll()

	// auto update ping
	go func() {
		for {
			time.Sleep(time.Second * 2)
			if clientStatus.client.Serving() {
				setPingLabel(clientStatus.client.TunnelDelay())
			} else {
				setPingLabel(clientStatus.client.Ping())
			}
		}
	}()

	// refresh delay
	_, _ = pingBtn.Emit("clicked")

}

// onAppDestroy close client
func onAppDestroy() {
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
	about.SetVersion("Version " + utils.Version)
	about.SetAuthors([]string{"桜風の狐"})
	about.SetCopyright("https://github.com/gotk3/gotk3 ISC License\n" +
		"https://github.com/lucas-clemente/quic-go MIT License\n" +
		"https://github.com/sirupsen/logrus MIT License\n" +
		"https://github.com/weilinfox/youmu-thlink AGPL-3.0 License\n" +
		"\n2022 weilinfox")
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
