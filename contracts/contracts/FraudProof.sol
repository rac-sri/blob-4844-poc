// SPDX-License-Identifier: MIT

pragma solidity ^0.8.10;

import "./Storage.sol";

contract FraudProof is Storage {
    function hashUint256(uint256 data) private pure returns (bytes32) {
        return keccak256(abi.encodePacked(data));
    }

    function merkleTreeRoot(uint256[9] memory values) public pure returns (bytes32) {
        bytes32[9] memory nodes;
        for (uint i = 0; i < 9; i++) {
            nodes[i] = hashUint256(values[i]);
        }

        bytes32[5] memory level2;
        for (uint i = 0; i < 8; i += 2) {
            level2[i / 2] = keccak256(abi.encodePacked(nodes[i], nodes[i + 1]));
        }
        level2[4] = nodes[8];
        bytes32[3] memory level3;
        level3[0] = keccak256(abi.encodePacked(level2[0], level2[1]));
        level3[1] = keccak256(abi.encodePacked(level2[2], level2[3]));
      
        level3[2] = level2[4];
  
        bytes32 merkleRoot = keccak256(abi.encodePacked(level3[0], level3[1], level3[2]));

        return merkleRoot;
    }

    function multiplyMatrices(uint256[3][3] memory a, uint256[3][3] memory b) private pure returns (uint256[9] memory, uint256[3][3] memory) {
        uint256[3][3] memory c;
        uint256[9] memory singleArrayResult;
        uint256 position = 0;
       
        for (uint i = 0; i < 3; i++) {
            for (uint j = 0; j < 3; j++) {
                uint256 sum = 0;
                for (uint k = 0; k < 3; k++) {
                    sum += a[i][k] * b[k][j];
                }
                c[i][j] = sum;
                singleArrayResult[position++] = c[i][j];
            }
        }

        return (singleArrayResult, c);
    }

    function raiseDispute(uint256 requestId) external {
        RequestReceipt storage receipt = matrix[requestId];

        require(block.timestamp <= receipt.timestamp + CHALLENGE_PERIOD, "Challenge period has expired");

        (uint256[9] memory result,uint256[3][3] memory correctMatrix ) = multiplyMatrices(receipt.matrix1, receipt.matrix2);

        bytes32 generatedRoot = merkleTreeRoot(result);

        require(generatedRoot != receipt.root, "No discrepancy found; computation correct");

        operators[receipt.solver].penalties += 1;
        operators[msg.sender].successfulDisputes += 1;

        receipt.matrixMul = correctMatrix;
        receipt.root = generatedRoot; 
    }
}
