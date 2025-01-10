package parser_test

import (
	"testing"

	balanceErrors "github.com/cosmostation/cvms/internal/packages/utility/balance/errors"
	parser "github.com/cosmostation/cvms/internal/packages/utility/balance/parser"
	"github.com/stretchr/testify/assert"
)

func TestCosmosBalanceParsing(t *testing.T) {

	resp := []byte(`{
  "balances": [
    {
      "denom": "factory/neutronxyzasdf/Governance",
      "amount": "123"
    },
    {
      "denom": "ibc/bla",
      "amount": "120"
    },
    {
      "denom": "untrn",
      "amount": "10000"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "3"
  }
}`)

	rep_no_balance := []byte(`{
  "balances": [],
  "pagination": {
    "next_key": null,
    "total": "0"
  }
}`)

	balance_untrn, err := parser.CosmosBalanceParser(resp, "untrn")
	assert.Nil(t, err)
	balance_unknown, err := parser.CosmosBalanceParser(resp, "asdf")
	assert.ErrorIs(t, balanceErrors.ErrBalanceNotFound, err)
	balance_unknown_no_balance, err := parser.CosmosBalanceParser(rep_no_balance, "asdf")
	assert.ErrorIs(t, balanceErrors.ErrBalanceNotFound, err)

	assert.Equal(t, float64(10000), balance_untrn)
	assert.Equal(t, float64(0), balance_unknown)
	assert.Equal(t, float64(0), balance_unknown_no_balance)

}
