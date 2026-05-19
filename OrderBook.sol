// Copyright © 2026 Kedar Iyer
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

pragma solidity ^0.8.20;

import "./libraries/SafeERC20.sol";

using SafeERC20 for IERC20;

contract Bank {
	address immutable owner;

	constructor(address _owner) {
		require (_owner != address(0), "must set owner");
		owner = _owner;
	}

	receive() external payable {}

	// This function limits the gas forwarded on ETH transfers to prevent re-entrancy
	function withdrawTo(address user, address token, uint amount) external {
		require(msg.sender == owner, "only owner can withdraw funds");
		require(amount > 0, "amount is zero");

		if (token == address(0)) {
			(bool ok,) = payable(user).call{value: amount, gas: 2300}("");
			require(ok, "eth transfer failed");
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
		Side side;
	}
	mapping(bytes32 => address payable) public banks;
	mapping(bytes32 => mapping(uint => Order)) orders;
	uint public orderCounter = 0;

	event OrderPlaced(uint indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, Side side, uint baseQuantity, uint quoteQuantity);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

	function createMarket(address baseToken, address quoteToken) external {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		require(banks[bankHash] == address(0), "market has already been created");

		address payable bankAddress = payable(address(new Bank(address(this))));
		banks[bankHash] = bankAddress;
	}

	function placeOrder(address baseToken, address quoteToken, Side side, uint baseQuantity, uint quoteQuantity) external payable returns (uint orderId) {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		address bankAddress = banks[bankHash];
		require(bankAddress != address(0), "createMarket before placing an order on it");
		require(baseQuantity > 0 && quoteQuantity > 0, "zero quantity orders not permitted");

		if (side == Side.SELL) {
			if (msg.value > 0) {
				require(baseToken == address(0), "base token should be 0x0 when selling ETH");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
				(bool ok,) = payable(bankAddress).call{value: msg.value, gas: 2300}("");
				require(ok, "eth transfer to bank failed");

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
				(bool ok,) = payable(bankAddress).call{value: msg.value, gas: 2300}("");
				require(ok, "eth transfer to bank failed");
			} else {
				IERC20 quoteTokenIERC20 = IERC20(quoteToken);
				uint beforeBalance = quoteTokenIERC20.balanceOf(bankAddress);
				quoteTokenIERC20.safeTransferFrom(msg.sender, bankAddress, quoteQuantity);
				uint afterBalance = quoteTokenIERC20.balanceOf(bankAddress);
				uint transferredQuoteQuantity = afterBalance - beforeBalance;
				if (transferredQuoteQuantity != quoteQuantity) {
					quoteQuantity = transferredQuoteQuantity;
				}
			}
		}
		unchecked {
			orderId = ++orderCounter;
		}
		orders[bankHash][orderId] = Order(msg.sender, baseQuantity, quoteQuantity, side);
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, bankHash, side, baseQuantity, quoteQuantity);
	}

	function cancelOrder(address baseToken, address quoteToken, uint orderId) external {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		Order memory order = orders[bankHash][orderId];
		require(msg.sender == order.user, "users can only cancel their own order / order may not exist");
		delete orders[bankHash][orderId];
		address payable bankAddress = banks[bankHash];

		// Bank.withdrawTo is gas limited so there is minimal re-entrancy risk, but orders are being deleted
		// before withdraws anyway to prevent hijinks within the 2300 forwarded gas 
		if (order.side == Side.SELL) {
			Bank(bankAddress).withdrawTo(order.user, baseToken, order.baseQuantity);
		} else if (order.side == Side.BUY) {
			Bank(bankAddress).withdrawTo(order.user, quoteToken, order.quoteQuantity);
		}
		emit OrderCanceled(orderId);
	}

	function fillOrder(uint orderId, address baseToken, address quoteToken, uint baseQuantity) external payable {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		Order memory order = orders[bankHash][orderId];

		address payable bankAddress = banks[bankHash];

		require(baseQuantity > 0, "zero quantity fills not permitted");
		require(baseQuantity <= order.baseQuantity, "trying to fill more than order size");

		uint quoteQuantity = (baseQuantity * order.quoteQuantity) / order.baseQuantity;
		require(quoteQuantity > 0, "calculated quote quantity is zero");

		if (msg.value > 0) {
			if (order.side == Side.SELL) {
				require(quoteToken == address(0), "quote token should be 0x0");
				require(quoteQuantity == msg.value, "mismatch between quoteQuantity and amount of ETH sent");
			} else if (order.side == Side.BUY) {
				require(baseToken == address(0), "base token should be 0x0");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}

			// if checks pass, forward all incoming ETH to the market's bank
			(bool ok,) = payable(bankAddress).call{value: msg.value, gas: 2300}("");
			require(ok, "eth transfer to bank failed");
		}

		// Bank.withdrawTo is gas limited so there is no real re-entrancy risk here, but the
		// base and quote quantities are being updated before any withdraws to remove risk anyway
		// Additionally orders are being deleted when possible to 0 all storage slots for that order
		orders[bankHash][orderId].baseQuantity -= baseQuantity;
		orders[bankHash][orderId].quoteQuantity -= quoteQuantity;
		if (orders[bankHash][orderId].baseQuantity == 0) {
			delete orders[bankHash][orderId];
		}

		if (order.side == Side.SELL) {
			if (quoteToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), quoteQuantity); // send ETH
			} else {
				IERC20(quoteToken).safeTransferFrom(msg.sender, order.user, quoteQuantity);
			}
			Bank(bankAddress).withdrawTo(msg.sender, baseToken, baseQuantity);
		} else if (order.side == Side.BUY) {
			if (baseToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), baseQuantity); // send ETH
			} else {
				IERC20(baseToken).safeTransferFrom(msg.sender, order.user, baseQuantity);
			}
			Bank(bankAddress).withdrawTo(msg.sender, quoteToken, quoteQuantity);
		}
		emit OrderFill(orderId, baseQuantity);
	}

	function bankhash(address baseToken, address quoteToken) public pure returns (bytes32) {
		return keccak256(abi.encodePacked(baseToken, quoteToken));
	}

	function getorder(address baseToken, address quoteToken, uint orderId) external view returns (Order memory) {
		bytes32 bankHash = bankhash(baseToken, quoteToken);
		return orders[bankHash][orderId];
	}
}
