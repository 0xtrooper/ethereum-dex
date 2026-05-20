// Copyright © 2026 Kedar Iyer
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

pragma solidity ^0.8.20;

import "../OrderBook.sol";
import "./libraries/SafeERC20.sol";

using SafeERC20 for IERC20;

contract ProxyWithFee {
	uint immutable FEE_NUMERATOR;
	uint immutable FEE_DENOMINATOR;
	address immutable ORDERBOOK_CONTRACT_ADDRESS; 

	constructor(uint fee_numerator, uint fee_denominator, address orderbook_contract) {
		FEE_NUMERATOR = fee_numerator;
		FEE_DENOMINATOR = fee_denominator;
		ORDERBOOK_CONTRACT_ADDRESS = orderbook_contract;
	}

	function fillOrderWithFee(uint orderId, address baseToken, address quoteToken, uint baseQuantity) external payable {
		OrderBook.Order memory order = OrderBook.getorder(baseToken, quoteToken, orderId);
		OrderBook(ORDERBOOK_CONTRACT_ADDRESS).fillOrder{value:msg.value}(orderId, baseToken, quoteToken, baseQuantity);
		if (order.side == OrderBook.Side.SELL) {
		} else {
		}
	}
}
