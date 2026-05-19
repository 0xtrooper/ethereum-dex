// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Test.sol";
import "../OrderBook.sol"; // modify according to your actual file structure
import "src/MockERC20.sol"; // Mock ERC20 token for testing

contract OrderBookTest is Test {
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
		orderBook.placeOrder(address(EURT), address(USDC), OrderBook.Side.SELL, 115 * 1e6, 100 * 1e6);
	}

	function testCreateMarket() public {
		orderBook.createMarket(address(EURT), address(USDC));
	}
	
	function testBankWithdrawToFail() public {
		orderBook.createMarket(address(EURT), address(USDC));
		bytes32 bankHash = orderBook.bankhash(address(EURT), address(USDC));
		address payable bank = orderBook.banks(bankHash);
		vm.expectRevert("only owner can withdraw funds");
		Bank(bank).withdrawTo(user1, address(0), 1e18);
	}

	function testCreateMarketTwiceAndFail() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.expectRevert("market has already been created");
		orderBook.createMarket(address(USDC), address(EURT));
	}
	
	function testPlaceOrder() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 100e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.quoteQuantity, 115 * 1e18, "Quote Quantity should match");
	}

	function testBankHash() public view {
		assertEq(bytes32(0xead70e2fa0ea60acdf1562c2bc09f0525cb2add475c45b55a0b296a8ea02a8f4), orderBook.bankhash(address(0), address(USDC)));
		assertEq(bytes32(0xdfa8f666370cc44e5e24a578c2d2e8407b7753e2f90ff46cc10a2e84e8a2a071), orderBook.bankhash(address(EURT), address(USDC)));
	}

	function testPlaceETHOrder() public {
		orderBook.createMarket(address(0), address(USDC));
		vm.deal(user1, 2e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder{value: 1e18}(address(0), address(USDC), OrderBook.Side.SELL, 1e18, 2150e6);
	        //OrderBook.Order memory order = orderBook.getorder(address(0), address(USDC), orderId);
		//assertEq(order.user, user1, "User should match");
		//assertEq(order.baseQuantity, 1 * 1e18, "Base Quantity should match");
		//assertEq(order.quoteQuantity, 2150 * 1e6, "Quote Quantity should match");
	}

	function testPlace2OrdersForGasDifference() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.quoteQuantity, 115 * 1e18, "Quote Quantity should match");
		vm.prank(user1);
		orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	}

	function testCancelOrder() public {
		orderBook.createMarket(address(USDC), address(EURT));
		vm.prank(user1);
		USDC.approve(address(orderBook), 300e18);
		USDC.mint(user1, 1000*1e18);
		vm.prank(user1);
		uint orderId = orderBook.placeOrder(address(USDC), address(EURT), OrderBook.Side.SELL, 100 * 1e18, 115 * 1e18);
	        OrderBook.Order memory order = orderBook.getorder(address(USDC), address(EURT), orderId);
		assertEq(order.user, user1, "User should match");
		assertEq(order.baseQuantity, 100 * 1e18, "Base Quantity should match");
		assertEq(order.quoteQuantity, 115 * 1e18, "Quote Quantity should match");
		vm.prank(user1);
		orderBook.cancelOrder(address(USDC), address(EURT), orderId);
	}
	
}
