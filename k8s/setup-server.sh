#!/usr/bin/env bash
set -euxo pipefail

curl -X POST ssrg:5060/register -d "instance=$POD_NAME&ip=$POD_IP" > nodeid.txt

echo -e "\n\n# Tendermint Hosts\n\n" >> /etc/hosts
curl ssrg:5060/ips >> /etc/hosts

ID=$(<nodeid.txt)

curl -o $HOME/.tendermint/config/config.toml "ssrg:5060/configfiles/node$ID/config/config.toml"
curl -o $HOME/.tendermint/config/genesis.json "ssrg:5060/configfiles/node$ID/config/genesis.json"
curl -o $HOME/.tendermint/config/node_key.json "ssrg:5060/configfiles/node$ID/config/node_key.json"
curl -o $HOME/.tendermint/config/priv_validator_key.json "ssrg:5060/configfiles/node$ID/config/priv_validator_key.json"
curl -o $HOME/.tendermint/data/priv_validator_state.json "ssrg:5060/configfiles/node$ID/data/priv_validator_state.json"