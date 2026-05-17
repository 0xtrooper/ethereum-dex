// Copyright © 2026 Kedar Iyer
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

pragma solidity ^0.8.1;

import "./libraries/SafeERC20.sol";

using SafeERC20 for IERC20;

contract Bank {
	address immutable owner;

	constructor(address _owner) {
		require (_owner != address(0), "must set owner");
		owner = _owner;
	}

	function withdrawTo(address user, address token, uint amount) public {
		require(msg.sender == owner, "only owner can withdraw funds");

		if (token == address(0)) {
			(bool ok, ) = payable(user).call{value: amount}("");
			require(ok); // Must check return value otherwise contract will lose all ETH
		} else {
			IERC20(token).safeTransfer(user, amount);
		}
	}
}

contract OrderBook {
	enum Side {
		BUY,
		SELL
	}
	struct Order {
		address user;
		uint baseQuantity;
		uint quoteQuantity;
		address baseToken;
		address quoteToken;
		Side side;
	}
	mapping(bytes32 => address) public banks;
	mapping(uint => Order) public orders;
	uint public orderCounter = 0;

	event OrderPlaced(uint orderId, address indexed user, address indexed baseToken, address indexed quoteToken, Side side, uint baseQuantity, uint quoteQuantity);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

	function createMarket(address baseToken, address quoteToken) public {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		require(banks[bankHash] == address(0), "market has already been created");

		address bankAddress = address(new Bank(address(this)));
		banks[bankHash] = bankAddress;
	}

	function placeOrder(address baseToken, address quoteToken, Side side, uint baseQuantity, uint quoteQuantity) public payable {
		address bankAddress = banks[bankhash(baseToken, quoteToken)];
		require(bankAddress != address(0), "createMarket before placing an order on it");
		require(baseQuantity > 0 && quoteQuantity > 0, "zero quantity orders not permitted");

		// forward all incoming ETH to the market's bank
		if (msg.value > 0) {
			payable(bankAddress).transfer(msg.value);
		}

		if (side == Side.SELL) {
			if (msg.value > 0) {
				require(baseToken == address(0), "base token should be 0x0 when selling ETH");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			} else {
				IERC20 baseTokenIERC20 = IERC20(baseToken);
				uint beforeBalance = baseTokenIERC20.balanceOf(bankAddress);
				baseTokenIERC20.safeTransferFrom(msg.sender, bankAddress, baseQuantity);
				uint afterBalance = baseTokenIERC20.balanceOf(bankAddress);
				uint transferredBaseQuantity = afterBalance - beforeBalance;
				if (transferredBaseQuantity != baseQuantity) {
					baseQuantity = transferredBaseQuantity;
				}
			}
		} else if (side == Side.BUY) {
			if (msg.value > 0) {
				require(quoteToken == address(0), "quote token should be 0x0 when buying with ETH");
				require(quoteQuantity == msg.value, "mismatch between provided quoteQuantity and amount of ETH sent");
			} else {
				uint beforeBalance = IERC20(quoteToken).balanceOf(bankAddress);
				IERC20(quoteToken).safeTransferFrom(msg.sender, bankAddress, quoteQuantity);
				uint afterBalance = IERC20(quoteToken).balanceOf(bankAddress);
				uint transferredQuoteQuantity = afterBalance - beforeBalance;
				if (transferredQuoteQuantity != quoteQuantity) {
					quoteQuantity = transferredQuoteQuantity;
				}
			}
		}
		uint orderId = ++orderCounter;
		orders[orderId] = Order(msg.sender, baseQuantity, quoteQuantity, baseToken, quoteToken, side);
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, side, baseQuantity, quoteQuantity);
	}

	function cancelOrder(uint orderId) public {
		Order memory order = orders[orderId];
		require(msg.sender == order.user && msg.sender != address(0), "users can only cancel their own order / order may not exist");
		delete orders[orderId];
		address bankAddress = banks[bankhash(order.baseToken, order.quoteToken)];

		// There is a re-entrancy risk here. The order has been deleted ahead of time to prevent hacks
		if (order.side == Side.SELL) {
			Bank(bankAddress).withdrawTo(order.user, order.baseToken, order.baseQuantity);
		} else if (order.side == Side.BUY) {
			Bank(bankAddress).withdrawTo(order.user, order.quoteToken, order.quoteQuantity);
		}
		emit OrderCanceled(orderId);
	}

	function fillOrder(uint orderId, uint baseQuantity) public payable {
		Order memory order = orders[orderId];

		address bankAddress = banks[bankhash(order.baseToken, order.quoteToken)];

		uint quoteQuantity = (baseQuantity * order.quoteQuantity) / order.baseQuantity;
		if (msg.value > 0) {
			if (order.side == Side.SELL) {
				require(order.quoteToken == address(0), "quote token should be 0x0");
				require(quoteQuantity == msg.value, "mismatch between quoteQuantity and amount of ETH sent");
			} else if (order.side == Side.BUY) {
				require(order.baseToken == address(0), "base token should be 0x0");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}

			// if checks pass, forward all incoming ETH to the market's bank
			(bool ok, ) = payable(bankAddress).call{value: msg.value}("");
			require(ok); // Must check return value otherwise contract will lose all ETH
		}

		require(baseQuantity > 0, "zero quantity fills not permitted");
		require(baseQuantity <= order.baseQuantity, "trying to fill more than order size");
		require(quoteQuantity > 0, "calculated quote quantity is zero");
		orders[orderId].baseQuantity -= baseQuantity;
		orders[orderId].quoteQuantity -= quoteQuantity;
		if (orders[orderId].baseQuantity == 0) {
			delete orders[orderId];
		}
		if (order.side == Side.SELL) {
			if (order.quoteToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), quoteQuantity); // send ETH
			} else {
				IERC20(order.quoteToken).safeTransferFrom(msg.sender, order.user, quoteQuantity);
			}
			Bank(bankAddress).withdrawTo(msg.sender, order.baseToken, baseQuantity);
		} else if (order.side == Side.BUY) {
			if (order.baseToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), baseQuantity); // send ETH
			} else {
				IERC20(order.baseToken).safeTransferFrom(msg.sender, order.user, baseQuantity);
			}
			Bank(bankAddress).withdrawTo(msg.sender, order.quoteToken, quoteQuantity);
		}
		emit OrderFill(orderId, baseQuantity);
	}

	function bankhash(address baseToken, address quoteToken) public pure returns (bytes32) {
		return keccak256(abi.encodePacked(baseToken, quoteToken));
	}
}
