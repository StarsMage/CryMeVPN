package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	_ "embed"

	"CryMeVPN/pkg/core"
	"CryMeVPN/pkg/deployer"
	"CryMeVPN/pkg/sshclient"
	"CryMeVPN/pkg/storage"
	"CryMeVPN/pkg/sysproxy"
	"CryMeVPN/pkg/xraycore"
	"CryMeVPN/pkg/xraygen"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	goruntime "runtime"
)

//go:embed build/icon.ico
var icon []byte

var domainRe = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

const repoURL = "https://github.com/StarsMage"

const (
	defaultDom = "www.openssl.org"
	defaultDomPort = "443"
)


type App struct {
	ctx context.Context
	manager xraycore.XrayManage
	appstorage storage.AppStorage
	key []byte
}

type ServerView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port uint16 `json:"port"`
}

var allowedDNS  = map[string]bool {
	xraygen.CloudFlare: true,
	xraygen.AdGuard: true,
	xraygen.Quad9: true,
}

var allowedFingerPrint = map[string]bool {
	xraygen.FingerChrome: true,
	xraygen.FingerFirefox: true,
	xraygen.FingerEdge: true,
	xraygen.FingerSafari:true,
}

func validLocalPort(httpPort ,socksPort string) (uint16, uint16, error) {
	socksPort = strings.TrimSpace(socksPort)
	httpPort = strings.TrimSpace(httpPort)

	if socksPort == "" {
		socksPort = strconv.Itoa(int(xraygen.LocalSocksPort))
	}

	if httpPort == "" {
		httpPort = strconv.Itoa(int(xraygen.LocalHTTPPort))
	}

	h, err := strconv.ParseUint(httpPort, 10, 16)

	if err != nil || h < 1024 {
		return 0, 0, fmt.Errorf("invaLID local http port")
	}

	s, err := strconv.ParseUint(socksPort,10,16)
	if err != nil || s < 1024 {
		return 0, 0, fmt.Errorf("invaLID local socks port")
	}

	if h == s {
		return 0,0, fmt.Errorf("http and socks ports must be different")
	}

	return uint16(h),uint16(s),nil
}

func validFindDns(dns, fing string) (string, string, error) {
	dns = strings.TrimSpace(dns)
	fing = strings.ToLower(strings.TrimSpace(fing))

	if dns == "" {
		dns = xraygen.CloudFlare
	}

	if fing == "" {
		fing = xraygen.FingerFirefox
	}

	if !allowedDNS[dns] {
		return "","",fmt.Errorf("unsupp dns: %s",dns)
	}

	if !allowedFingerPrint[fing] {
		return "","",fmt.Errorf("unsupp fing: %s",fing)
	}

	return dns,fing,nil
}

func validDomain(domain, port string) (string, string, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	port = strings.TrimSpace(port)

	if domain == "" {
		domain = defaultDom
	}

	if port == "" {
		port = defaultDomPort
	}

	if len(domain) > 253 || !domainRe.MatchString(domain) {
		return "", "", fmt.Errorf("invalid domain: %q",domain)
	}

	p, err := strconv.Atoi(port)

	if err != nil || p < 1 || p > 65535 {
		return "", "", fmt.Errorf("invalid domain port: %q",port)
	}

	return  domain, net.JoinHostPort(domain, port), nil
}
func (a *App) SysTray() {
	systray.SetIcon(icon)
	systray.SetTitle("Cry")

	githunLink := systray.AddMenuItem("Maybe you could check my GitHub?", "git no more")
	systray.AddSeparator()
	openTray := systray.AddMenuItem("open","open app")
	closeTray := systray.AddMenuItem("close","close app")
	systray.AddSeparator()


	go func() {
		for {
			select {
			case <- openTray.ClickedCh:
				runtime.WindowUnminimise(a.ctx)
				runtime.WindowShow(a.ctx)
			case <- closeTray.ClickedCh:
				runtime.Quit(a.ctx)
				return
			case <- githunLink.ClickedCh:
				runtime.BrowserOpenURL(a.ctx, repoURL)
			}
		}
	}()
}

func (a *App) emit(event string, data ...interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data...)
	}
}

