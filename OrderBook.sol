pragma solidity ^0.8.1;

import "../libraries/SafeERC20.sol";

using SafeERC20 for IERC20;

contract OrderBook {
        enum Side { BUY, SELL }
        struct Order {
		address user;
                uint baseQuantity;
                uint quoteQuantity;
                address baseToken;
                address quoteToken;
                Side side;
        }
	struct MarketBalances {
		uint baseBalance;
		uint quoteBalance;
	}
        mapping(uint => Order) public orders;
        mapping(bytes32 => MarketBalances) public MARKET_BALANCES;
        uint public orderCounter = 0; 

	event OrderPlaced(uint orderId, address indexed user, address indexed baseToken, address indexed quoteToken, Side side, uint baseQuantity, uint quoteQuantity);
	event OrderCanceled(uint indexed orderId);
	event OrderFill(uint indexed orderId, uint baseQuantity);

	function placeOrder (address baseToken, address quoteToken, Side side, uint baseQuantity, uint quoteQuantity) public payable {
		require(baseQuantity > 0 && quoteQuantity > 0, "zero quantity orders not permitted");
		if (side == Side.SELL) {
			if (msg.value > 0) {
				require(baseToken == address(0), "base token should be 0x0 when selling ETH");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}
			else {
				uint beforeBalance = IERC20(baseToken).balanceOf(address(this));
				IERC20(baseToken).safeTransferFrom(msg.sender, address(this), baseQuantity);
				uint afterBalance = IERC20(baseToken).balanceOf(address(this));
				require(afterBalance - beforeBalance == baseQuantity, "token error: tokens that charge transfer fees are not permitted");
			}
			MARKET_BALANCES[marketHash(baseToken, quoteToken)].baseBalance += baseQuantity;
		}
		else if (side == Side.BUY) {
			if (msg.value > 0) {
				require(quoteToken == address(0), "quote token should be 0x0 when buying with ETH");
				require(quoteQuantity == msg.value, "mismatch between provided quoteQuantity and amount of ETH sent");
			}
			else {
				uint beforeBalance = IERC20(quoteToken).balanceOf(address(this));
				IERC20(quoteToken).safeTransferFrom(msg.sender, address(this), quoteQuantity);
				uint afterBalance = IERC20(quoteToken).balanceOf(address(this));
				require(afterBalance - beforeBalance == quoteQuantity, "token error: tokens that charge transfer fees are not permitted");
			}
			MARKET_BALANCES[marketHash(baseToken, quoteToken)].quoteBalance += quoteQuantity;
		}
                uint orderId = ++orderCounter;
                orders[orderId] = Order(msg.sender, baseQuantity, quoteQuantity, baseToken, quoteToken, side);
		emit OrderPlaced(orderId, msg.sender, baseToken, quoteToken, side, baseQuantity, quoteQuantity);
        }

	function cancelOrder (uint orderId) public {
		Order memory order = orders[orderId];
		require(msg.sender == order.user, "users can only cancel their own order");
                delete orders[orderId];
		if (order.side == Side.SELL) {
			require(order.baseQuantity <= MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].baseBalance, "HACK ALERT: cancel amount is overflowing market balance");
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].baseBalance -= order.baseQuantity;
			if (order.baseToken == address(0)) {
				payable(order.user).transfer(order.baseQuantity);
			} 
			else {
				IERC20(order.baseToken).safeTransfer(msg.sender, order.baseQuantity);
			}
		}
		else if (order.side == Side.BUY) {
			require(order.quoteQuantity <= MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].quoteBalance, "HACK ALERT: cancel amount is overflowing market balance");
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].quoteBalance -= order.quoteQuantity;
			if (order.quoteToken == address(0)) {
				payable(order.user).transfer(order.quoteQuantity);
			} 
			else {
				IERC20(order.quoteToken).safeTransfer(msg.sender, order.quoteQuantity);
			}
		}
		emit OrderCanceled(orderId);
        }
	
	function fillOrder (uint orderId, uint baseQuantity) public payable {
                Order memory order = orders[orderId];
		uint quoteQuantity = baseQuantity * order.quoteQuantity / order.baseQuantity;
		if (msg.value > 0) {
			if (order.side == Side.SELL) {
				require(order.quoteToken == address(0), "quote token should be 0x0");
				require(quoteQuantity == msg.value, "mismatch between quoteQuantity and amount of ETH sent");
			}
			else if (order.side == Side.BUY) {
				require(order.baseToken == address(0), "base token should be 0x0");
				require(baseQuantity == msg.value, "mismatch between provided baseQuantity and amount of ETH sent");
			}
		}
                require(baseQuantity > 0, "zero quantity fills not permitted");
                require(baseQuantity <= order.baseQuantity, "trying to fill more than order size");
                require(baseQuantity <= MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].baseBalance, "HACK ALERT: fill amount is overflowing market balance");
                require(quoteQuantity > 0, "calculated quote quantity is zero");
                require(quoteQuantity <= MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].quoteBalance, "HACK ALERT: fill amount is overflowing market balance");
		orders[orderId].baseQuantity -= baseQuantity;
		orders[orderId].quoteQuantity -= quoteQuantity;
		if (orders[orderId].baseQuantity == 0) {
			delete orders[orderId];
		}
		if (order.side == Side.SELL) {
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].quoteBalance -= quoteQuantity;
			if (order.quoteToken == address(0)) {
				payable(order.user).transfer(quoteQuantity);
			}
			else {
				IERC20(order.quoteToken).safeTransferFrom(msg.sender, order.user, quoteQuantity);
			}
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].baseBalance -= baseQuantity;
			if (order.baseToken == address(0)) {
				payable(msg.sender).transfer(baseQuantity);
			}
			else {
				IERC20(order.baseToken).safeTransfer(msg.sender, baseQuantity);
			}
		}
		else if (order.side == Side.BUY) {
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].baseBalance -= baseQuantity;
			if (order.baseToken == address(0)) {
				payable(order.user).transfer(baseQuantity);
			}
			else {
				IERC20(order.baseToken).safeTransferFrom(msg.sender, order.user, baseQuantity);
			}
			MARKET_BALANCES[marketHash(order.baseToken, order.quoteToken)].quoteBalance -= quoteQuantity;
			if (order.quoteToken == address(0)) {
				payable(msg.sender).transfer(quoteQuantity);
			}
			else {
				IERC20(order.quoteToken).safeTransfer(msg.sender, quoteQuantity);
			}
		}
		emit OrderFill(orderId, baseQuantity);
	}

	function marketHash (address baseToken, address quoteToken) public pure returns (bytes32) {
		return keccak256(abi.encodePacked(baseToken, quoteToken));
	}

}

