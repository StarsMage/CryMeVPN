package sysproxy

import (
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const keyPath = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

var (
	saved      bool
	prevEnable uint64
	prevServer string
)

var (
	wininet				    = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption	= wininet.NewProc("InternetSetOptionW")
)

const (
	optionSettingChanged = 39
	optionRefresh		 = 37
)

func notify() {
	procInternetSetOption.Call(0,optionSettingChanged,0,0)
	procInternetSetOption.Call(0,optionRefresh,0,0)
}

func Enable(addr string) error {
	key, err := registry.OpenKey(registry.CURRENT_USER,keyPath,
		registry.QUERY_VALUE|registry.SET_VALUE)

	if err != nil {
		return err
	}

	defer key.Close()

	if !saved {
		prevEnable, _, _ = key.GetIntegerValue("ProxyEnable")
		prevServer, _, _ = key.GetStringValue("ProxyServer")
		saved = true
	}

	if err := key.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}

	if err := key.SetStringValue("ProxyServer",addr); err != nil {
		return err
	}

	notify()
	return nil
}

func Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath,
		registry.QUERY_VALUE|registry.SET_VALUE)
	
	if err != nil {
		return  err
	}

	defer key.Close()

	if saved {
		if err := key.SetDWordValue("ProxyEnable", uint32(prevEnable)); err != nil {
			return err
		}

		if err := key.SetStringValue("ProxyServer",prevServer); err != nil {
			return err
		}
	}else {
		if err := key.SetDWordValue("ProxyEnable", 0); err != nil {
			return err
		}
	}

	saved = false
	notify()
	return nil 
}