pragma solidity ^0.4.24;

// Enodeinfo
contract Enodeinfo {

    struct Enode {
        bytes32 id1;// nodeid info about the pervious 32 bytes
        bytes32 id2;// nodeid info about the post 32 bytes
        bytes8 nextId;// the nextId for the next nodeid
        uint64 ipport;// ipport means ip and port info,
                        // encode by ip and port
    }

    mapping (bytes8 => Enode ) public Enodes;
    // Masternode contract address
    address public MasterAddr = 0x000000000000000000000000000000000000000a;
    // count means the number of the register node info in
    // Enodeinfo smart contact
    uint public count;
    // the initial lastId,mean
    bytes8 public lastId;
    constructor() public {
    }

    function register(bytes32 id1, bytes32 id2, uint64 ipport) public {
        require(
             msg.value == 0 &&
             id1 != bytes32(0) &&
             id2 != bytes32(0) &&
             ipport != uint64(0)
             );
        bytes32[2] memory input;
        bytes32[1] memory output;
        input[0] = id1;
        input[1] = id2;

        assembly {
            if  iszero(call(not(0), 0x0b, 0, input, 128, output, 32)) {
                revert(0, 0)
            }
        }
        // this is the account should be generated from the masternode publickey ,
        // it means that the the masternode send his own enode info
        address account = address(output[0]);

        // should be the valid enode id,
        // masternode can only send his own block
        require(
                account != address(0) &&
                address(msg.sender)==address(account)
                );

        bool isMasternode;

        bytes8 id = bytes8(id1);

        require(id != bytes8(0));

        isMasternode = Masternode(MasterAddr).has(id); // id means whether it is masternode or not

        require(bool(isMasternode) == bool(true));


        // save to Enodes
        // head insert to the link

        if (Enodes[id].id1 == bytes32(0)){
            count += 1;
        }

        Enodes[id] = Enode(id1, id2,lastId,ipport);


        if(lastId != bytes8(0)){
            Enodes[lastId].nextId = id;
        }
        lastId = id;
    }

    // find single enode info
    function getSingleEnode(bytes8 id) constant public returns (
        bytes32 id1 ,
        bytes32 id2 ,
        bytes8 nextId,
        uint64 ipport
    )
    {
        require( id != bytes8(0));
        id1 = Enodes[id].id1;
        id2 = Enodes[id].id2;
        nextId = Enodes[id].nextId;
        ipport = Enodes[id].ipport;
    }
    // getCount number
    function getCount() constant public returns (uint)
    {
        return count;
    }
}

contract Masternode {
    function has(bytes8 ) constant public returns (bool ){}
}