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
    }

   function submitResult(bytes32 root, uint256[3][3] memory results, uint256 requestId) onlyOperator external {
        // TODO: input validations
        RequestReciept storage reciept = matrix[requestId];
        reciept.matrixMul = results;
        reciept.timestamp = block.timestamp;
        reciept.root = root;
        reciept.solver = msg.sender;
   }
}
