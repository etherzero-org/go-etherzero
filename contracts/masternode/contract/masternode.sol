pragma solidity ^0.4.11;

contract Masternode {

    uint public constant etzPerNode = 20 * 10 ** 18;

    bytes32 public lastId;
    uint public count;

    struct info {
        bytes32 subId;
        bytes32 misc;
        bytes32 preId;
        bytes32 nextId;
        uint block;
        address account;
    }

    mapping (address => bytes32) ids;
    mapping (bytes32 => info) store;

    event join(bytes32 id);
    event quit(bytes32 id);

    function MasterNode() public {
        lastId = bytes32(0);
        count = 0;
    }

    function register(bytes32 id, bytes32 subId, bytes32 misc) payable public {
        require(
            bytes32(0) != id &&
            bytes32(0) != subId &&
            bytes32(0) != misc &&
            bytes32(0) == ids[msg.sender] &&
            bytes32(0) == store[id].subId &&
            msg.value == etzPerNode
        );

        ids[msg.sender] = id;

        store[id] = info(
            subId,
            misc,
            lastId,
            bytes32(0),
            block.number,
            msg.sender
        );

        if(lastId != bytes32(0)){
            store[lastId].nextId = id;
        }
        lastId = id;
        count += 1;
        emit join(id);
    }

    function() payable public {
        bytes32 id = ids[msg.sender];
        require(
            msg.value == 0 &&
            bytes32(0) != id &&
            bytes32(0) != store[id].subId &&
            address(this).balance >= etzPerNode &&
            count > 0
        );

        bytes32 preId = store[id].preId;
        bytes32 nextId = store[id].nextId;
        if(preId != bytes32(0)){
            store[preId].nextId = nextId;
        }
        if(nextId != bytes32(0)){
            store[nextId].preId = preId;
        }else{
            lastId = preId;
        }
        store[id] = info(
            bytes32(0),
            bytes32(0),
            bytes32(0),
            bytes32(0),
            uint(0),
            address(0)
        );
        ids[msg.sender] = bytes32(0);
        count -= 1;
        emit quit(id);
        msg.sender.transfer(etzPerNode);
    }

    function getInfo(bytes32 id) constant public returns (
        bytes32 subId,
        bytes32 misc,
        bytes32 preId,
        bytes32 nextId,
        uint blockNumber,
        address account
    )
    {
        subId = store[id].subId;
        misc = store[id].misc;
        preId = store[id].preId;
        nextId = store[id].nextId;
        blockNumber = store[id].block;
        account = store[id].account;
    }

    function getId(address addr) constant public returns (bytes32 id)
    {
        id = ids[addr];
    }

    function has(bytes32 id) constant public returns (bool)
    {
        return store[id].subId != bytes32(0);
    }

}