package protocbuf

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/eolinker/eosc"
)

func TestParseFile(t *testing.T) {
	var data = `[
            {
                "data":"H4sIAAAAAAAAA12OMU4DMRBF+z3FKAfAYrdcbU8Noo7MMjIma4/xeFEQooAqQkRCAkTJDQg0AUUch7XouAJeSESgfX/mv88nNsgxVDBwngIVgzLLyAVNFhQNnaxHUmEfKx0O2r2NmoxAarQdoRfSaRtI4Fga16BQ3tViHw0NGf2xrrEvW1Xs/KBEhIAtbBraxqMWOcDH7CU+X2QGmfvDP9lpBsDBa6vAStMP2SwTStoiB9uaKi9/LzTvJnFVlNnZuoUdWUbobqbd4vbz7SpOrrvLhzifxPNZvJ8nYli9vy4OmWy8e4rTx/9blg1rY9LHasuSoPeJ5N/uL+SfHy9VAQAA",
                "name":"msg.proto",
                "size":341,
                "type":"text/plain"
            },
            {
                "data":"H4sIAAAAAAAAA53OvQrCMBAA4L1PcXTSxQyO4uDmbB+gxHjEYJKLyVWU0ne3NhERFMTl4P6+u3TzLK+whjpEYlrWq6qiwIY8aGqDVCep8dHWho/dfqHICSRr/AmjkMF4JoFX6YJFoWNQ4oCO2oTxYhSOmHGBIkPtkl5MFx4HnmyTx8ZKWYAtWkvQVwCjlbPZFHd47jDxHCJyF32CZzkF8gnn/bAqSw1HlK7Mz9KUwZ9Gbnz74N0uCLyUjbUZ+vjFD8xwB3HHA1WeAQAA",
                "name":"service.proto",
                "size":414,
                "type":"text/plain"
            }
        ]`
	fileData := make(eosc.EoFiles, 0)
	err := json.Unmarshal([]byte(data), &fileData)
	if err != nil {
		return
	}

	desc, err := parseFiles(fileData)
	fmt.Println(desc, err)
}
