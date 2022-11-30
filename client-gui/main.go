package main

import (
	"fmt"
	client "github.com/weilinfox/youmu-thlink/client/lib"
	"github.com/weilinfox/youmu-thlink/utils"
	"os"
	"strconv"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client-gui", "internal")

type Status struct {
	localPort  int
	serverHost string
	tunnelType string

	userConfigChange bool

	client *client.Client
}

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

	app.Connect("activate", onAppActivate)

	// go client.Main(10800, "thlink.inuyasha.love", 4646, 't')

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
		app.Quit()
	})
	app.AddAction(aQuit)
	aAbout := glib.SimpleActionNew("about", nil)
	aAbout.Connect("activate", func() {

	})
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
	menu.Append("Info", "app.info")
	menu.Append("About", "app.about")
	menu.Append("Quit", "app.quit")
	menuBtn.SetMenuModel(&menu.MenuModel)
	header.PackStart(menuBtn)
	header.SetShowCloseButton(true)
	header.SetTitle("白玉楼製作所 ThLink")
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
	pingBtn.Connect("clicked", func() {
		err := onConfigUpdate()
		if err != nil {
			logger.WithError(err).Error("Update ping failed")
			dialog := gtk.MessageDialogNew(nil, gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "Update ping failed: %e", err)
			dialog.Show()
			dialog.Destroy()
			return
		}

		delay := onUpdatePing()
		logger.Debugf("Get new delay %.2f ms", float64(delay.Nanoseconds())/1000000)
		pingLabel.SetText(fmt.Sprintf("%.2f ms", float64(delay.Nanoseconds())/1000000))
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
	protoRadio, err := gtk.RadioButtonNewWithLabelFromWidget(nil, "TCP")
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio button TCP.")
	}
	protoRadio.Connect("toggled", func(r *gtk.RadioButton) {
		if r.GetActive() {
			clientStatus.tunnelType = "tcp"
		} else {
			clientStatus.tunnelType = "quic"
		}
		clientStatus.userConfigChange = true
		logger.Debug("Protocol change to ", clientStatus.tunnelType)
	})
	protoRadioBox.Add(protoRadio)
	protoRadio, err = gtk.RadioButtonNewWithLabelFromWidget(protoRadio, "QUIC")
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio button QUIC.")
	}
	protoRadioBox.Add(protoRadio)
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
		err := onConfigUpdate()
		if err != nil {
			logger.WithError(err).Error("Connect failed")
			dialog := gtk.MessageDialogNew(nil, gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "Connect failed: %e", err)
			dialog.Show()
			dialog.Destroy()
			return
		}

		logger.Debug("Connect")
	})
	refreshBtn, err := gtk.ButtonNewWithLabel("Refresh")
	if err != nil {
		logger.WithError(err).Fatal("Could not create refresh button.")
	}
	refreshBtn.SetHExpand(true)
	refreshBtn.Connect("clicked", func() {
		err := onConfigUpdate()
		if err != nil {
			logger.WithError(err).Error("Refresh failed")
			dialog := gtk.MessageDialogNew(nil, gtk.DIALOG_DESTROY_WITH_PARENT|gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "Refresh failed: %e", err)
			dialog.Show()
			dialog.Destroy()
			return
		}

		logger.Debug("Refresh")
	})
	copyBtn, err := gtk.ButtonNewWithLabel("  Copy  ")
	if err != nil {
		logger.WithError(err).Fatal("Could not create copy button.")
	}
	copyBtn.SetHExpand(true)
	copyBtn.Connect("clicked", func() {
		logger.Debug("Copy")
	})
	ctlBtnBox.Add(connBtn)
	ctlBtnBox.Add(refreshBtn)
	ctlBtnBox.Add(copyBtn)
	ctlBtnBox.SetHAlign(gtk.ALIGN_FILL)
	ctlBtnBox.SetHExpand(true)
	ctlBtnBox.SetMarginTop(10)

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
	appWindow.Add(mainGrid)
	appWindow.SetResizable(false)
	appWindow.SetDefaultSize(240, 320)
	appWindow.SetShowMenubar(true)
	appWindow.ShowAll()

}

func onAppDestroy(appWin *gtk.ApplicationWindow) {
}

func onConfigUpdate() error {
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

func onUpdatePing() time.Duration {
	return clientStatus.client.Ping()
}
