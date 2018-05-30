pragma solidity ^0.4.11;

contract Masternode {

    uint public constant etzPerNode = 20 * 10 ** 18;

    bytes8 public lastId;
    uint public count;

    struct node {
        bytes32 id1;
        bytes32 id2;
        bytes32 misc;
        bytes8 preId;
        bytes8 nextId;
        uint block;
        address account;
    }

    mapping (address => bytes8) ids;
    mapping (bytes8 => node) nodes;

    event join(bytes8 id, address addr);
    event quit(bytes8 id, address addr);

    constructor() public {
        lastId = bytes8(0);
        count = 0;
    }

    function register(bytes32 id1, bytes32 id2, bytes32 misc) payable public {
        bytes8 id = bytes8(id1);
        require(
            bytes32(0) != id1 &&
            bytes32(0) != id2 &&
            bytes32(0) != misc &&
            bytes8(0) == ids[msg.sender] &&
            bytes32(0) == nodes[id].id1 &&
            msg.value == etzPerNode
        );

        ids[msg.sender] = id;

        nodes[id] = node(
            id1,
            id2,
            misc,
            lastId,
            bytes8(0),
            block.number,
            msg.sender
        );

        if(lastId != bytes8(0)){
            nodes[lastId].nextId = id;
        }
        lastId = id;
        count += 1;
        emit join(id, msg.sender);
    }

    function() payable public {
        bytes8 id = ids[msg.sender];
        bytes32 id1 = nodes[id].id1;
        require(
            msg.value == 0 &&
            bytes8(0) != id &&
            bytes32(0) != id1 &&
            address(this).balance >= etzPerNode &&
            count > 0
        );

        bytes8 preId = nodes[id].preId;
        bytes8 nextId = nodes[id].nextId;
        if(preId != bytes8(0)){
            nodes[preId].nextId = nextId;
        }
        if(nextId != bytes8(0)){
            nodes[nextId].preId = preId;
        }else{
            lastId = preId;
        }
        nodes[id] = node(
            bytes32(0),
            bytes32(0),
            bytes32(0),
            bytes8(0),
            bytes8(0),
            uint(0),
            address(0)
        );
        ids[msg.sender] = bytes8(0);
        count -= 1;
        emit quit(id, msg.sender);
        msg.sender.transfer(etzPerNode);
    }

    function getInfo(bytes8 id) constant public returns (
        bytes32 id1,
        bytes32 id2,
        bytes32 misc,
        bytes8 preId,
        bytes8 nextId,
        uint blockNumber,
        address account
    )
    {
        id1 = nodes[id].id1;
        id2 = nodes[id].id2;
        misc = nodes[id].misc;
        preId = nodes[id].preId;
        nextId = nodes[id].nextId;
        blockNumber = nodes[id].block;
        account = nodes[id].account;
    }

    function getId(address addr) constant public returns (bytes8 id)
    {
        id = ids[addr];
    }

    function has(bytes8 id) constant public returns (bool)
    {
        return nodes[id].id1 != bytes32(0);
    }

}