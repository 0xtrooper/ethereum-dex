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
	struct MarketDetails {
		address baseToken;
		address quoteToken;
		uint baseMinPostSize;
		uint quoteMinPostSize;
		address payable bankAddress;
	}
	mapping(bytes32 => MarketDetails) public MARKET_DETAILS; // market_id -> MarketDetails
	mapping(address => mapping(address => mapping(Side => uint))) orderbooks; // baseToken -> quoteToken -> Side -> firstOrderId
	mapping(address => mapping(address => mapping(Side => mapping(uint => Order)))) orders; // baseToken -> quoteToken -> Side -> orderId -> Order
	uint public orderCounter = 1;

	event OrderPlaced(uint indexed orderId, address indexed user, address baseToken, address quoteToken, bytes32 indexed markethash, Side side, uint baseQuantity, uint price);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

	function createMarket(address baseToken, address quoteToken, uint baseMinimum, uint quoteMinimum) external {
		bytes32 marketId = getMarketId(baseToken, quoteToken, baseMinimum, quoteMinimum);
		require(MARKET_DETAILS[marketId].bankAddress == address(0), "market has already been created");
		address payable bankAddress = payable(address(new Bank(address(this))));
		MARKET_DETAILS[marketId] = MarketDetails(baseToken, quoteToken, baseMinimum, quoteMinimum, bankAddress);
	}

	function placeOrder(bytes32 marketId, Side side, uint baseQuantity, uint price) external payable returns (uint orderId) {
		MarketDetails memory marketDetails = MARKET_DETAILS[marketId];
		require(marketDetails.bankAddress != address(0), "createMarket before placing an order on it");
		require(baseQuantity > 0 && price > 0, "zero quantity/price orders not permitted");

		(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(marketDetails.quoteToken).tryGetDecimals();
		require(decimalCallSuccess, "failed to get decimals for token");
		uint quoteQuantity = baseQuantity * price / 10**quoteTokenDecimals;
		require(quoteQuantity > 0, "calculated quote quantity is zero");

		require(msg.value == 0, "Cannot send ETH. Use WETH instead.");
		if (side == Side.SELL) {
			IERC20 baseTokenIERC20 = IERC20(marketDetails.baseToken);
			uint beforeBalance = baseTokenIERC20.balanceOf(marketDetails.bankAddress);
			baseTokenIERC20.safeTransferFrom(msg.sender, marketDetails.bankAddress, baseQuantity);
			uint afterBalance = baseTokenIERC20.balanceOf(marketDetails.bankAddress);
			uint transferredBaseQuantity = afterBalance - beforeBalance;
			if (transferredBaseQuantity != baseQuantity) {
				baseQuantity = transferredBaseQuantity;
			}
		} else if (side == Side.BUY) {
			IERC20 quoteTokenIERC20 = IERC20(marketDetails.quoteToken);
			uint beforeBalance = quoteTokenIERC20.balanceOf(marketDetails.bankAddress);
			quoteTokenIERC20.safeTransferFrom(msg.sender, marketDetails.bankAddress, quoteQuantity);
			uint afterBalance = quoteTokenIERC20.balanceOf(marketDetails.bankAddress);
			uint transferredQuoteQuantity = afterBalance - beforeBalance;
			if (transferredQuoteQuantity != quoteQuantity) {
				quoteQuantity = transferredQuoteQuantity;
				price = baseQuantity * 10**quoteTokenDecimals / transferredQuoteQuantity;
			}
		}

		bool filled = false;

		// Order Matching
		Side makerSide = side == Side.SELL ? Side.BUY : Side.SELL;
		if (side == Side.SELL) {
			uint fillOrderId = orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide];
			Order memory fillOrder = orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];

			// Filling Opposite Book
			while (price <= fillOrder.price && fillOrder.user != address(0)) {
				filled = true;
				if (baseQuantity == fillOrder.baseQuantity) {
					delete orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];
					orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide] = fillOrder.nextOrderId;
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillOrderQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, baseQuantity);
					break;
				}
				else if (baseQuantity > fillOrder.baseQuantity) {
					delete orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillOrderQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, fillOrder.baseQuantity);
					baseQuantity -= fillOrder.baseQuantity;
					fillOrder = orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrder.nextOrderId];
				}
				else { // baseQuantity < fillOrder.baseQuantity
					orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide] = fillOrderId;
					orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId].baseQuantity -= baseQuantity;
					emit OrderFill(fillOrderId, baseQuantity);
					uint fillQuoteQuantity = baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, baseQuantity);
					baseQuantity = 0;
					return 0;
				}
			}

			uint remainingQuoteQuantity = baseQuantity * price / 10**quoteTokenDecimals;
			bool canPost = baseQuantity > marketDetails.baseMinPostSize && remainingQuoteQuantity >  marketDetails.baseMinPostSize;
			require(filled || canPost, "fill or kill. didn't fill");

			// refund orders which filled but can't post
			if (filled && !canPost) {
				Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.baseToken, baseQuantity);
				return 0;
			}

			// Place leftover orders in book
			unchecked {
				orderId = ++orderCounter;
			}
			uint nextOrderId = orderbooks[marketDetails.baseToken][marketDetails.quoteToken][Side.SELL];
			if (nextOrderId == 0) {
				orderbooks[marketDetails.baseToken][marketDetails.quoteToken][Side.SELL] = orderId;
			}
			else {
				uint previousOrderId = 0;
				while (nextOrderId != 0 && orders[marketDetails.baseToken][marketDetails.quoteToken][side][nextOrderId].price <= price) {
					previousOrderId = nextOrderId;
					nextOrderId = orders[marketDetails.baseToken][marketDetails.quoteToken][side][nextOrderId].nextOrderId;
				}
				orders[marketDetails.baseToken][marketDetails.quoteToken][side][previousOrderId].nextOrderId = orderId;
			}
			orders[marketDetails.baseToken][marketDetails.quoteToken][side][orderId] = Order(msg.sender, baseQuantity, price, nextOrderId);
		} else if (side == Side.BUY) {
			uint fillOrderId = orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide];
			Order memory fillOrder = orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];

			// Filling Opposite Book
			while (price >= fillOrder.price && fillOrder.user != address(0)) {
				filled = true;
				if (baseQuantity == fillOrder.baseQuantity) {
					delete orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];
					orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide] = fillOrder.nextOrderId;
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillOrderQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, baseQuantity);
					break;
				}
				else if (baseQuantity > fillOrder.baseQuantity) {
					delete orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId];
					emit OrderFill(fillOrderId, fillOrder.baseQuantity);
					uint fillOrderQuoteQuantity = fillOrder.baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillOrderQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, fillOrder.baseQuantity);
					baseQuantity -= fillOrder.baseQuantity;
					fillOrder = orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrder.nextOrderId];
				}
				else { // baseQuantity < fillOrder.baseQuantity
					orderbooks[marketDetails.baseToken][marketDetails.quoteToken][makerSide] = fillOrderId;
					orders[marketDetails.baseToken][marketDetails.quoteToken][makerSide][fillOrderId].baseQuantity -= baseQuantity;
					emit OrderFill(fillOrderId, baseQuantity);
					uint fillQuoteQuantity = baseQuantity * fillOrder.price / 10**quoteTokenDecimals;
					Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, fillQuoteQuantity);
					Bank(marketDetails.bankAddress).withdrawTo(fillOrder.user, marketDetails.baseToken, baseQuantity);
					baseQuantity = 0;
					return 0;
				}
			}

			uint remainingQuoteQuantity = baseQuantity * price / 10**quoteTokenDecimals;
			bool canPost = baseQuantity > marketDetails.baseMinPostSize && remainingQuoteQuantity >  marketDetails.baseMinPostSize;
			require(filled || canPost, "fill or kill. didn't fill");

			// refund orders which filled but can't post
			if (filled && !canPost) {
				Bank(marketDetails.bankAddress).withdrawTo(msg.sender, marketDetails.quoteToken, quoteQuantity);
				return 0;
			}

			// Place leftover orders in book
			unchecked {
				orderId = ++orderCounter;
			}
			uint nextOrderId = orderbooks[marketDetails.baseToken][marketDetails.quoteToken][Side.BUY];
			if (nextOrderId == 0) {
				orderbooks[marketDetails.baseToken][marketDetails.quoteToken][Side.BUY] = orderId;
			}
			else {
				uint previousOrderId = 0;
				while (nextOrderId != 0 && orders[marketDetails.baseToken][marketDetails.quoteToken][side][nextOrderId].price >= price) {
					previousOrderId = nextOrderId;
					nextOrderId = orders[marketDetails.baseToken][marketDetails.quoteToken][side][nextOrderId].nextOrderId;
				}
				orders[marketDetails.baseToken][marketDetails.quoteToken][side][previousOrderId].nextOrderId = orderId;
			}
			orders[marketDetails.baseToken][marketDetails.quoteToken][side][orderId] = Order(msg.sender, baseQuantity, price, nextOrderId);
		}

		bytes32 markethash = keccak256(abi.encodePacked(marketDetails.baseToken, marketDetails.quoteToken));
		emit OrderPlaced(orderId, msg.sender, marketDetails.baseToken, marketDetails.quoteToken, markethash, side, baseQuantity, price);
	}

	function cancelOrder(bytes32 marketId, Side side, uint orderId) external {
		MarketDetails memory marketDetails = MARKET_DETAILS[marketId];
		Order memory order = orders[marketDetails.baseToken][marketDetails.quoteToken][side][orderId];
		require(msg.sender == order.user, "users can only cancel their own order / order may not exist");
		delete orders[marketDetails.baseToken][marketDetails.quoteToken][side][orderId];

		// Re-entrancy here is limited to malicious tokens
		if (side == Side.SELL) {
			Bank(marketDetails.bankAddress).withdrawTo(order.user, marketDetails.baseToken, order.baseQuantity);
		} else if (side == Side.BUY) {
			(bool decimalCallSuccess, uint8 quoteTokenDecimals) = IERC20(marketDetails.quoteToken).tryGetDecimals();
			require(decimalCallSuccess, "failed to get decimals for token");
			uint quoteQuantity = order.baseQuantity * order.price / 10**quoteTokenDecimals;
			Bank(marketDetails.bankAddress).withdrawTo(order.user, marketDetails.quoteToken, quoteQuantity);
		}
		emit OrderCanceled(orderId);
	}

	function getMarketId(address baseToken, address quoteToken, uint baseMinimum, uint quoteMinimum) public pure returns (bytes32) {
		return keccak256(abi.encodePacked(baseToken, quoteToken, baseMinimum, quoteMinimum));
	}

	function getMarketDetails(bytes32 marketId) public view returns (MarketDetails memory) {
		return MARKET_DETAILS[marketId];
	}

	function getOrder(address baseToken, address quoteToken, Side side, uint orderId) external view returns (Order memory) {
		return orders[baseToken][quoteToken][side][orderId];
	}

	function getFirstOrderId(address baseToken, address quoteToken, Side side) external view returns (uint) {
		return orderbooks[baseToken][quoteToken][side];
	}

	function getOrderBook(address baseToken, address quoteToken, Side side, uint depth) external view returns (Order[] memory) {
		Order[] memory returnOrders = new Order[](depth); 
		uint orderId = orderbooks[baseToken][quoteToken][side];
		for (uint i=0; i < depth; i++) {
			returnOrders[i] = orders[baseToken][quoteToken][side][orderId];
			orderId = orders[baseToken][quoteToken][side][orderId].nextOrderId;
		}
		return returnOrders;
	}
}
