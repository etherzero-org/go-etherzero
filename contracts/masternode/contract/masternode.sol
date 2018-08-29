pragma solidity ^0.4.11;

contract Masternode {

    uint public constant etzPerNode = 20000 * 10 ** 18;
    uint public constant etzMin = 10 ** 16;
    uint public constant blockPingTimeout = 3600;

    bytes8 public lastId;
    uint public count;

    struct node {
        bytes32 id1;
        bytes32 id2;
        bytes8 preId;
        bytes8 nextId;
        address account;
        uint block;

        uint blockOnlineAcc;
        uint blockLastPing;
    }
    
    mapping (bytes8 => node) nodes;
    mapping (address => bytes8) ids;
    mapping (address => bytes8) nodeAddressToId;
    
    uint public constant proposalPeriod = 1200000;
    uint public constant proposalFee = 100 * 10 ** 18;
    uint public proposalCount;
    address public governanceAddress;
    address public lastProposalAddress;
    struct vote {
        uint voteCount;
        uint startBlock;
        uint stopBlock;
        address creator;
        address lastAddress;
    }
    
    mapping(address => mapping(address => bool)) voters;
    mapping (address => vote) votes;

    event join(bytes8 id, address addr);
    event quit(bytes8 id, address addr);
    event ping(bytes8 id, uint blockOnlineAcc, uint blockLastPing);
    event newVote(address from, address to);
    event newProposal(address from, address to);
    event governanceAddressChange(address from, address to);

    constructor() public {
    }

    function register(bytes32 id1, bytes32 id2) payable public {
        bytes8 id = bytes8(id1);
        require(
            bytes32(0) != id1 &&
            bytes32(0) != id2 &&
            bytes8(0) != id &&
            bytes8(0) == ids[msg.sender] &&
            bytes32(0) == nodes[id].id1 &&
            msg.value == etzPerNode
        );
        bytes32[2] memory input;
        bytes32[1] memory output;
        input[0] = id1;
        input[1] = id2;
        assembly {
            if iszero(call(not(0), 0x0b, 0, input, 128, output, 32)) {
              revert(0, 0)
            }
        }
        address account = address(output[0]);
        require(account != address(0));
        
        ids[msg.sender] = id;
        nodes[id] = node(
            id1,
            id2,
            lastId,
            bytes8(0),
            msg.sender,
            block.number,
            uint(0),
            uint(0)
        );
        if(lastId != bytes8(0)){
            nodes[lastId].nextId = id;
        }
        lastId = id;
        count += 1;
        nodeAddressToId[account] = id;
        account.transfer(etzMin);
        emit join(id, msg.sender);
    }

    function() payable public {
        require(msg.value == 0);
        bytes8 id = nodeAddressToId[msg.sender];
        if (id != bytes8(0) && has(id)){
            // ping
            uint blockLastPing = nodes[id].blockLastPing;
            if(blockLastPing > 0){
                uint blockGap = block.number - blockLastPing;
                if(blockGap > blockPingTimeout){
                    nodes[id].blockOnlineAcc = 0;
                }else{
                    nodes[id].blockOnlineAcc += blockGap;
                }
            }
            nodes[id].blockLastPing = block.number;
            emit ping(id, nodes[id].blockOnlineAcc, block.number);
        }else{
            id = ids[msg.sender];
            bytes32 id1 = nodes[id].id1;
            require(
                msg.value == 0 &&
                bytes8(0) != id &&
                bytes32(0) != id1 &&
                address(this).balance >= (etzPerNode - etzMin) &&
                count > 0
            );
            
            bytes32[2] memory input;
            bytes32[1] memory output;
            input[0] = id1;
            input[1] = nodes[id].id2;
            assembly {
                if iszero(call(not(0), 0x0b, 0, input, 128, output, 32)) {
                  revert(0, 0)
                }
            }
            address account = address(output[0]);
            nodeAddressToId[account] = bytes8(0);
            
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
            bool notGenesisNode = nodes[id].block > 0;
            nodes[id] = node(
                bytes32(0),
                bytes32(0),
                bytes8(0),
                bytes8(0),
                address(0),
                uint(0),
                uint(0),
                uint(0)
            );
            ids[msg.sender] = bytes8(0);
            count -= 1;
            emit quit(id, msg.sender);
            if(notGenesisNode){
                msg.sender.transfer(etzPerNode - etzMin);
            }
        }

    }

    function getInfo(bytes8 id) constant public returns (
        bytes32 id1,
        bytes32 id2,
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

    function initGovernanceAddress(address addr) payable public
    {
        require(0 == proposalCount 
        && address(0) == governanceAddress
        && bytes8(0) != getId(msg.sender));
        governanceAddress = addr;
    }

    function createGovernanceAddressVote(address addr) payable public
    {
        require(0 == votes[addr].voteCount
        && 0 == votes[addr].startBlock
        && msg.value == proposalFee);
        votes[addr] = vote({
            voteCount: 0,
            startBlock: block.number,
            stopBlock: block.number + proposalPeriod,
            creator: msg.sender,
            lastAddress: lastProposalAddress
        });
        lastProposalAddress = addr;
        proposalCount += 1;
        emit newProposal(msg.sender, addr);
    }
    
    function voteForGovernanceAddress(address addr) public
    {
        vote storage v = votes[addr];
        bytes8 id = getId(msg.sender);
        require(v.startBlock > 0 
        && block.number > v.startBlock
        && block.number < v.stopBlock
        && id != bytes8(0)
        && (block.number - nodes[id].block) > proposalPeriod
        && false == voters[addr][msg.sender]);
        voters[addr][msg.sender] = true;
        v.voteCount += 1;
        emit newVote(msg.sender, addr);
        if (v.voteCount > (count / 2))
        {
            v.stopBlock = block.number;
            governanceAddress = addr;
            emit governanceAddressChange(msg.sender, addr);
        }
    }
    
    function getVoteInfo(address addr) constant public returns (
        uint voteCount,
        uint startBlock,
        uint stopBlock,
        address creator,
        address lastAddress
    )
    {
        voteCount = votes[addr].voteCount;
        startBlock = votes[addr].startBlock;
        stopBlock = votes[addr].stopBlock;
        creator = votes[addr].creator;
        lastAddress = votes[addr].lastAddress;
    }
    
    function checkVote(address proposalAddr, address voter) constant public returns(bool)
    {
        return voters[proposalAddr][voter];
    }
}