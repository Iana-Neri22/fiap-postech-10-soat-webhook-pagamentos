$ErrorActionPreference = "Stop"
$LambdaName = "WebhookFiap"

Write-Host "🧹 Cleaning previous build..."
if (Test-Path bootstrap) {
    Remove-Item bootstrap -Force -ErrorAction SilentlyContinue
}
if (Test-Path function.zip) {
    Remove-Item function.zip -Force -ErrorAction SilentlyContinue
}

Write-Host "⚙️ Building Go binary..."
Push-Location ./webhook-pagamentos
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o ../bootstrap
Pop-Location

if (-Not (Test-Path "bootstrap")) {
    Write-Error "❌ Build failed. O binário 'bootstrap' não foi gerado."
    exit 1
}

Write-Host "📦 Creating ZIP package..."
Compress-Archive -Path ./bootstrap -DestinationPath function.zip

Write-Host "☁️ Deploying to AWS Lambda..."
aws lambda update-function-code `
  --function-name $LambdaName `
  --zip-file fileb://function.zip

Write-Host "✅ Deployment complete."