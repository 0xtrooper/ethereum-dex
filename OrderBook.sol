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
        mapping(uint => Order) public orders;
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
				bool success = IERC20(baseToken).transferFrom(msg.sender, address(this), baseQuantity);
				require(success, "failed transfer");
			}
		}
		else if (side == Side.BUY) {
			if (msg.value > 0) {
				require(quoteToken == address(0), "quote token should be 0x0 when buying with ETH");
				require(quoteQuantity == msg.value, "mismatch between provided quoteQuantity and amount of ETH sent");
			}
			else {
				bool success = IERC20(quoteToken).transferFrom(msg.sender, address(this), quoteQuantity);
				require(success, "failed transfer");
			}
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
			if (order.baseToken == address(0)) {
				payable(order.user).transfer(order.baseQuantity);
			} 
			else {
				bool success = IERC20(order.baseToken).transfer(msg.sender, order.baseQuantity);
				require(success, "failed transfer");
			}
		}
		else if (order.side == Side.BUY) {
			if (order.quoteToken == address(0)) {
				payable(order.user).transfer(order.quoteQuantity);
			} 
			else {
				bool success = IERC20(order.quoteToken).transfer(msg.sender, order.quoteQuantity);
				require(success, "failed transfer");
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
                require(quoteQuantity > 0, "calculated quote quantity is zero");
		orders[orderId].baseQuantity -= baseQuantity;
		orders[orderId].quoteQuantity -= quoteQuantity;
		if (orders[orderId].baseQuantity == 0) {
			delete orders[orderId];
		}
		if (order.side == Side.SELL) {
			if (order.quoteToken == address(0)) {
				payable(order.user).transfer(quoteQuantity);
			}
			else {
				bool success = IERC20(order.quoteToken).transferFrom(msg.sender, order.user, quoteQuantity);
				require(success, "failed transfer");
			}
			if (order.baseToken == address(0)) {
				payable(msg.sender).transfer(baseQuantity);
			}
			else {
				bool success = IERC20(order.baseToken).transfer(msg.sender, baseQuantity);
				require(success, "failed transfer");
			}
		}
		else if (order.side == Side.BUY) {
			if (order.baseToken == address(0)) {
				payable(order.user).transfer(baseQuantity);
			}
			else {
				bool success = IERC20(order.baseToken).transferFrom(msg.sender, order.user, baseQuantity);
				require(success, "failed transfer");
			}
			if (order.quoteToken == address(0)) {
				payable(msg.sender).transfer(quoteQuantity);
			}
			else {
				bool success = IERC20(order.quoteToken).transfer(msg.sender, quoteQuantity);
				require(success, "failed transfer");
			}
		}
		emit OrderFill(orderId, baseQuantity);
	}

}

