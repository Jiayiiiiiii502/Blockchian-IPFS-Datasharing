# delete the previous network
docker-compose -f explorer/docker-compose.yaml down -v
docker-compose -f docker-compose-byfn.yaml down -v
rm -rf channel-artifacts
rm -rf crypto-config

# generate certificate
./bin/cryptogen generate --config=./crypto-config.yaml
mkdir channel-artifacts

#generate the first block in the network
./bin/configtxgen -profile TwoOrgsOrdererGenesis -channelID byfn-sys-channel -outputBlock ./channel-artifacts/genesis.block

#config of channel
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID mychannel

#config of node
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP

#config of anchor
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP

docker-compose -f docker-compose-byfn.yaml up -d 

peer0org1="CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer0.org1.example.com:7051"
peer1org1="CORE_PEER_LOCALMSPID="Org1MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp CORE_PEER_ADDRESS=peer1.org1.example.com:8051"
peer0org2="CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer0.org2.example.com:9051"
peer1org2="CORE_PEER_LOCALMSPID="Org2MSP" CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp CORE_PEER_ADDRESS=peer1.org2.example.com:10051"

docker exec cli bash -c "$peer0org1 peer channel create -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/channel.tx"

#add all nodes into the channel
docker exec cli bash -c "$peer0org1 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer1org1 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer0org2 peer channel join -b mychannel.block"
docker exec cli bash -c "$peer1org2 peer channel join -b mychannel.block"

docker exec cli bash -c "$peer0org1 peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org1MSPanchors.tx"
docker exec cli bash -c "$peer0org2 peer channel update -o orderer.example.com:7050 -c mychannel -f ./channel-artifacts/Org2MSPanchors.tx"

# install the chaincode
docker exec cli bash -c "$peer0org1 peer chaincode install -n datashare -v 1.0 -l golang -p github.com/chaincode"
docker exec cli bash -c "$peer0org2 peer chaincode install -n datashare -v 1.0 -l golang -p github.com/chaincode"

docker exec cli bash -c "$peer0org1 peer chaincode instantiate -o orderer.example.com:7050  -C mychannel -n datashare -l golang -v 1.0 -c '{\"Args\":[\"test\",\"链码实例化成功\"]}' -P 'AND ('\''Org1MSP.peer'\'','\''Org2MSP.peer'\'')'"

sleep 5
# query test on peer0org1 to verify the network
docker exec cli bash -c "peer chaincode query -C mychannel -n datashare -c '{\"Args\":[\"queryRecord\",\"test\"]}'"
# invoke test on peer0org2上
docker exec cli bash -c "peer chaincode invoke -o orderer.example.com:7050 -C mychannel -n datashare --peerAddresses peer0.org1.example.com:7051 --peerAddresses peer0.org2.example.com:9051 -c '{\"Args\":[\"sendData\",\"a\",\"b\",\"c\",\"d\",\"e\"]}'"
sleep 5
# install chaincode and query test on peer1org1
docker exec cli bash -c "peer chaincode query -C mychannel -n datashare -c '{\"Args\":[\"queryRecord\",\"a\"]}'"

# start the blockchain network
# substitute keys
priv_sk=$(ls crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore)
cp -rf ./explorer/connection-profile/test-network-temp.json ./explorer/connection-profile/test-network.json
sed -i "s/priv_sk/$priv_sk/" ./explorer/connection-profile/test-network.json
docker-compose -f explorer/docker-compose.yaml up -d

# substitute keys used in the tape
cp -rf ./tape/config-temp.yaml ./tape/config.yaml
sed -i "s/priv_sk/$priv_sk/" ./tape/config.yaml

#backend download directory
rm -rf ../application/server/files
mkdir -p ../application/server/files/downloadfiles
mkdir -p ../application/server/files/uploadfiles
