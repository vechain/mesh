package utils

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
)

// The rationale for this is to support both api.JSONClause and api.Clause
// in the future, after amending the thorclient, we should not need this adapter

// ClauseValue interface for handling both JSONClause.Value and Clause.Value
type ClauseValue interface {
	MarshalText() ([]byte, error)
}

// ClauseData interface for handling both JSONClause and Clause
type ClauseData interface {
	GetValue() ClauseValue
	GetTo() *thor.Address
	GetData() string
}

// JSONClauseAdapter adapts api.JSONClause to ClauseData interface
type JSONClauseAdapter struct {
	clause *api.JSONClause
}

func (j JSONClauseAdapter) GetValue() ClauseValue { return &j.clause.Value }
func (j JSONClauseAdapter) GetTo() *thor.Address  { return j.clause.To }
func (j JSONClauseAdapter) GetData() string       { return j.clause.Data }

// ClauseAdapter adapts api.Clause to ClauseData interface
type ClauseAdapter struct {
	clause *api.Clause
}

func (c ClauseAdapter) GetValue() ClauseValue { return c.clause.Value }
func (c ClauseAdapter) GetTo() *thor.Address  { return c.clause.To }
func (c ClauseAdapter) GetData() string       { return c.clause.Data }

// parseTransactionOperationsFromClauseData is a generic helper function that parses operations from clause data
func parseTransactionOperationsFromClauseData(clauseData []ClauseData, originAddr string, gas uint64, status *string) []*types.Operation {
	var operations []*types.Operation
	hasValueTransfer, hasContractInteraction, hasEnergyTransfer := false, false, gas > 0

	// Analyze clauses
	for _, clause := range clauseData {
		valueBytes, _ := clause.GetValue().MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)
		if value.Cmp(big.NewInt(0)) > 0 {
			hasValueTransfer = true
		}
		if len(clause.GetData()) > 0 || (clause.GetTo() != nil && !clause.GetTo().IsZero()) {
			hasContractInteraction = true
		}
	}

	if !hasValueTransfer && !hasContractInteraction && !hasEnergyTransfer {
		return operations
	}

	operationIndex := 0
	for clauseIndex, clause := range clauseData {
		valueBytes, _ := clause.GetValue().MarshalText()
		value := new(big.Int)
		value.SetString(string(valueBytes), 10)

		if value.Cmp(big.NewInt(0)) > 0 {
			valueStr := value.String()
			// Sender operation
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                OperationTypeTransfer,
				Status:              status,
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "-" + valueStr, Currency: VETCurrency},
				Metadata:            map[string]any{"clauseIndex": clauseIndex},
			})
			operationIndex++

			// Receiver operation
			if clause.GetTo() != nil && !clause.GetTo().IsZero() {
				operations = append(operations, &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
					Type:                OperationTypeTransfer,
					Status:              status,
					Account:             &types.AccountIdentifier{Address: clause.GetTo().String()},
					Amount:              &types.Amount{Value: valueStr, Currency: VETCurrency},
					Metadata:            map[string]any{"clauseIndex": clauseIndex},
				})
				operationIndex++
			}
		}

		// Contract interaction operation
		if len(clause.GetData()) > 0 || (clause.GetTo() != nil && !clause.GetTo().IsZero()) {
			toAddr := ""
			if clause.GetTo() != nil {
				toAddr = clause.GetTo().String()
			}
			operations = append(operations, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
				Type:                OperationTypeContractCall,
				Status:              status,
				Account:             &types.AccountIdentifier{Address: originAddr},
				Amount:              &types.Amount{Value: "0", Currency: VETCurrency},
				Metadata: map[string]any{
					"clauseIndex": clauseIndex,
					"to":          toAddr,
					"data":        "0x" + fmt.Sprintf("%x", clause.GetData()),
				},
			})
			operationIndex++
		}
	}

	// Energy transfer operation
	if hasEnergyTransfer {
		operations = append(operations, &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
			Type:                OperationTypeFee,
			Status:              status,
			Account:             &types.AccountIdentifier{Address: originAddr},
			Amount:              &types.Amount{Value: "-" + strconv.FormatUint(gas, 10), Currency: VTHOCurrency},
			Metadata:            map[string]any{"gas": strconv.FormatUint(gas, 10)},
		})
	}

	return operations
}

// parseTransactionOperationsFromClauses is a helper function that parses operations from clauses
func parseTransactionOperationsFromClauses(clauses []*api.JSONClause, originAddr string, gas uint64, status *string) []*types.Operation {
	clauseData := make([]ClauseData, len(clauses))
	for i, clause := range clauses {
		clauseData[i] = JSONClauseAdapter{clause: clause}
	}
	return parseTransactionOperationsFromClauseData(clauseData, originAddr, gas, status)
}

// ParseTransactionOperationsFromTransactionClauses is a helper function that parses operations from transactions.Clauses
func ParseTransactionOperationsFromTransactionClauses(clauses api.Clauses, originAddr string, gas uint64, status *string) []*types.Operation {
	clauseData := make([]ClauseData, len(clauses))
	for i, clause := range clauses {
		clauseData[i] = ClauseAdapter{clause: clause}
	}
	return parseTransactionOperationsFromClauseData(clauseData, originAddr, gas, status)
}
