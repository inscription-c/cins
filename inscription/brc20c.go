package inscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dotbitHQ/insc/constants"
	"github.com/dotbitHQ/insc/model"
	"github.com/shopspring/decimal"
	"regexp"
)

var tickNameRegexp = regexp.MustCompile(`^[a-z0-9-]+$`)

type BRC20C struct {
	DefaultProtocol

	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Tick      string `json:"tick"`

	Max      string `json:"max,omitempty"` // deploy
	Limit    string `json:"lim,omitempty"`
	Decimals string `json:"dec,omitempty"`

	TkId   string `json:"tkid,omitempty"` // mint
	Amount string `json:"amt,omitempty"`
	To     string `json:"to,omitempty"`
}

func (b *BRC20C) Name() string {
	return constants.ProtocolBRC20C
}

func (b *BRC20C) Check() error {
	p := &BRC20C{}
	if err := json.Unmarshal(b.Data(), p); err != nil {
		return err
	}
	if p.Protocol != constants.ProtocolBRC20C {
		return errors.New("protocol not match")
	}
	if !tickNameRegexp.MatchString(p.Tick) {
		return errors.New("tick name invalid")
	}

	switch p.Operation {
	case constants.OperationDeploy:
		p.TkId = ""
		p.Amount = ""
		p.To = ""

		var err error
		var m decimal.Decimal
		if p.Max == "" {
			return errors.New("max can not empty")
		}

		m, err = decimal.NewFromString(p.Max)
		if err != nil {
			return err
		}
		if m.LessThanOrEqual(decimal.Zero) {
			return errors.New("max must greater than 0")
		}

		if p.Limit != "" {
			lim, err := decimal.NewFromString(p.Limit)
			if err != nil {
				return err
			}
			if lim.LessThanOrEqual(decimal.Zero) {
				return errors.New("limit must greater than 0")
			}
			if p.Max != "" {
				if lim.GreaterThan(m) {
					return errors.New("lim must less than or equal max")
				}
			}
		}

		if p.Decimals != "" {
			p.Decimals = constants.DecimalsDefault
		} else {
			dec, err := decimal.NewFromString(p.Decimals)
			if err != nil {
				return err
			}
			if dec.LessThan(decimal.Zero) {
				return errors.New("dec must greater than or equal 0, default is 18")
			}
		}
	case constants.OperationMint:
		p.Max = ""
		p.Limit = ""
		p.Decimals = ""
		if model.InscriptionIdToOutpoint(p.TkId) == nil {
			return errors.New("tkid invalid")
		}
	default:
		return fmt.Errorf("op `%s` not support", p.Operation)
	}
	body, _ := json.Marshal(p)
	p.Reset(body)
	*b = *p
	return nil
}
