package grey_strategy

type Weighted interface {
	GetId() uint32
	GetWeight() int
}

// NewRobin 初始化一个池子
func NewRobin(servers ...Weighted) *Robin {
	newRobin := &Robin{}
	newRobin.updateServers(servers)
	return newRobin
}

type Training struct {
	Server        Weighted
	Weight        int //初始化设置权重值
	CurrentWeight int //目前的权重值
}

type Robin struct {
	Weighted []Weighted
	Training []*Training
}

func (l *Robin) updateServers(servers []Weighted) {
	weighted := make([]*Training, 0)
	for _, v := range servers {
		w := &Training{
			Server:        v,
			Weight:        v.GetWeight(),
			CurrentWeight: 0,
		}
		weighted = append(weighted, w)
	}
	l.Training = weighted
	l.Weighted = servers
}

// Select remove为需要屏蔽的ID，
func (l *Robin) Select(remove ...uint) Weighted {
	if len(l.Training) == 0 {
		return nil
	}
	w := l.nextWeighted(remove)
	if w == nil {
		return nil
	}
	return w.Server
}
func (l *Robin) nextWeighted(remove []uint) (best *Training) {
	total := 0
	for i := 0; i < len(l.Training); i++ {
		w := l.Training[i]
		if w == nil {
			continue
		}
		isFind := false
		for _, v := range remove {
			if v == uint(w.Server.GetId()) {
				isFind = true
			}
		}
		if isFind {
			continue
		}
		//每次都加原始的权重值
		w.CurrentWeight += w.Weight
		//所有权重之和
		total += w.Weight
		//判断当前最大的权重。不管有没有最大  先取第一个、然后依次对比、取出最大
		if best == nil || w.CurrentWeight > best.CurrentWeight {
			best = w
		}
	}
	if best == nil {
		return best
	}
	//抽出后-最大权重值
	best.CurrentWeight -= total
	return best
}
