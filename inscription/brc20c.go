package inscription

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/shopspring/decimal"
	"regexp"
)

// tickNameRegexp is a regular expression that matches valid tick names.
var tickNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// BRC20C is a struct that represents a BRC20C protocol.
// It contains the protocol name, operation, tick, max, limit, decimals, tkId, amount, and to.
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

// Name is a method of the BRC20C struct.
// It returns the name of the BRC20C protocol.
func (b *BRC20C) Name() string {
	return constants.ProtocolBRC20C
}

// Check is a method of the BRC20C struct.
// It checks the BRC20C protocol.
// It unmarshals the data into a new BRC20C struct and checks the protocol name, tick name, and operation.
// If the operation is "deploy", it checks the max, limit, and decimals.
// If the operation is "mint", it checks the tkId.
// If the operation is not supported, it returns an error.
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
		// Check the deployment operation
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
		// Check the mint operation
		p.Max = ""
		p.Limit = ""
		p.Decimals = ""
		if index.InscriptionIdToOutpoint(p.TkId) == nil {
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
