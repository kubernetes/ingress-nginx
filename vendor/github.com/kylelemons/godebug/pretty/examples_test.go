// Copyright 2013 Google Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pretty_test

import (
	"fmt"
	"net"
	"reflect"

	"github.com/kylelemons/godebug/pretty"
)

func ExampleConfig_Sprint() {
	type Pair [2]int
	type Map struct {
		Name      string
		Players   map[string]Pair
		Obstacles map[Pair]string
	}

	m := Map{
		Name: "Rock Creek",
		Players: map[string]Pair{
			"player1": {1, 3},
			"player2": {0, -1},
		},
		Obstacles: map[Pair]string{
			Pair{0, 0}: "rock",
			Pair{2, 1}: "pond",
			Pair{1, 1}: "stream",
			Pair{0, 1}: "stream",
		},
	}

	// Specific output formats
	compact := &pretty.Config{
		Compact: true,
	}
	diffable := &pretty.Config{
		Diffable: true,
	}

	// Print out a summary
	fmt.Printf("Players: %s\n", compact.Sprint(m.Players))

	// Print diffable output
	fmt.Printf("Map State:\n%s", diffable.Sprint(m))

	// Output:
	// Players: {player1:[1,3],player2:[0,-1]}
	// Map State:
	// {
	//  Name: "Rock Creek",
	//  Players: {
	//   player1: [
	//    1,
	//    3,
	//   ],
	//   player2: [
	//    0,
	//    -1,
	//   ],
	//  },
	//  Obstacles: {
	//   [0,0]: "rock",
	//   [0,1]: "stream",
	//   [1,1]: "stream",
	//   [2,1]: "pond",
	//  },
	// }
}

func ExampleConfig_fmtFormatter() {
	pretty.DefaultFormatter[reflect.TypeOf(&net.IPNet{})] = fmt.Sprint
	pretty.DefaultFormatter[reflect.TypeOf(net.HardwareAddr{})] = fmt.Sprint
	pretty.Print(&net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.CIDRMask(24, 32),
	})
	pretty.Print(net.HardwareAddr{1, 2, 3, 4, 5, 6})

	// Output:
	// 192.168.1.100/24
	// 01:02:03:04:05:06
}

func ExampleConfig_customFormatter() {
	pretty.DefaultFormatter[reflect.TypeOf(&net.IPNet{})] = func(n *net.IPNet) string {
		return fmt.Sprintf("CIDR=%s", n)
	}
	pretty.Print(&net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.CIDRMask(24, 32),
	})

	// Output:
	// CIDR=192.168.1.100/24
}

func ExamplePrint() {
	type ShipManifest struct {
		Name     string
		Crew     map[string]string
		Androids int
		Stolen   bool
	}

	manifest := &ShipManifest{
		Name: "Spaceship Heart of Gold",
		Crew: map[string]string{
			"Zaphod Beeblebrox": "Galactic President",
			"Trillian":          "Human",
			"Ford Prefect":      "A Hoopy Frood",
			"Arthur Dent":       "Along for the Ride",
		},
		Androids: 1,
		Stolen:   true,
	}

	pretty.Print(manifest)

	// Output:
	// {Name:     "Spaceship Heart of Gold",
	//  Crew:     {Arthur Dent:       "Along for the Ride",
	//             Ford Prefect:      "A Hoopy Frood",
	//             Trillian:          "Human",
	//             Zaphod Beeblebrox: "Galactic President"},
	//  Androids: 1,
	//  Stolen:   true}
}

var t = struct {
	Errorf func(string, ...interface{})
}{
	Errorf: func(format string, args ...interface{}) {
		fmt.Println(fmt.Sprintf(format, args...) + "\n")
	},
}

