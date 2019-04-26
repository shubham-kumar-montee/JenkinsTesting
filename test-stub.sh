# Terminal 1

docker-compose -f docker-compose-simple.yaml up


# Terminal 2
cd /opt/go/src/github.com/hyperledger/fabric/fabric-samples/chaincode/loyalty/.
rm -rf *.go
cp /home/owner/src/loyalty_solutions/ChainCode/*.go /opt/go/src/github.com/hyperledger/fabric/fabric-samples/chaincode/loyalty/.
docker exec -it chaincode bash
cd loyalty 
go build

CORE_PEER_ADDRESS=peer:7052 
CORE_CHAINCODE_ID_NAME=loy:0 ./loyalty


# Terminal 3

docker exec -it cli bash
peer chaincode install -p chaincodedev/chaincode/loyalty/ -n loy -v 0


# Setting up the ACL conditions for all the respective functions
peer chaincode instantiate -n loy -v 0 -c '{"Args":["ACL","JSON-str"]}' -C myc
                

peer chaincode invoke -n loy -c '{"Args":["requestRewardPoints","100"]}' -C myc

peer chaincode invoke -n loy -c '{"Args":["getRewardPoints"]}' -C myc

peer chaincode invoke -n loy -c '{"Args":["getAllACLConditions"]}' -C myc

peer chaincode invoke -n loy -c '{"Args":["getACLConditionsByFuncAndOrg","requestRewardPoints","HILTON"]}' -C myc

peer chaincode invoke -n loy -c '{"Args":["approveRequest" ,"REQUEST_1","sunil","01-01-2018"]}' -C myc



#Remove docker containers
for con_id in ` docker ps -a |tail -4| cut -f1 -d " "`; do docker rm $con_id; done


