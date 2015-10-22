package tablerunner

import (
	. "github.com/concourse/fly/commands"
	"github.com/fatih/color"
)

func main() {
	table := Table{
		Headers: TableRow{
			{Contents: "column1", Color: color.New(color.Bold)},
			{Contents: "column2", Color: color.New(color.Bold)},
		},
		Data: []TableRow{
			{
				{Contents: "r1c1"},
				{Contents: "r1c2"},
			},
			{
				{Contents: "r2c1"},
				{Contents: "r2c2"},
			},
			{
				{Contents: "r3c1"},
				{Contents: "r3c2"},
			},
		},
	}
	table.Render()
}
