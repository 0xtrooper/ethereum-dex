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

	function withdrawTo(address user, address token, uint amount) external {
		require(msg.sender == owner, "only owner can withdraw funds");
		require(amount > 0, "amount is zero");

		IERC20(token).safeTransfer(user, amount);
	}
}

contract MatchingOrderBook {
	enum Side {
		BUY,
		SELL
	}
	struct Order {
		address user;
		uint baseQuantity;
		uint price;
		uint nextOrderId;
	}
	mapping(address => mapping(address => address payable)) banks; // baseToken -> quoteToken -> bankAddress
	mapping(address => mapping(address => mapping(Side => uint))) orderbooks; // baseToken -> quoteToken -> Side -> firstOrderId
	mapping(address => mapping(address => mapping(Side => mapping(uint => Order)))) orders; // baseToken -> quoteToken -> Side -> orderId -> Order
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
		address payable bankAddress = banks[baseToken][quoteToken];
		require(bankAddress != address(0), "createMarket before placing an order on it");
		require(baseQuantity > 0 && price > 0, "zero quantity/price orders not permitted");

		(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(quoteToken).tryGetDecimals();
		require(decimalCallSuccess, "failed to get decimals for token");
		uint quoteQuantity = baseQuantity * price / quoteTokenDecimals;
		require(quoteQuantity > 0, "calculated quote quantity is zero");

		require(msg.value == 0, "Cannot send ETH. Use WETH instead.");
		if (side == Side.SELL) {
			IERC20 baseTokenIERC20 = IERC20(baseToken);
			uint beforeBalance = baseTokenIERC20.balanceOf(bankAddress);
			baseTokenIERC20.safeTransferFrom(msg.sender, bankAddress, baseQuantity);
			uint afterBalance = baseTokenIERC20.balanceOf(bankAddress);
			uint transferredBaseQuantity = afterBalance - beforeBalance;
			if (transferredBaseQuantity != baseQuantity) {
				baseQuantity = transferredBaseQuantity;
			}
		} else if (side == Side.BUY) {
			IERC20 quoteTokenIERC20 = IERC20(quoteToken);
			uint beforeBalance = quoteTokenIERC20.balanceOf(bankAddress);
			quoteTokenIERC20.safeTransferFrom(msg.sender, bankAddress, quoteQuantity);
			uint afterBalance = quoteTokenIERC20.balanceOf(bankAddress);
			uint transferredQuoteQuantity = afterBalance - beforeBalance;
			if (transferredQuoteQuantity != quoteQuantity) {
				quoteQuantity = transferredQuoteQuantity;
			}
		}

		// Order Matching
		if (side == Side.SELL) {
			uint fillOrderId = orderbooks[baseToken][quoteToken][Side.BUY];
			Order memory fillOrder = orders[baseToken][quoteToken][Side.BUY][fillOrderId];

			// Filling Opposite Book
			while (price <= fillOrder.price) {
				if (baseQuantity == fillOrder.baseQuantity) {
					delete orders[baseToken][quoteToken][Side.BUY][fillOrderId];
					orderbooks[baseToken][quoteToken][Side.BUY] = fillOrder.nextOrderId;
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillOrderQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, baseQuantity);
					break;
				}
				else if (baseQuantity > fillOrder.baseQuantity) {
					delete orders[baseToken][quoteToken][Side.BUY][fillOrderId];
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillOrderQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, fillOrder.baseQuantity);
					baseQuantity -= fillOrder.baseQuantity;
					fillOrder = orders[baseToken][quoteToken][Side.BUY][fillOrder.nextOrderId];
				}
				else { // baseQuantity < fillOrder.baseQuantity
					orderbooks[baseToken][quoteToken][Side.BUY] = fillOrderId;
					orders[baseToken][quoteToken][Side.BUY][fillOrderId].baseQuantity -= baseQuantity;
					emit OrderFill(fillOrderId, baseQuantity);
					uint fillQuoteQuantity = baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, baseQuantity);
					baseQuantity = 0;
					return 0;
				}
			}

			// Place leftover orders in book
			uint nextOrderId = orderbooks[baseToken][quoteToken][Side.SELL];
			while (orders[baseToken][quoteToken][side][nextOrderId].price <= price) {
				nextOrderId = orders[baseToken][quoteToken][side][nextOrderId].nextOrderId;
			}
			unchecked {
				orderId = ++orderCounter;
			}
			orders[baseToken][quoteToken][side][orderId] = Order(msg.sender, baseQuantity, price, nextOrderId);
		} else if (side == Side.BUY) {
			uint fillOrderId = orderbooks[baseToken][quoteToken][Side.SELL];
			Order memory fillOrder = orders[baseToken][quoteToken][Side.SELL][fillOrderId];

			// Filling Opposite Book
			while (price >= fillOrder.price) {
				if (baseQuantity == fillOrder.baseQuantity) {
					delete orders[baseToken][quoteToken][Side.SELL][fillOrderId];
					orderbooks[baseToken][quoteToken][Side.SELL] = fillOrder.nextOrderId;
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillOrderQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, baseQuantity);
					break;
				}
				else if (baseQuantity > fillOrder.baseQuantity) {
					delete orders[baseToken][quoteToken][Side.SELL][fillOrderId];
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillOrderQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, fillOrder.baseQuantity);
					baseQuantity -= fillOrder.baseQuantity;
					fillOrder = orders[baseToken][quoteToken][Side.SELL][fillOrder.nextOrderId];
				}
				else { // baseQuantity < fillOrder.baseQuantity
					orderbooks[baseToken][quoteToken][Side.SELL] = fillOrderId;
					orders[baseToken][quoteToken][Side.SELL][fillOrderId].baseQuantity -= baseQuantity;
					emit OrderFill(fillOrderId, baseQuantity);
					uint fillQuoteQuantity = baseQuantity * fillOrder.price;
					Bank(bankAddress).withdrawTo(msg.sender, quoteToken, fillQuoteQuantity);
					Bank(bankAddress).withdrawTo(fillOrder.user, baseToken, baseQuantity);
					baseQuantity = 0;
					return 0;
				}
			}

			// Place leftover orders in book
			uint nextOrderId = orderbooks[baseToken][quoteToken][Side.SELL];
			while (orders[baseToken][quoteToken][side][nextOrderId].price >= price) {
				nextOrderId = orders[baseToken][quoteToken][side][nextOrderId].nextOrderId;
			}
			unchecked {
				orderId = ++orderCounter;
			}
			orders[baseToken][quoteToken][side][orderId] = Order(msg.sender, baseQuantity, price, nextOrderId);
		}

		bytes32 markethash = keccak256(abi.encodePacked(baseToken, quoteToken));
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, markethash, side, baseQuantity, price);
	}

	function cancelOrder(address baseToken, address quoteToken, Side side, uint orderId) external {
		Order memory order = orders[baseToken][quoteToken][side][orderId];
		require(msg.sender == order.user, "users can only cancel their own order / order may not exist");
		delete orders[baseToken][quoteToken][side][orderId];
		address payable bankAddress = banks[baseToken][quoteToken];

		// Bank.withdrawTo is gas limited so there is minimal re-entrancy risk, but orders are being deleted
		// before withdraws anyway to prevent hijinks within the 2300 forwarded gas 
		if (side == Side.SELL) {
			Bank(bankAddress).withdrawTo(order.user, baseToken, order.baseQuantity);
		} else if (side == Side.BUY) {
			(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(quoteToken).tryGetDecimals();
			require(decimalCallSuccess, "failed to get decimals for token");
			uint quoteQuantity = order.baseQuantity * order.price / quoteTokenDecimals;
			Bank(bankAddress).withdrawTo(order.user, quoteToken, quoteQuantity);
		}
		emit OrderCanceled(orderId);
	}

	function getBankAddress(address baseToken, address quoteToken) public view returns (address payable) {
		return banks[baseToken][quoteToken];
	}

	function getorder(address baseToken, address quoteToken, Side side, uint orderId) external view returns (Order memory) {
		return orders[baseToken][quoteToken][side][orderId];
	}
}
