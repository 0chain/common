
#!/usr/bin/env bash

BASEDIR=$(pwd)

echo "Making mocks..."

 generate_mock() {
	OUTPUT=$1
	mockery --case underscore --output=$OUTPUT --all
}

cd $BASEDIR/core || exit
rm -rf ../core/mocks
generate_mock "../core/mocks"

cd $BASEDIR || exit
go generate -run="mockery" ./...

echo "Mocks files are generated."