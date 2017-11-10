package gamedef

import (
	"github.com/davyxu/goobjfmt"
	"github.com/mutousay/cellnet"
	_ "github.com/mutousay/cellnet/codec/json"
	"github.com/mutousay/cellnet/util"
	"reflect"
)

type TestEchoJsonACK struct {
	Content string
}

func (m *TestEchoJsonACK) String() string { return goobjfmt.CompactTextString(m) }

func init() {

	// coredef.proto
	cellnet.RegisterMessageMeta("json", "gamedef.TestEchoJsonACK", reflect.TypeOf((*TestEchoJsonACK)(nil)).Elem(), util.StringHash("gamedef.TestEchoJsonACK"))
}
