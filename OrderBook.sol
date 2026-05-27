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

	// This function introduces re-entrancy risk. Be very careful before using!
	// Checks-effects-interactions pattern should always be used when withdrawing
	// https://docs.soliditylang.org/en/latest/security-considerations.html
	function withdrawTo(address user, address token, uint amount) external {
		require(msg.sender == owner, "only owner can withdraw funds");
		require(amount > 0, "amount is zero");

		if (token == address(0)) {
			(bool ok,) = payable(user).call{value: amount}("");
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
		uint price;
		Side side;
	}
	mapping(address => mapping(address => address payable)) banks; // baseToken -> quoteToken -> bankAddress
	mapping(address => mapping(address => mapping(uint => Order))) orders; // baseToken -> quoteToken -> orderId -> Order
	uint public orderCounter = 0;

	event OrderPlaced(uint indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, Side side, uint baseQuantity, uint price);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

	function createMarket(address baseToken, address quoteToken) external {
		require(banks[baseToken][quoteToken] == address(0), "market has already been created");

		address payable bankAddress = payable(address(new Bank(address(this))));
		banks[baseToken][quoteToken] = bankAddress;
	}

	function placeOrder(address baseToken, address quoteToken, Side side, uint baseQuantity, uint price) external payable returns (uint orderId) {
		address bankAddress = banks[baseToken][quoteToken];
		require(bankAddress != address(0), "createMarket before placing an order on it");
		require(baseQuantity > 0 && price > 0, "zero quantity/price orders not permitted");

		(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(quoteToken).tryGetDecimals();
		require(decimalCallSuccess, "failed to get decimals for token");
		uint quoteQuantity = baseQuantity * price / quoteTokenDecimals;
		require(quoteQuantity > 0, "calculated quote quantity is zero");

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
				require(quoteQuantity == msg.value, "mismatch between calculated quoteQuantity and amount of ETH sent");
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
		orders[baseToken][quoteToken][orderId] = Order(msg.sender, baseQuantity, price, side);
		bytes32 markethash = keccak256(abi.encodePacked(baseToken, quoteToken));
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, markethash, side, baseQuantity, price);
	}

	function cancelOrder(address baseToken, address quoteToken, uint orderId) external {
		Order memory order = orders[baseToken][quoteToken][orderId];
		require(msg.sender == order.user, "users can only cancel their own order / order may not exist");
		delete orders[baseToken][quoteToken][orderId];
		address payable bankAddress = banks[baseToken][quoteToken];

		// WARNING: Re-entrancy risk! 
		// Orders are being deleted before withdraws to mitigate risk
		if (order.side == Side.SELL) {
			Bank(bankAddress).withdrawTo(order.user, baseToken, order.baseQuantity);
		} else if (order.side == Side.BUY) {
			(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(quoteToken).tryGetDecimals();
			require(decimalCallSuccess, "failed to get decimals for token");
			uint quoteQuantity = order.baseQuantity * order.price / quoteTokenDecimals;
			Bank(bankAddress).withdrawTo(order.user, quoteToken, quoteQuantity);
		}
		emit OrderCanceled(orderId);
	}

	function fillOrder(uint orderId, address baseToken, address quoteToken, uint baseQuantityToFill) external payable {
		Order memory order = orders[baseToken][quoteToken][orderId];

		address payable bankAddress = banks[baseToken][quoteToken];

		require(baseQuantityToFill > 0, "zero quantity fills not permitted");
		require(baseQuantityToFill <= order.baseQuantity, "trying to fill more than order size");

		(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(quoteToken).tryGetDecimals();
		require(decimalCallSuccess, "failed to get decimals for token");
		uint quoteQuantityToFill = baseQuantityToFill * order.price / quoteTokenDecimals;
		require(quoteQuantityToFill > 0, "calculated quote quantity is zero");

		if (msg.value > 0) {
			if (order.side == Side.SELL) {
				require(quoteToken == address(0), "quote token should be 0x0");
				require(quoteQuantityToFill == msg.value, "mismatch between quoteQuantityToFill and amount of ETH sent");
			} else if (order.side == Side.BUY) {
				require(baseToken == address(0), "base token should be 0x0");
				require(baseQuantityToFill == msg.value, "mismatch between provided baseQuantityToFill and amount of ETH sent");
			}

			// if checks pass, forward all incoming ETH to the market's bank
			(bool ok,) = payable(bankAddress).call{value: msg.value, gas: 2300}("");
			require(ok, "eth transfer to bank failed");
		}

		// WARNING: Re-entrancy risk! 
		// base and quote quantities are being updated before any withdraws to remove risk 
		// Additionally orders are being deleted when possible to 0 all storage slots for that order
		orders[baseToken][quoteToken][orderId].baseQuantity -= baseQuantityToFill;
		if (orders[baseToken][quoteToken][orderId].baseQuantity == 0) {
			delete orders[baseToken][quoteToken][orderId];
		}

		if (order.side == Side.SELL) {
			if (quoteToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), quoteQuantityToFill); // send ETH
			} else {
				IERC20(quoteToken).safeTransferFrom(msg.sender, order.user, quoteQuantityToFill);
			}
			Bank(bankAddress).withdrawTo(msg.sender, baseToken, baseQuantityToFill);
		} else if (order.side == Side.BUY) {
			if (baseToken == address(0)) {
				Bank(bankAddress).withdrawTo(order.user, address(0), baseQuantityToFill); // send ETH
			} else {
				IERC20(baseToken).safeTransferFrom(msg.sender, order.user, baseQuantityToFill);
			}
			Bank(bankAddress).withdrawTo(msg.sender, quoteToken, quoteQuantityToFill);
		}
		emit OrderFill(orderId, baseQuantityToFill);
	}

	function getBankAddress(address baseToken, address quoteToken) public view returns (address payable) {
		return banks[baseToken][quoteToken];
	}

	function getorder(address baseToken, address quoteToken, uint orderId) external view returns (Order memory) {
		return orders[baseToken][quoteToken][orderId];
	}
}
