pragma solidity ^0.4.11;

contract Masternode {

    uint public constant etzPerNode = 20 * 10 ** 18;
    uint public constant etzMin = 10 ** 16;
    uint public constant blockPingTimeout = 360;

    bytes8 public lastId;
    uint public count;

    address public governanceAddress;

    struct vote {
        uint vote;
        uint startBlock;
        uint stopBlock;
        address creator;
    }

    mapping(address => mapping(address => bool)) voters;
    mapping (address => vote) votes;

    struct node {
        bytes32 id1;
        bytes32 id2;
        bytes32 misc;
        bytes8 preId;
        bytes8 nextId;
        uint block;
        address account;

        uint blockOnlineAcc;
        uint blockLastPing;
    }

    mapping (address => bytes8) ids;
    mapping (bytes8 => node) nodes;

    event join(bytes8 id, address addr);
    event quit(bytes8 id, address addr);

    constructor() public {
        lastId = bytes8(0);
        count = 0;
    }


    function createGovernanceAddressVote(address addr) payable public
    {
        require(votes[addr].vote == 0 && votes[addr].startBlock == 0);
        votes[addr] = vote(0, block.number, 0, msg.sender);
    }

    function voteForGovernanceAddress(address addr) public
    {
        vote storage v = votes[addr];
        require(v.startBlock > 0
        && getId(msg.sender) != bytes8(0)
        && v.stopBlock == 0
        && voters[addr][msg.sender] == false);
        voters[addr][msg.sender] = true;
        v.vote += 1;
        if (v.vote >= (count * 66 / 100))
        {
            v.stopBlock = block.number;
            governanceAddress = addr;
        }
    }

    function register(bytes32 id1, bytes32 id2, bytes32 misc, address account) payable public {
        bytes8 id = bytes8(id1);
        require(
            bytes32(0) != id1 &&
            bytes32(0) != id2 &&
            bytes32(0) != misc &&
            bytes8(0) != id &&
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
            msg.sender,
            uint(0),
            uint(0)
        );

        if(lastId != bytes8(0)){
            nodes[lastId].nextId = id;
        }
        lastId = id;
        count += 1;
        account.transfer(etzMin);
        emit join(id, msg.sender);
    }

    function() payable public {
        bytes8 id = ids[msg.sender];
        bytes32 id1 = nodes[id].id1;
        require(
            msg.value == 0 &&
            bytes8(0) != id &&
            bytes32(0) != id1 &&
            address(this).balance >= (etzPerNode - etzMin) &&
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
            address(0),
            uint(0),
            uint(0)
        );
        ids[msg.sender] = bytes8(0);
        count -= 1;
        emit quit(id, msg.sender);
        msg.sender.transfer(etzPerNode - etzMin);
    }

    function getInfo(bytes8 id) constant public returns (
        bytes32 id1,
        bytes32 id2,
        bytes32 misc,
        bytes8 preId,
        bytes8 nextId,
        uint blockNumber,
        address account,
        uint blockOnlineAcc,
        uint blockLastPing
    )
    {
        id1 = nodes[id].id1;
        id2 = nodes[id].id2;
        misc = nodes[id].misc;
        preId = nodes[id].preId;
        nextId = nodes[id].nextId;
        blockNumber = nodes[id].block;
        account = nodes[id].account;
        blockOnlineAcc = nodes[id].blockOnlineAcc;
        blockLastPing = nodes[id].blockLastPing;
    }

    function getId(address addr) constant public returns (bytes8 id)
    {
        id = ids[addr];
    }

    function has(bytes8 id) constant public returns (bool)
    {
        return nodes[id].id1 != bytes32(0);
    }

    event pingNotice(bytes8 id, uint blockOnlineAcc, uint blockLastPing);
    function ping(uint blockNumber, bytes32 r, bytes32 s, bytes32 v) public returns(bool) {
        require(block.number >= blockNumber && (block.number - blockNumber) < (blockPingTimeout / 2));

        bytes32[4] memory input;
        bytes8[1] memory output;

        input[0] = blockhash(blockNumber);
        input[1] = r;
        input[2] = s;
        input[3] = v;

        assembly {
            if iszero(call(not(0), 0x09, 0, input, 128, output, 32)) {
              revert(0, 0)
            }
        }

        bytes8 id = output[0];
        require(has(id));

        uint blockLastPing = nodes[id].blockLastPing;
        nodes[id].blockLastPing = block.number;
        if(blockLastPing > 0){
            uint blockGap = block.number - blockLastPing;
            if(blockGap > blockPingTimeout){
                nodes[id].blockOnlineAcc = 0;
            }else if(blockGap < (blockPingTimeout / 2)){
                return false;
            }else{
                nodes[id].blockOnlineAcc += blockGap;

            }
        }
        emit pingNotice(id, nodes[id].blockOnlineAcc, block.number);
        return true;
    }

}