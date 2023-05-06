package round_robin

type nodeDemo struct {
	label map[string]string
	ip    string
	port  int
	down  bool
}

var testDemos = []struct {
	name  string
	nodes map[string]nodeDemo
	count map[string]int
}{
	{
		name: "权重相等",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo2": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo3": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
		},
		count: map[string]int{
			"demo1": 5,
			"demo2": 5,
			"demo3": 5,
			"demo4": 5,
		},
	},
	{
		name: "权重4:3:2:1",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "40",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo2": {
				label: map[string]string{
					"weight": "30",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo3": {
				label: map[string]string{
					"weight": "20",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
		},
		count: map[string]int{
			"demo1": 8,
			"demo2": 6,
			"demo3": 4,
			"demo4": 2,
		},
	},
	{
		name: "权重4:3:2:1，down调权重40的节点",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "40",
				},
				ip:   "127.0.0.1",
				port: 8580,
				down: true,
			},
			"demo2": {
				label: map[string]string{
					"weight": "30",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo3": {
				label: map[string]string{
					"weight": "20",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
		},
		count: map[string]int{
			"demo1": 1,
			"demo2": 10,
			"demo3": 6,
			"demo4": 3,
		},
	},
	{
		name: "权重4:3:2:1，down调权重30的节点",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "40",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo2": {
				label: map[string]string{
					"weight": "30",
				},
				ip:   "127.0.0.1",
				port: 8580,
				down: true,
			},
			"demo3": {
				label: map[string]string{
					"weight": "20",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
		},
		count: map[string]int{
			"demo1": 12,
			"demo2": 1,
			"demo3": 5,
			"demo4": 2,
		},
	},
	{
		name: "权重4:3:2:1，down调权重20的节点",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "40",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo2": {
				label: map[string]string{
					"weight": "30",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo3": {
				label: map[string]string{
					"weight": "20",
				},
				ip:   "127.0.0.1",
				port: 8580,
				down: true,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
		},
		count: map[string]int{
			"demo1": 10,
			"demo2": 7,
			"demo3": 1,
			"demo4": 2,
		},
	},
	{
		name: "权重4:3:2:1，down调权重10的节点",
		nodes: map[string]nodeDemo{
			"demo1": {
				label: map[string]string{
					"weight": "40",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo2": {
				label: map[string]string{
					"weight": "30",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo3": {
				label: map[string]string{
					"weight": "20",
				},
				ip:   "127.0.0.1",
				port: 8580,
			},
			"demo4": {
				label: map[string]string{
					"weight": "10",
				},
				ip:   "127.0.0.1",
				port: 8580,
				down: true,
			},
		},
		count: map[string]int{
			"demo1": 9,
			"demo2": 6,
			"demo3": 4,
			"demo4": 1,
		},
	},
}
