package library

import (
	"mixerfw/fwpack"
	"mixerfw/plugins/session"
)

type (
	Adding struct {
	}
)

func (a *Adding) InitSessionCarry(c *Mixer.MixerCarry) *session.MixerSessionCarry {
	ccc := c.Get("session")
	if ccc != nil {
		cs := ccc.(*session.MixerSessionCarry)
		cs.Data.Set("stack", "realpath", c.RealPath())
		cs.Data.Set("stack", "command", c.GetParam("command"))
		cs.Data.Set("stack", "id", c.GetParam("id"))
		cs.Data.Set("stack", "path", c.Request().URL.Path)
		return cs

	} else {
		//cs.Session.Logger.Warning("[makeNewHandlerCarry] не найден `SessionObject`\n")
		return nil
	}
}
