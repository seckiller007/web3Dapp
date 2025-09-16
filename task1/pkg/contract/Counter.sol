// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;
//
contract Counter {
    uint256 private _count;

    // 初始化计数器为0
    constructor() {
        _count = 0;
    }

    // 获取当前计数
    function getCount() public view returns (uint256) {
        return _count;
    }

    // 增加计数
    function increment() public {
        _count++;
    }

    // 减少计数
    function decrement() public {
        require(_count > 0, "Count cannot be negative");
        _count--;
    }

    // 重置计数为0
    function reset() public {
        _count = 0;
    }
}
