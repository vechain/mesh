//go:generate sh -c "docker run --rm -v $(pwd):/sources ghcr.io/argotorg/solc:0.8.20 --evm-version paris --overwrite --optimize --optimize-runs 200 -o /sources/compiled --abi --bin /sources/IVIP180.sol /sources/VIP180.sol"

package contracts

// This file is used to generate the compiled contract files using go generate
// Run: go generate ./common/vip180/contracts
// This will create:
// - compiled/IVIP180.abi (IVIP180 interface ABI)
// - compiled/IVIP180.bin (IVIP180 interface bytecode)
// - compiled/VIP180.abi (VIP180 contract ABI)
// - compiled/VIP180.bin (VIP180 bytecode for deployment)
