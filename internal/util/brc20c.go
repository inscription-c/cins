package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/inscription-c/cins/constants"
	"regexp"
	"strconv"
)

func init() {
	RegisterProtocol(&CBRC20{})
}

// tickNameRegexp is a regular expression that matches valid tick names.
var tickNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// CBRC20 is a struct that represents a BRC20C protocol.
// It contains the protocol name, operation, tick, max, limit, decimals, tkId, amount, and to.
type CBRC20 struct {
	DefaultProtocol

	Protocol  string `json:"p"`
	Operation string `json:"op"`
	Tick      string `json:"tick"`

	Max      string `json:"max,omitempty"` // deploy
	Limit    string `json:"lim,omitempty"`
	Decimals string `json:"dec,omitempty"`

	//TkId   string `json:"tkid,omitempty"` // mint
	//Amount string `json:"amt,omitempty"`
	//To     string `json:"to,omitempty"`
}

// Name is a method of the BRC20C struct.
// It returns the name of the BRC20C protocol.
func (b *CBRC20) Name() string {
	return constants.ProtocolCBRC20
}

// Check is a method of the BRC20C struct.
// It checks the BRC20C protocol.
// It unmarshals the data into a new BRC20C struct and checks the protocol name, tick name, and operation.
// If the operation is "deploy", it checks the max, limit, and decimals.
// If the operation is "mint", it checks the tkId.
// If the operation is not supported, it returns an error.
func (b *CBRC20) Check() error {
	p := &CBRC20{}
	if err := json.Unmarshal(b.Data(), p); err != nil {
		return err
	}
	if p.Protocol != constants.ProtocolCBRC20 {
		return errors.New("protocol not match")
	}
	if !tickNameRegexp.MatchString(p.Tick) {
		return errors.New("tick name invalid")
	}

	switch p.Operation {
	case constants.OperationDeploy:
		var err error
		var tokenMax uint64
		if p.Max != "" {
			tokenMax, err = strconv.ParseUint(p.Max, 10, 64)
			if err != nil {
				return err
			}
		}

		if p.Limit != "" {
			var limit uint64
			limit, err = strconv.ParseUint(p.Limit, 10, 64)
			if err != nil {
				return err
			}
			if p.Max != "" && limit > tokenMax {
				return errors.New("limit must be less than or equal max")
			}
		}

		if p.Decimals == "" {
			p.Decimals = constants.DecimalsDefault
		} else {
			if _, err := strconv.ParseUint(p.Decimals, 10, 64); err != nil {
				return err
			}
		}
	//case constants.OperationMint:
	//	p.Max = ""
	//	p.Limit = ""
	//	p.Decimals = ""
	//	if tables.StringToInscriptionId(p.TkId) == nil {
	//		return errors.New("tkid invalid")
	//	}
	default:
		return fmt.Errorf("op `%s` not support", p.Operation)
	}
	body, _ := json.Marshal(p)
	p.Reset(body)
	*b = *p
	return nil
}

// Clone returns a new DefaultProtocol.
func (b *CBRC20) Clone() Protocol {
	return &CBRC20{}
}
