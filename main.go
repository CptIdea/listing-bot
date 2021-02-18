package main

import (
	"fmt"

	"github.com/CptIdea/go-vk-api-2"
)

func main() {
	s := vk.NewSession("4e92464a8b8836c257a214862f0ef8e0cbc036b609369fe640699ded8e6f247941aa53d39a95f24e9775c", "5.130")
	for {
		nUpd, err := s.UpdateCheck(202676872)
		if err != nil {
			fmt.Println(err)
		}
		for _, upd := range nUpd.Updates {
			if upd.Object.MessageNew.Text == "/help" {
				s.SendMessage(upd.Object.MessageNew.PeerId, "Fuck you")
			}
		}
	}
}