func ExampleCompare_testing() {
	// Code under test:

	type ShipManifest struct {
		Name     string
		Crew     map[string]string
		Androids int
		Stolen   bool
	}

	// AddCrew tries to add the given crewmember to the manifest.
	AddCrew := func(m *ShipManifest, name, title string) {
		if m.Crew == nil {
			m.Crew = make(map[string]string)
		}
		m.Crew[title] = name
	}

	// Test function:
	tests := []struct {
		desc        string
		before      *ShipManifest
		name, title string
		after       *ShipManifest
	}{
		{
			desc:   "add first",
			before: &ShipManifest{},
			name:   "Zaphod Beeblebrox",
			title:  "Galactic President",
			after: &ShipManifest{
				Crew: map[string]string{
					"Zaphod Beeblebrox": "Galactic President",
				},
			},
		},
		{
			desc: "add another",
			before: &ShipManifest{
				Crew: map[string]string{
					"Zaphod Beeblebrox": "Galactic President",
				},
			},
			name:  "Trillian",
			title: "Human",
			after: &ShipManifest{
				Crew: map[string]string{
					"Zaphod Beeblebrox": "Galactic President",
					"Trillian":          "Human",
				},
			},
		},
		{
			desc: "overwrite",
			before: &ShipManifest{
				Crew: map[string]string{
					"Zaphod Beeblebrox": "Galactic President",
				},
			},
			name:  "Zaphod Beeblebrox",
			title: "Just this guy, you know?",
			after: &ShipManifest{
				Crew: map[string]string{
					"Zaphod Beeblebrox": "Just this guy, you know?",
				},
			},
		},
	}

	for _, test := range tests {
		AddCrew(test.before, test.name, test.title)
		if diff := pretty.Compare(test.before, test.after); diff != "" {
			t.Errorf("%s: post-AddCrew diff: (-got +want)\n%s", test.desc, diff)
		}
	}

	// Output:
	// add first: post-AddCrew diff: (-got +want)
	//  {
	//   Name: "",
	//   Crew: {
	// -  Galactic President: "Zaphod Beeblebrox",
	// +  Zaphod Beeblebrox: "Galactic President",
	//   },
	//   Androids: 0,
	//   Stolen: false,
	//  }
	//
	// add another: post-AddCrew diff: (-got +want)
	//  {
	//   Name: "",
	//   Crew: {
	// -  Human: "Trillian",
	// +  Trillian: "Human",
	//    Zaphod Beeblebrox: "Galactic President",
	//   },
	//   Androids: 0,
	//   Stolen: false,
	//  }
	//
	// overwrite: post-AddCrew diff: (-got +want)
	//  {
	//   Name: "",
	//   Crew: {
	// -  Just this guy, you know?: "Zaphod Beeblebrox",
	// -  Zaphod Beeblebrox: "Galactic President",
	// +  Zaphod Beeblebrox: "Just this guy, you know?",
	//   },
	//   Androids: 0,
	//   Stolen: false,
	//  }
}

func ExampleCompare_debugging() {
	type ShipManifest struct {
		Name     string
		Crew     map[string]string
		Androids int
		Stolen   bool
	}

	reported := &ShipManifest{
		Name: "Spaceship Heart of Gold",
		Crew: map[string]string{
			"Zaphod Beeblebrox": "Galactic President",
			"Trillian":          "Human",
			"Ford Prefect":      "A Hoopy Frood",
			"Arthur Dent":       "Along for the Ride",
		},
		Androids: 1,
		Stolen:   true,
	}

	expected := &ShipManifest{
		Name: "Spaceship Heart of Gold",
		Crew: map[string]string{
			"Trillian":      "Human",
			"Rowan Artosok": "Captain",
		},
		Androids: 1,
		Stolen:   false,
	}

	fmt.Println(pretty.Compare(reported, expected))
	// Output:
	//  {
	//   Name: "Spaceship Heart of Gold",
	//   Crew: {
	// -  Arthur Dent: "Along for the Ride",
	// -  Ford Prefect: "A Hoopy Frood",
	// +  Rowan Artosok: "Captain",
	//    Trillian: "Human",
	// -  Zaphod Beeblebrox: "Galactic President",
	//   },
	//   Androids: 1,
	// - Stolen: true,
	// + Stolen: false,
	//  }
}
