$ErrorActionPreference = "Stop"

$login = Invoke-WebRequest -Method POST -Uri 'http://localhost:3001/api/v1/users/login' -Headers @{'Content-Type'='application/json'} -Body '{"email":"admin_20260619212314@catalyst.dev","password":"Password123!"}' -UseBasicParsing | ConvertFrom-Json
$token = $login.data.access_token

try {
    $res = Invoke-WebRequest -Method POST -Uri 'http://localhost:3001/api/v1/transactions' -Headers @{'Content-Type'='application/json'; 'Authorization'="Bearer $token"; 'Idempotency-Key'='test-key-123'} -Body '{"products":[{"id":54,"qty":1}]}' -UseBasicParsing
    Write-Host $res.Content
} catch {
    $stream = $_.Exception.Response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($stream)
    Write-Host "STATUS CODE:"
    Write-Host $_.Exception.Response.StatusCode
    Write-Host "ERROR RESPONSE:"
    Write-Host $reader.ReadToEnd()
}
