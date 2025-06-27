#!/bin/bash
# Usage: ./build.sh <TpeBiConsumer|FtyBiProducer> <dev|prod>
set -e
if [ $# -ne 2 ]; then
  echo "Usage: $0 <TpeBiConsumer|FtyBiProducer> <dev|prod>" >&2
  exit 1
fi
PROJECT="$1"
ENV="$2"
if [[ "$ENV" != "dev" && "$ENV" != "prod" ]]; then
  echo "Environment must be dev or prod" >&2
  exit 1
fi
PLAINTEXT="$PROJECT/config/config.${ENV}.plaintext.yaml"
ENCRYPTED="$PROJECT/config/config.${ENV}.yaml"
CONFIG="$PROJECT/config/config.yaml"

if [ ! -f "$PLAINTEXT" ]; then
  echo "Plaintext config not found: $PLAINTEXT" >&2
  exit 1
fi

# Generate encrypted config
GOFILE="$(dirname "$0")/tools/encryptdb/main.go"
go run "$GOFILE" "$PLAINTEXT" "$ENCRYPTED"
cp "$ENCRYPTED" "$CONFIG"

# Build binary
(cd "$PROJECT" && go build -o "../${PROJECT}_${ENV}")
