package main

import (
	client "github.com/weilinfox/youmu-thlink/client/lib"
	"github.com/weilinfox/youmu-thlink/utils"
	"os"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("client-gui", "internal")

type Config struct {
	LocalPort  int
	ServerHost string
	TunnelType string
}

var config = Config{
	LocalPort:  client.DefaultLocalPort,
	ServerHost: client.DefaultServerHost,
	TunnelType: client.DefaultTunnelType,
}

func main() {

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
	serverEntryBuf, err := gtk.EntryBufferNew(config.ServerHost, len(config.ServerHost))
	if err != nil {
		logger.WithError(err).Fatal("Could not create server entry buffer.")
	}
	serverEntry, err := gtk.EntryNewWithBuffer(serverEntryBuf)
	if err != nil {
		logger.WithError(err).Fatal("Could not create server entry.")
	}
	serverEntry.SetHExpand(true)

	// ping button
	pingBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	pingBtn, err := gtk.ButtonNewWithLabel("Ping")
	if err != nil {
		logger.WithError(err).Fatal("Could not create ping button.")
	}
	pingBtn.SetHExpand(true)
	pingLabel, err := gtk.LabelNew("Null delay")
	if err != nil {
		logger.WithError(err).Fatal("Could not create delay label.")
	}
	pingLabel.SetHExpand(true)
	pingBox.Add(pingLabel)
	pingBox.Add(pingBtn)

	// setup label
	setupLabel, err := gtk.LabelNew("Tunnel protocol")
	if err != nil {
		logger.WithError(err).Fatal("Could not create setup label.")
	}
	setupLabel.SetHAlign(gtk.ALIGN_START)

	// protocol choose
	protoRadioBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 50)
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio box.")
	}
	protoRadio, err := gtk.RadioButtonNewWithLabelFromWidget(nil, "TCP")
	if err != nil {
		logger.WithError(err).Fatal("Could not create protocol radio button TCP.")
	}
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
	refreshBtn, err := gtk.ButtonNewWithLabel("Refresh")
	if err != nil {
		logger.WithError(err).Fatal("Could not create refresh button.")
	}
	refreshBtn.SetHExpand(true)
	copyBtn, err := gtk.ButtonNewWithLabel("  Copy  ")
	if err != nil {
		logger.WithError(err).Fatal("Could not create copy button.")
	}
	copyBtn.SetHExpand(true)
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
