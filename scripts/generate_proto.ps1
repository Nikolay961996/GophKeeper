# generate_proto.ps1 - Генерация Go кода из .proto файлов

Write-Host "Generating Go code from protobuf..." -ForegroundColor Green

# Проверяем наличие protoc
$protocPath = Get-Command protoc -ErrorAction SilentlyContinue
if (-not $protocPath) {
    Write-Host "Error: protoc not found. Please install protobuf compiler." -ForegroundColor Red
    Write-Host "Download from: https://github.com/protocolbuffers/protobuf/releases" -ForegroundColor Yellow
    exit 1
}

# Создаем директорию для сгенерированного кода
$outputDir = "pkg/api"
if (Test-Path $outputDir) {
    Remove-Item -Recurse -Force $outputDir
}
New-Item -ItemType Directory -Path $outputDir -Force | Out-Null

# Генерируем код
protoc --go_out=$outputDir --go_opt=paths=source_relative `
       --go-grpc_out=$outputDir --go-grpc_opt=paths=source_relative `
       --proto_path=api `
       api/gophkeeper.proto

if ($LASTEXITCODE -eq 0) {
    Write-Host "Protobuf code generated successfully in $outputDir" -ForegroundColor Green
} else {
    Write-Host "Failed to generate protobuf code" -ForegroundColor Red
}