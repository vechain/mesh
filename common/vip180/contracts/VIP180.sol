// SPDX-License-Identifier: MIT
pragma solidity 0.8.20;

/**
 * @title VIP180
 * @dev Abstract contract for the full VIP180 Token standard
 * @notice This is the VIP180 interface that VIP180Token will implement
 */
abstract contract VIP180 {
    /// @notice Returns the name of the token
    function name() public view virtual returns (string memory);

    /// @notice Returns the symbol of the token
    function symbol() public view virtual returns (string memory);

    /// @notice Returns the decimals of the token
    function decimals() public view virtual returns (uint8);

    /// @notice Returns the total supply of the token
    function totalSupply() public view virtual returns (uint256);

    /// @notice Returns the bridge address
    function bridge() public view virtual returns (address);

    /// @param _owner The address from which the balance will be retrieved
    /// @return The balance
    function balanceOf(address _owner) public view virtual returns (uint256);

    /// @notice send `_value` token to `_to` from `msg.sender`
    /// @param _to The address of the recipient
    /// @param _value The amount of token to be transferred
    /// @return Whether the transfer was successful or not
    function transfer(address _to, uint256 _value) public virtual returns (bool);

    /// @notice send `_value` token to `_to` from `_from` on the condition it is approved by `_from`
    /// @param _from The address of the sender
    /// @param _to The address of the recipient
    /// @param _value The amount of token to be transferred
    /// @return Whether the transfer was successful or not
    function transferFrom(address _from, address _to, uint256 _value) public virtual returns (bool);

    /// @notice `msg.sender` approves `_addr` to spend `_value` tokens
    /// @param _spender The address of the account able to transfer the tokens
    /// @param _value The amount of wei to be approved for transfer
    /// @return Whether the approval was successful or not
    function approve(address _spender, uint256 _value) public virtual returns (bool);

    /// @param _owner The address of the account owning tokens
    /// @param _spender The address of the account able to transfer the tokens
    /// @return Amount of remaining tokens allowed to spent
    function allowance(address _owner, address _spender) public view virtual returns (uint256);

    event Transfer(address indexed _from, address indexed _to, uint256 _amount);
    event Approval(address indexed _owner, address indexed _spender, uint256 _amount);
}
