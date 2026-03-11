Write-Host "Building Mathy with cgo tag..."
go build -x -v -tags enablecgo

Write-Host "Starting Mathy..."
./mathy.exe