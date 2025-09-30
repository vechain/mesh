// SPDX-License-Identifier: MIT
pragma solidity 0.8.20;

import "./IVIP180.sol";

/**
 * @title VIP180
 * @dev A VIP180 token for testing purposes
 * @notice This contract implements the VIP180 standard for VeChain
 */
contract VIP180 is IVIP180 {
    string private _name;
    string private _symbol;
    uint8 private _decimals;
    uint256 private _totalSupply;
    address private _bridge;
    
    mapping(address => uint256) private _balanceOf;
    mapping(address => mapping(address => uint256)) private _allowance;
    
    constructor(
        string memory name_,
        string memory symbol_,
        uint8 decimals_,
        address bridge_
    ) {
        _name = name_;
        _symbol = symbol_;
        _decimals = decimals_;
        _bridge = bridge_;
        _totalSupply = 1000000 * 10**decimals_; // 1M tokens
        _balanceOf[msg.sender] = _totalSupply;
        emit Transfer(address(0), msg.sender, _totalSupply);
    }
    
    // Implement VIP180 interface functions
    function name() public view override returns (string memory) {
        return _name;
    }
    
    function symbol() public view override returns (string memory) {
        return _symbol;
    }
    
    function decimals() public view override returns (uint8) {
        return _decimals;
    }
    
    function totalSupply() public view override returns (uint256) {
        return _totalSupply;
    }
    
    function bridge() public view override returns (address) {
        return _bridge;
    }
    
    function balanceOf(address _owner) public view override returns (uint256) {
        return _balanceOf[_owner];
    }
    
    function allowance(address _owner, address _spender) public view override returns (uint256) {
        return _allowance[_owner][_spender];
    }
    
    function transfer(address _to, uint256 _value) public override returns (bool) {
        require(_balanceOf[msg.sender] >= _value, "Insufficient balance");
        _balanceOf[msg.sender] -= _value;
        _balanceOf[_to] += _value;
        emit Transfer(msg.sender, _to, _value);
        return true;
    }
    
    function approve(address _spender, uint256 _value) public override returns (bool) {
        _allowance[msg.sender][_spender] = _value;
        emit Approval(msg.sender, _spender, _value);
        return true;
    }
    
    function transferFrom(address _from, address _to, uint256 _value) public override returns (bool) {
        require(_balanceOf[_from] >= _value, "Insufficient balance");
        require(_allowance[_from][msg.sender] >= _value, "Insufficient allowance");
        
        _balanceOf[_from] -= _value;
        _balanceOf[_to] += _value;
        _allowance[_from][msg.sender] -= _value;
        
        emit Transfer(_from, _to, _value);
        return true;
    }
}
