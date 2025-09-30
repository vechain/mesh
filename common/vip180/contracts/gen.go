//go:generate sh -c "docker run --rm -v $(pwd):/sources ethereum/solc:0.8.20 --evm-version paris --overwrite --optimize --optimize-runs 200 -o /sources/compiled --abi --bin /sources/VIP180.sol /sources/VIP180Token.sol"

package contracts

// This file is used to generate the compiled contract files using go generate
// Run: go generate ./common/vip180/contracts
// This will create:
// - compiled/VIP180.abi (VIP180 base contract ABI)
// - compiled/VIP180.bin (VIP180 base contract bytecode)
// - compiled/VIP180Token.abi (VIP180Token ABI interface)
// - compiled/VIP180Token.bin (VIP180Token bytecode for deployment)
