package data_mask_strategy

import (
	"fmt"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
	"github.com/eolinker/apinto/strategy"
	"github.com/eolinker/eosc"
)

func checkConfig(conf *Config) error {
	_, err := strategy.ParseFilter(conf.Filters)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	if len(conf.DataMask.Rules) < 1 {
		return fmt.Errorf("at least one rule is required")
	}
	for _, rule := range conf.DataMask.Rules {
		if rule.Match == nil {
			return fmt.Errorf("match is required")
		}
		if rule.Mask == nil {
			return fmt.Errorf("mask is required")
		}
		if rule.Mask.Begin < 0 {
			return fmt.Errorf("begin must be greater than or equal to 0")
		}
		if rule.Mask.Length < -1 {
			return fmt.Errorf("length must be greater than or equal to -1")
		}
		if rule.Match.Type == mask.MatchInner && (rule.Match.Value == mask.MatchInnerValueDate || rule.Match.Value == mask.MatchInnerValueAmount) && rule.Mask.Type == mask.MaskShuffling {
			return fmt.Errorf("date and amount cannot be shuffled")
		}
		if rule.Mask.Type == mask.MaskReplacement && rule.Mask.Replace == nil {
			return fmt.Errorf("replace is required")
		}
	}

	return nil
}

func Check(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {

	return checkConfig(cfg)
}

func Create(id, name string, v *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	if err := Check(v, workers); err != nil {
		return nil, err
	}

	lg := &executor{
		id:   id,
		name: name,
	}

	err := lg.reset(v, workers)
	if err != nil {
		return nil, err
	}

	controller.Store(id)
	return lg, nil
}
