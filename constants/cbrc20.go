package constants

import "regexp"

const (
	ProtocolCBRC20  = "c-brc-20"
	OperationDeploy = "deploy"
	OperationMint   = "mint"
	DecimalsDefault = "18"

	CInsDescriptionTypeBlockchain = "blockchain"
	CInsDescriptionTypeOrdinals   = "ordinals"
)

// TickNameRegexp is a regular expression that matches valid tick names.
var TickNameRegexp = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
