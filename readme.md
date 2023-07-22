# Chainlink feed change SOS

## Build
```bash
go build
```

## Usage
Note: If you don't specify `--output` tag the script will output the TX hashed on screen

### transferPayeeship
```bash
# --private-key asdasd REQUIRED
# --rpc https://chain.infura.com REQUIRED
# --chain-id 5 REQUIRED
# --new-private-key dsadsa REQUIRED
# --file input.json REQUIRED
# --type transfer | accept REQUIRED
# --output output.json

./chainlink-sos --private-key --rpc --chain-id --new-private-key --file input.json --type transfer --output output.json
```

### acceptPayeeship
```bash
# --private-key asdasd REQUIRED
# --rpc https://chain.infura.com REQUIRED
# --chain-id 5 REQUIRED
# --new-private-key dsadsa REQUIRED
# --file input.json REQUIRED
# --type transfer | accept REQUIRED
# --output output.json

./chainlink-sos --private-key --rpc --chain-id --new-private-key --file input.json --type accept --output output.json
```

### input.json
```json
{
    "feeds": [
        {
            "feed": "0x0",
            "transmitter": "0x0"
        }
    ]
}
```