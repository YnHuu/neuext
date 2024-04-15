package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/YnHuu/neuext"
	"log"
	"os"
)

var mDebug = flag.Bool("dev", false, "Enable development mode")
var Ext = new(neuext.Extension)

type Service struct{}

// await Neutralino.extensions.dispatch('js.neutralino.goDevExtension', 'mapExtension', { "id":1 })
func (s *Service) MapExtension(in map[string]any) {
	fmt.Println("MapExtension", in)
}

// await Neutralino.extensions.dispatch('js.neutralino.goDevExtension', 'strExtension', 'string')
func (s *Service) StrExtension(in string) {
	fmt.Println("StrExtension", in)
}

// await Neutralino.extensions.dispatch('js.neutralino.goDevExtension', 'nilExtension')
func (s *Service) NilExtension() {
	fmt.Println("NilExtension")
}

// Neutralino.events.on("mapResult", (e)=>{console.log(e.detail)})
// Neutralino.events.on("strResult", (e)=>{console.log(e.detail)})
func (s *Service) Send() {
	m := make(map[string]any)
	m["id"] = 0
	Ext.Send("mapResult", m)
	Ext.Send("strResult", "string")
}

func main() {
	flag.Parse()
	nl := make(map[string]any)
	if *mDebug {
		Ext.Debug()
		d, _ := os.ReadFile("../.tmp/auth_info.json")
		_ = json.Unmarshal(d, &nl)
		nl["nlExtensionId"] = `js.neutralino.goDevExtension`
	} else {
		decoder := json.NewDecoder(os.Stdin)
		err := decoder.Decode(&nl)
		if err != nil {
			log.Panicln(err)
		}
	}

	if len(nl) != 4 {
		log.Panicln("insufficient parameter")
	}
	Ext.Register(&Service{})
	wsUrl := fmt.Sprintf("ws://localhost:%v?extensionId=%v&connectToken=%v", nl["nlPort"], nl["nlExtensionId"], nl["nlConnectToken"])
	Ext.WSDial(wsUrl, nl["nlToken"].(string))
}
