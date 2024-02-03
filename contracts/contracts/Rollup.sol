// SPDX-License-Identifier: MIT

pragma solidity ^0.8.10;

import "./FraudProof.sol";

contract Rollup is FraudProof {
    modifier onlyOperator()  {
        require(operators[msg.sender].stake > 0,"invalid operator");
        _;
        // Gas optimisation: change modifier to a function
    }

    function registerAsAnOperator() payable external {
        require(msg.value > 1 ether / 100, "Invalid minimum staking amount");    // min staking amount = 0.01 eth
        operators[msg.sender].stake += msg.value; 
        emit OperatorRegistered(msg.sender, msg.value);
    }

    function addNewReceipt(uint256[3][3] memory matrix1, uint256[3][3] memory matrix2) external {
        RequestReceipt memory receipt = matrix[++receiptCounter];
        receipt.matrix1 = matrix1;
        receipt.matrix2 = matrix2;
        matrix[receiptCounter] = receipt;
        emit NewReceipt(msg.sender, receiptCounter);
    }

   function submitResult(bytes32 root, uint256[3][3] memory results, uint256 requestId) onlyOperator external {
        // TODO: input validations
        RequestReceipt memory receipt = matrix[requestId];
        receipt.matrixMul = results;
        receipt.timestamp = block.timestamp;
        receipt.root = root;
        receipt.solver = msg.sender;
        matrix[requestId] = receipt;
        emit ResultSubmitted(msg.sender, requestId, root);
   }
}
