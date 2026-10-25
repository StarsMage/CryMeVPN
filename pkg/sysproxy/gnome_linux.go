package sysproxy

import (

	"os/exec"
	"strings"
)

const gnomeSchema = "org.gnome.system.proxy"

type gnome struct {
	prevMode string
	saved    bool
}

func gsettings(args ...string) (string, error) {
	out, err := exec.Command("gsettings", args...).Output()
	return strings.TrimSpace(string(out)), err
}

func (g *gnome) Enable(host, port string) error {
	if !g.saved {
		mode, err := gsettings("get", gnomeSchema, "mode")

		if err != nil {
			return err
		}

		g.prevMode = strings.Trim(mode, "'")
		g.saved = true
	}

	for _, schema := range []string{gnomeSchema + ".http", gnomeSchema + ".https"} {
		if _, err := gsettings("set",schema,"host",host); err != nil {
			return err
		}

		if _, err := gsettings("set",schema,"port",port); err != nil {
			return err
		}
	}

	_, err := gsettings("set",gnomeSchema,"mode","manual")

	return err
}

func (g *gnome) Disable() error {
	mode := "none"
	if g.saved {
		mode = g.prevMode
	}

	if _, err := gsettings("set",gnomeSchema,"mode",mode); err != nil {
		return err
	}
	
	g.saved = false

	return nil
}