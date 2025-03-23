# Setup

Here's a walkthrough on how to setup the Hyperledger Fabric network locally.


# Required Tools

**Windows:** You'll need to perform all steps within your Linux subsystem (WSL)

[Docker Desktop](https://www.docker.com/products/docker-desktop/) to run the network locally in a container

***NOTE:** If you're using Windows, make sure to enable WSL integration in Docker Desktop*

[just](https://github.com/casey/just#installation) to run all the commands here directly

[nvm](https://github.com/nvm-sh/nvm#installing-and-updating) to install node and npm
```shell
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash
```
[node LTS and npm](https://github.com/nvm-sh/nvm#usage) to run node chaincode and applications
```shell
nvm install --lts
nvm use --lts
```
[Golang](https://go.dev/doc/install) to compile chaincode and run backend

[weft ](https://www.npmjs.com/package/@hyperledger-labs/weft) Hyperledger-Labs cli to work with identities and chaincode packages
```shell
npm install -g @hyperledger-labs/weft
```

[jq](https://stedolan.github.io/jq/) JSON command-line processor
```shell
sudo apt-get update && sudo apt-get install -y jq
```

Fabric peer CLI
```shell
curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh | bash -s -- binary
export WORKSHOP_PATH=$(pwd)
export PATH=${WORKSHOP_PATH}/bin:$PATH
export FABRIC_CFG_PATH=${WORKSHOP_PATH}/config
```
***NOTE:*** 
- *Sometimes you'll get an error saying that the environment variable FABRIC_CFG_PATH is not set or set to a non-existent file path*
- *To fix this, just reset the environment variable to where your "config" directory lies at.*

***Example:***
 ```shell 
 export FABRIC_CFG_PATH=/home/brian_nguyen/config
```

> to check if you have all the necessary tools and configurations, run `./check.sh` from the repo



# Start the Fabric Infrastructure
We're using MicroFab for the Fabric infrastructure as it's a single container that is fast to start.
The MicroFab container includes an ordering service node and a peer process that is pre-configured to create a channel and call external chaincodes.
It also includes credentials for an `org1` organization, which will be used to run the peer. We'll use an `org1` admin user when interacting with the environment.

We'll use `just` recipes to execute multiple commands. `just` recipes are similar to `make` but simpler to understand. You can open the [justfile](../../justfile) in the project root directory to see which commands are run with each recipe.

Start the MicroFab container by running the `just` recipe. This will set some properties for the MicroFab environment and start the MicroFab docker container.

```bash
just microfab
```

This will start the docker container (automatically download it if necessary), and also write out some configuration/data files in the `_cfg/uf` directory.

```bash
ls -1 _cfg/uf

_cfg
_gateways
_wallets
org1admin.env
org2admin.env
```

A file `org1admin.env` is written out that contains the environment variables needed to run applications _as the org1 admin identity_.

# Installing Chaincode

After having the microfab container running locally, we can begin installing the chaincode for onto the network.

Start  by navigating to the chaincode
```bash
cd ./contracts/real-estate-ledger-go/realestatesec_chaincode
```

Then we'll need to install the smart contract dependencies
```shell
GO111MODULE=on go mod vendor
```
*If the command is successful, the go packages will be installed inside a `vendor` folder.*

Now navigate back to the project's root directory and create the chaincode package

```shell
peer lifecycle chaincode package realestatesec.tar.gz --path <chaincode_proj_dir_path> --lang golang --label realestatesec_1.0
```

Now that we've packaged the chaincode, lets install it onto the network.

Let's first install the chaincode onto the channel's peers

Set environment variables to given peer's admin and target the given peer we want to install chaincode onto

```shell
source _cfg/uf/org1admin.env
```

Install onto peer

```shell
peer lifecycle chaincode install realestatesec.tar.gz
```

Save chaincode package ID as environment variable, this is used to approve the chaincode at the organization level

```shell
export CC_PACKAGE_ID=basic_1.0:69de748301770f6ef64b42aa6bb6cb291df20aa39542c3ef94008615704007f3
```
***NOTE:** Replace example with your package's ID*

***Repeat for a peer for each org within the channel (on microfab there is only one org and one peer, so just do it once)***

Next let's approve the chaincode definition at the org level

Make sure that the chaincode is install on the peer
```shell
peer lifecycle chaincode queryinstalled
```

Approve the chaincode

```shell
peer lifecycle chaincode approveformyorg -o $ORDERER_ENDPOINT --channelID mychannel --name realestatesec --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1
```
***Repeat for each org within the channel (on microfab there is only one org and one peer, so just do it once)***

Next, commit the chaincode to the channel

First check the commit readiness of the chaincode

```shell
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name realestatesec --version 1.0  --sequence 1 –output json
```
***Needs to be approved by all orgs in the channel***

Then commit to the channel
```shell
peer lifecycle chaincode commit -o $ORDERER_ENDPOINT --channelID mychannel --name realestatesec --version 1.0  --sequence 1
```

Finally, check if the chaincode is committed successfully
```shell
peer lifecycle chaincode querycommitted --channelID mychannel --name realestatesec
```

# Test Chaincode

We can test the chaincode by running queries using the peer CLI tool. 

Here are some examples for the current version of the chaincode

Check if property exists

    peer chaincode query -C mychannel -o $ORDERER_ENDPOINT -n realestatesec -c '{"Args":["PropertyExists", "1"]}' --connTimeout 15s

  

Register Property

    peer chaincode invoke -C mychannel -o $ORDERER_ENDPOINT -n realestatesec -c '{"Args":["RegisterProperty", "1", "123 Jane Street", "Brian Nguyen", "Malik"]}' --connTimeout 15s

  

View All Properties

    peer chaincode query -C mychannel -o $ORDERER_ENDPOINT -n realestatesec -c '{"Args":["ViewProperties"]}' --connTimeout 15s

List Property

    peer chaincode invoke -C mychannel -o $ORDERER_ENDPOINT -n realestatesec -c '{"Args":["ListProperty", "1"]}' --connTimeout 15s

Place Bid

    peer chaincode invoke -C mychannel -o $ORDERER_ENDPOINT -n realestatesec -c '{"Args":["PlaceBid", “1”, “1”, 100000, “John Doe”, “Malik”]}' --connTimeout 15s
