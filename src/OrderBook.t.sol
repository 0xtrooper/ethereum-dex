// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../OrderBook.sol"; // modify according to your actual file structure
import "src/MockERC20.sol"; // Mock ERC20 token for testing

contract BankTest is Test {
	OrderBook orderBook;
	MockERC20 USDC;
	MockERC20 EURT;
	address owner = address(0x1);
	address user1 = address(0x2);
	address user2 = address(0x3);

	function setUp() public {
		USDC = new MockERC20("USDC", "USDC", 6);
		EURT = new MockERC20("EURT", "EURT", 6);
		orderBook = new OrderBook();
		USDC.mint(user1, 10000e6);
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
	}

	function testPlaceOrderBeforeMarketCreation() public {
		vm.expectRevert("createMarket before placing an order on it"); 
		vm.prank(user1);
		orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 100 * 1e18);
	}

	function testCreateMarket() public {
		orderBook.createMarket(address(USDC), address(EURT));
	}
	
	function testPlaceOrder() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	        (address user, uint baseQuantity, uint quoteQuantity, address baseToken, address quoteToken, OrderBook.Side side) = orderBook.orders(orderId);
		assertEq(user, user1, "User should match");
		assertEq(baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(quoteQuantity, 115 * 1e18, "Quote Quantity should match");
	}

	function testPlace2OrdersForGasDifference() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	        (address user, uint baseQuantity, uint quoteQuantity, address baseToken, address quoteToken, OrderBook.Side side) = orderBook.orders(orderId);
		assertEq(user, user1, "User should match");
		assertEq(baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(quoteQuantity, 115 * 1e18, "Quote Quantity should match");
		vm.prank(user1);
		orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	}
	
}
