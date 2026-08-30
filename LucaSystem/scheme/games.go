package scheme

type Game interface {
	ExportScript(name string, extra []byte, payload []byte) (string, bool, error)
}

var Games = map[string]func() Game{
	"dreamher": func() Game { return &DreamHerVM{} },
	"usonatsu": func() Game { return &usonatsuVM{} },
	"sprb":     func() Game { return &SPRBVM{} },
}