func NewApp() *App {
	key,err := storage.GetOnCreateKey()
	if err != nil {
		return &App{}
	}

	loadedServ, err := storage.Load([]byte(key))

	if err != nil || loadedServ == nil {
		loadedServ = &storage.AppStorage{Servers: []storage.ServerData{}}
	}
	return &App{
		manager: *xraycore.NewXrayManage(),
		appstorage: *loadedServ,
		key: key,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	if goruntime.GOOS == "windows" { go systray.Run(a.SysTray,func() {}) }
}

func (a *App) OnSecondInstance(data options.SecondInstanceData) {
	if a.ctx == nil {
		return
	}

	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
} 

func (a *App) Shutdown(ctx context.Context) {
	a.StopCore()
	systray.Quit()
}

func (a *App) StartCore(id string) error {
	var conf string
	var httpPort uint16
	found := false

	for _,s := range a.appstorage.Servers {
		if s.ID == id {
			found = true
			conf = s.ClientConfig
			httpPort = s.HTTPPort
			break
		}
	}

	if !found {
		return fmt.Errorf("server not found: %q",id)
	}

	if conf == "" {
		return fmt.Errorf("server conf not found: %q",id)
	}

	if err := a.BinaryXray(); err != nil {
		return err
	}

	if err := storage.SaveConfig(conf); err != nil {
		return err
	}

	if err := a.manager.StartXray(); err != nil {
		return err
	}


	if httpPort == 0 {
		httpPort = xraygen.LocalHTTPPort
	}
	proxAddr := fmt.Sprintf("127.0.0.128:%d",httpPort)

	if err := sysproxy.Enable(proxAddr); err != nil {
		if !errors.Is(err,sysproxy.ErrUnsupported) {
			a.manager.Stop()
			return err
		}
	}

	return nil
}

func (a *App) StopCore() error {
	sysErr := sysproxy.Disable()
	managErr := a.manager.Stop()

	if sysErr != nil {
		return sysErr
	}
	
	return managErr
}

func (a *App) DeployServer(ip,sshPort,port,user,pass,domain,domainPort, dns, fingerprint, localHttp, localSocks string) error {
	localHttpC, localSocksC, err := validLocalPort(localHttp, localSocks)

	if err != nil {
		return err
	}

	dnsC, fing, err := validFindDns(dns, fingerprint)

	if err != nil {
		return err
	}

	domainC, destC, err := validDomain(domain, domainPort)
	if err != nil {
		return err
	}

	p64, err := strconv.ParseUint(port,10,16)
	if err != nil {
		return err
	}
	client, err := sshclient.Connect(ip,sshPort, user,pass)

	if err != nil {
		return err
	}

	defer client.Close()
	cred, err := xraygen.GenerateCredentials()
	if err != nil {
		return err
	}


	cfgStruct := xraygen.ConfigStruct{
		Port: uint16(p64),
		Destination: destC,
		ServerName:  domainC,
		DNS: dnsC,
		UUID: cred.UUID,
		PrivateKey: cred.PrivateKey,
		ShortID: cred.ShortID,
	}

	jsonTemp, err := xraygen.GenerateConfigOnBaseTemplate(cfgStruct)

	if err != nil {
		return err
	}

	err = deployer.DeployAndConfigureXray(client, jsonTemp, port, ip)

	if err != nil {
		return err
	}

	clientJson, err := xraygen.GenerateClientConfigOnBaseTemplate(xraygen.ClientStruct{
		Address: ip,
		Port: uint16(p64),
		UUID:        cred.UUID,
		Flow:        xraygen.FlowVision,
		SNI:         domainC,
		Fingerprint: fing,
		PublicKey:   cred.OpenKey,
		ShortID:     cred.ShortID,
		SocksPort:   localSocksC,
		HTTPPort:    localHttpC,
	})

	if err != nil {
		return err
	}

	ipBase64 := strings.ToLower(base64.StdEncoding.EncodeToString([]byte(ip)))
	for len(ipBase64) < 10{ipBase64 += "x"}
	serverRem := fmt.Sprintf("C%sR%sY%s",string(ipBase64[0]),string(ipBase64[1]),string(ipBase64[9]))
	serverHashName := deployer.GenerateNameServer(ip)

	newServ := storage.NewServer(
		serverHashName,
		serverRem,
		ip,
		clientJson,
		cred.PrivateKey,
		uint16(p64),
	)

	newServ.HTTPPort = localHttpC

	a.appstorage.DeleteServer(serverHashName)
	a.appstorage.Servers = append(a.appstorage.Servers, *newServ)

	return a.appstorage.Save(a.key)
}

func(a *App) GetSavedServer() ([]ServerView, error) {
	out := make([]ServerView, 0, len(a.appstorage.Servers))

	for _, v := range a.appstorage.Servers {
		out = append(out, ServerView{v.ID,v.Name,v.ServerIp,v.Port})
	}
	return out,nil
}

func (a *App) DeleteServer(id string) error {
	a.appstorage.DeleteServer(id)
	return a.appstorage.Save(a.key)
}

func (a *App) BinaryXray() error {
	path, err := core.GetXrayPath()

	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		asset, err := core.GetLastXrayAsset()

		if err != nil {
			return err
		}

		zipPath := filepath.Join(filepath.Dir(path),"xray_manager.zip")
		defer os.Remove(zipPath)

		a.emit("xray:status", "Downloading...")
		if err = core.DownloadXray(asset.URL,zipPath); err != nil {
			a.emit("xray:status","")
			return err
		}
		a.emit("xray:status","Validation sha256...")
		if err = core.Sha256Validation(zipPath,asset.Digest); err != nil {
			a.emit("xray:status","")
			return err
		}
		binDir := filepath.Dir(path)

		a.emit("xray:status","Unpacking...")
		if err = core.ExtractXray(zipPath,binDir); err != nil {
			a.emit("xray:status","")
			return err
		
		}

		a.emit("xray:status", "") 
	}
	return nil
}