pragma solidity ^0.4.24;

import "../externals/openzeppelin-solidity/contracts/token/ERC721/ERC721Full.sol";
import "../externals/openzeppelin-solidity/contracts/ownership/Ownable.sol";
import "./INFTReceiver.sol";


contract ServiceChainNFT is ERC721Full("ServiceChainNFT", "SCN"), Ownable {
    address public bridge;

    constructor (address _bridge) public {
        if (!_bridge.isContract()) {
            revert("bridge is not a contract");
        }

        bridge = _bridge;
    }

    // Owner mints the NFT to the user.
    function register(address _user, uint256 _tokenId) onlyOwner external {
        _mint(_user, _tokenId);
    }

    // registerBulk registers (startID, endID-1) NFTs to the user once.
    // This is only for load test.
    function registerBulk(address _user, uint256 _startID, uint256 _endID) onlyOwner external {
        for (uint256 uid = _startID; uid < _endID; uid++) {
            _mint(_user, uid);
        }
    }

    bytes4 private constant _ERC721_RECEIVED = 0x150b7a02;

    // user request value transfer to main / service chain.
    function requestValueTransfer(uint256 _uid, address _to) external {
        transferFrom(msg.sender, bridge, _uid);

        bytes4 retval = INFTReceiver(bridge).onNFTReceived(msg.sender, _uid, _to);
        require(retval == _ERC721_RECEIVED, "Sent to a bridge which is not an ERC721 receiver" );
    }
}
