// SPDX-License-Identifier: MIT

pragma solidity ^0.8.10;

contract Storage {
     struct RequestReceipt {
        uint256[3][3] matrix1;
        uint256[3][3] matrix2;
        uint256[3][3] matrixMul;
        address solver;
        bytes32 root;
        uint256 timestamp;
    }

    struct Operator {
        uint256 stake;
        uint256 penalties;
        uint256 successfulDisputes;
    }

	uint256 public constant CHALLENGE_PERIOD = 7 days;
    uint256 receiptCounter;

    // in the context of the whole code, operators can be:
    // - proof submitter ( requires prior registration )
    // - dispute raise
    mapping(address => Operator) public operators;
    mapping(uint256 => RequestReceipt) matrix;
}