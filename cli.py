import argparse
import sys
from web3 import Web3

CONTRACT_ADDRESS = "0x0000000000000000000000000000000000000000"
ABI = []

def instantiate_contract(provider_url, contract_address, abi):
    """
    Instantiate a Web3 contract instance.

    :param provider_url: The Ethereum node provider URL (e.g. HTTP or IPC).
    :param contract_address: The address of the contract to interact with.
    :param abi: The ABI of the contract (as a list/dict).
    :return: A web3.contract.Contract instance.
    """
    web3 = Web3(Web3.HTTPProvider(provider_url))
    if not web3.isConnected():
        raise ConnectionError("Web3 provider is not connected.")

    contract = web3.eth.contract(address=web3.to_checksum_address(contract_address), abi=abi)
    return contract


def placeorder(args):
    print(f"Placing order: Product={args.product}, Quantity={args.quantity}, Price={args.price}")

def cancelorder(args):
    print(f"Cancelling order: OrderID={args.orderid}")

def fillorder(args):
    print(f"Filling order: OrderID={args.orderid}")

def main():
    parser = argparse.ArgumentParser(description="Command line client for order operations.")
    parser.add_argument('--rpc', default='http://localhost:8545', help='Ethereum node RPC endpoint (default: http://localhost:8545)')
    subparsers = parser.add_subparsers(dest='command', required=True, help='Available commands')
    
    # placeorder
    parser_place = subparsers.add_parser('placeorder', help='Place a new order')
    parser_place.add_argument('--product', required=True, help='Product name')
    parser_place.add_argument('--quantity', required=True, type=int, help='Quantity to order')
    parser_place.add_argument('--price', required=True, type=float, help='Order price')
    parser_place.set_defaults(func=placeorder)
    
    # cancelorder
    parser_cancel = subparsers.add_parser('cancelorder', help='Cancel an order')
    parser_cancel.add_argument('--orderid', required=True, help='Order ID to cancel')
    parser_cancel.set_defaults(func=cancelorder)
    
    # fillorder
    parser_fill = subparsers.add_parser('fillorder', help='Fill an order')
    parser_fill.add_argument('--orderid', required=True, help='Order ID to fill')
    parser_fill.set_defaults(func=fillorder)
    
    args = parser.parse_args()
    args.func(args)

if __name__ == "__main__":
    main()
