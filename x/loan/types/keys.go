package types

import (
	sdkmath "cosmossdk.io/math"
)

const (
	// ModuleName defines the module name
	ModuleName = "loan"

	// secondary module key for splitting collateral not_bonded_tokens_pool
	Nbtp = "not_bonded_tokens_pool"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_loan"
)

// cwei is the smallest unit of collateral
var Cwei = sdkmath.NewInt(1000000000)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	LoanKey      = "Loan/value/"
	LoanCountKey = "Loan/count/"
)
