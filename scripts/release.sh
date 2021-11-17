VERSION=`git describe --tags`
PREFIX=hetzner-rescaler_${VERSION}_

echo "Building assets for the release ${VERSION}..."

# Cleanup relaese dir
rm -rf ./release

# Create temp dirs
mkdir ./release/ \
./release/${PREFIX}windows-amd64/ \
./release/${PREFIX}linux-amd64/ \
./release/${PREFIX}linux-arm/ \
./release/${PREFIX}linux-arm64/

# Copy assets
cp -R ./LICENCE ./README.md ./CHANGELOG.md ./release/${PREFIX}windows-amd64/ \
cp -R ./LICENCE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-amd64/ \
cp -R ./LICENCE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-arm/ \
cp -R ./LICENCE ./README.md ./CHANGELOG.md ./release/${PREFIX}linux-arm64/

# Build for each platform
GOOS=windows GOARCH=amd64 go build -o ./release/${PREFIX}windows-amd64/hetzner-rescaler.exe . &
GOOS=linux GOARCH=amd64 go build -o ./release/${PREFIX}linux-amd64/hetzner-rescaler . &
GOOS=linux GOARCH=arm64 go build -o ./release/${PREFIX}linux-arm64/hetzner-rescaler . &
GOOS=linux GOARCH=arm go build -o ./release/${PREFIX}linux-arm/hetzner-rescaler . &
wait

# Zip folders
cd ./release/
zip -r ./${PREFIX}windows-amd64.zip ./${PREFIX}windows-amd64/ &
zip -r ./${PREFIX}linux-amd64.zip ./${PREFIX}linux-amd64/ &
zip -r ./${PREFIX}linux-arm64.zip ./${PREFIX}linux-arm64/ &
zip -r ./${PREFIX}linux-arm.zip ./${PREFIX}linux-arm/ &
wait

# Destroy temp dirs
rm -rf ./${PREFIX}windows-amd64 &
rm -rf ./${PREFIX}linux-amd64 &
rm -rf ./${PREFIX}linux-arm64 &
rm -rf ./${PREFIX}linux-arm &
wait