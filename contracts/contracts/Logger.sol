
// SPDX-License-Identifier: MIT

pragma solidity ^0.8.10;

contract Logger {
    event OperatorRegistered(address operator, uint256 stakeAmount);
    event ResultSubmitted(address solver, uint256 indexed requestId, bytes32 resultRoot);
    event DisputeRaised(address disputer, uint256 indexed requestId);
    event PenaltyApplied(address solver, uint256 penalties);
    event SuccessfulDispute(address disputer);
    event NewReceipt(address sender, uint256 indexed requestId);
}