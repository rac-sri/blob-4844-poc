// SPDX-License-Identifier: MIT

pragma solidity ^0.8.10;

contract Storage {
     struct RequestReciept {
        uint256[3][3] matrix1;
        uint256[3][3] matrix2;
        uint256[3][3] matrixMul;
        address solver;
        uint256 timestamp;
        bytes32 root;
    }

    struct Operator {
        uint256 stake;
        uint256 penalities;
        uint256 successfulDisputes;
    }

	uint256 public constant CHALLENGE_PERIOD = 7 days;

    // in the context of the whole code, operators can be:
    // - proof submitter ( requires prior registration )
    // - dispute raise
    mapping(address => Operator) public operators;
    mapping(uint256 => RequestReciept) matrix;
}