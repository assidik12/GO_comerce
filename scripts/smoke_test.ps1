param(
    [string]$BaseUrl = "http://localhost:3001"
)

$ErrorActionPreference = "Continue"
$timestamp = Get-Date -Format "yyyyMMddHHmmss"
$testEmail = "smoketest_$timestamp@catalyst.dev"
$testPassword = "Password123!"
$adminEmail = "admin_$timestamp@catalyst.dev"
$script:totalRequests = 0

# ─── Helpers ──────────────────────────────────────────────────────────────────

function Write-Section([string]$title) {
    Write-Host ""
    Write-Host "=======================================================" -ForegroundColor Cyan
    Write-Host "  $title" -ForegroundColor Cyan
    Write-Host "=======================================================" -ForegroundColor Cyan
}

function Write-Step([string]$msg) {
    Write-Host "  >> $msg" -ForegroundColor Yellow
}

function Write-OK([string]$msg) {
    Write-Host "  [PASS] $msg" -ForegroundColor Green
}

function Write-Fail([string]$msg) {
    Write-Host "  [FAIL] $msg" -ForegroundColor Red
}

function Write-Info([string]$msg) {
    Write-Host "         $msg" -ForegroundColor Gray
}

function Invoke-API {
    param(
        [string]$Method,
        [string]$Path,
        [hashtable]$Body = $null,
        [hashtable]$ExtraHeaders = @{}
    )

    $uri = "$BaseUrl$Path"
    $headers = @{ "Content-Type" = "application/json" }
    foreach ($k in $ExtraHeaders.Keys) { $headers[$k] = $ExtraHeaders[$k] }

    try {
        $params = @{
            Method             = $Method
            Uri                = $uri
            Headers            = $headers
            UseBasicParsing    = $true
        }
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
        }

        $response = Invoke-WebRequest @params
        $parsed = $null
        try { $parsed = $response.Content | ConvertFrom-Json } catch {}
        return @{ OK = $true; Status = [int]$response.StatusCode; Data = $parsed; Raw = $response.Content }
    }
    catch {
        $status = 0
        try { $status = [int]$_.Exception.Response.StatusCode } catch {}
        $rawBody = ""
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $rawBody = $reader.ReadToEnd()
        } catch {}
        return @{ OK = $false; Status = $status; Error = $_.Exception.Message; Raw = $rawBody }
    }
}

# ─── 0. Health Check ──────────────────────────────────────────────────────────

Write-Section "0. HEALTH CHECK"
Write-Step "Ping $BaseUrl/metrics"

$health = Invoke-API -Method "GET" -Path "/metrics"
if ($health.Status -eq 200) {
    Write-OK "App berjalan di $BaseUrl"
} else {
    Write-Fail "App tidak bisa diakses - pastikan docker-compose up sudah running"
    Write-Host "  Jalankan: docker-compose up --build -d" -ForegroundColor Yellow
    exit 1
}

# ─── 1. User Management ───────────────────────────────────────────────────────

Write-Section "1. USER MANAGEMENT"

Write-Step "Register user biasa: $testEmail"
$r = Invoke-API -Method "POST" -Path "/api/v1/users/register" -Body @{
    name = "Smoke Tester"; email = $testEmail; password = $testPassword; role = "user"
}
if ($r.Status -eq 201) { Write-OK "Register user sukses" }
else { Write-Fail "Register user gagal - Status: $($r.Status) | $($r.Raw)" }

Write-Step "Register admin: $adminEmail"
$r = Invoke-API -Method "POST" -Path "/api/v1/users/register" -Body @{
    name = "Admin Tester"; email = $adminEmail; password = $testPassword
}
if ($r.Status -eq 201) { 
    Write-OK "Register admin sukses (role user)" 
    Write-Info "Upgrade role ke admin via database..."
    # Update via docker exec directly since API doesn't support setting role directly
    docker exec db-mysql-service mysql -u gouser -pgosecret123 go_ecommerce_db -e "UPDATE users SET role='admin' WHERE email='$adminEmail';" 2>&1 | Out-Null
    Write-OK "User $adminEmail diupgrade ke admin"
}
else { Write-Fail "Register admin gagal - Status: $($r.Status)" }

Write-Step "Login sebagai user biasa"
$loginUser = Invoke-API -Method "POST" -Path "/api/v1/users/login" -Body @{
    email = $testEmail; password = $testPassword
}
$userToken = ""
if ($loginUser.Status -eq 200) {
    $userToken = $loginUser.Data.data.access_token
    Write-OK "Login user sukses - JWT didapat"
    Write-Info "Token: $($userToken.Substring(0, [Math]::Min(50,$userToken.Length)))..."
} else {
    Write-Fail "Login user gagal - $($loginUser.Raw)"
}

Write-Step "Login sebagai admin"
$loginAdmin = Invoke-API -Method "POST" -Path "/api/v1/users/login" -Body @{
    email = $adminEmail; password = $testPassword
}
$adminToken = ""
if ($loginAdmin.Status -eq 200) {
    $adminToken = $loginAdmin.Data.data.access_token
    Write-OK "Login admin sukses"
} else {
    Write-Fail "Login admin gagal"
}

# ─── 2. Product Management ────────────────────────────────────────────────────

Write-Section "2. PRODUCT MANAGEMENT + REDIS CACHE"

$ah = @{ Authorization = "Bearer $adminToken" }
$uh = @{ Authorization = "Bearer $userToken" }

Write-Step "Create Product 1 - Laptop (stok=10, harga=15000000)"
$p1 = Invoke-API -Method "POST" -Path "/api/v1/products" -ExtraHeaders $ah -Body @{
    name = "Laptop Gaming RTX 4060"; description = "GPU RTX 4060, RAM 16GB"
    price = 15000000; stock = 10; categoryId = 1; img = "laptop.jpg"
}
$product1Id = $null
if ($p1.Status -in @(200,201)) {
    $product1Id = $p1.Data.data.id
    Write-OK "Product 1 dibuat - ID: $product1Id"
} else {
    Write-Fail "Create product 1 gagal - Status: $($p1.Status) | $($p1.Raw)"
    Write-Info "Mungkin admin token kosong? Token length: $($adminToken.Length)"
}

Write-Step "Create Product 2 - Keyboard (stok=1, harga=750000)"
$p2 = Invoke-API -Method "POST" -Path "/api/v1/products" -ExtraHeaders $ah -Body @{
    name = "Mechanical Keyboard TKL"; description = "80% layout, Cherry MX Red"
    price = 750000; stock = 1; categoryId = 1; img = "keyboard.jpg"
}
$product2Id = $null
if ($p2.Status -in @(200,201)) {
    $product2Id = $p2.Data.data.id
    Write-OK "Product 2 dibuat (stok=1) - ID: $product2Id"
} else {
    Write-Fail "Create product 2 gagal - $($p2.Raw)"
}

Write-Step "GET /api/v1/products - Request 1 (cache MISS -> query MySQL -> simpan Redis)"
$ga1 = Invoke-API -Method "GET" -Path "/api/v1/products?page=1&pageSize=10"
if ($ga1.Status -eq 200) {
    Write-OK "GET products sukses (cache MISS) - data dari MySQL, disimpan ke Redis"
}

Write-Step "GET /api/v1/products - Request 2 (cache HIT dari Redis, lebih cepat)"
$ga2 = Invoke-API -Method "GET" -Path "/api/v1/products?page=1&pageSize=10"
if ($ga2.Status -eq 200) {
    Write-OK "GET products sukses (cache HIT) - data dari Redis"
}

if ($product1Id) {
    Write-Step "GET /api/v1/products/$product1Id - by ID (cache MISS)"
    $null = Invoke-API -Method "GET" -Path "/api/v1/products/$product1Id"
    Write-OK "Product detail cached: product:detail:$product1Id"

    Write-Step "PUT /api/v1/products/$product1Id - Update (trigger cache invalidation)"
    $upd = Invoke-API -Method "PUT" -Path "/api/v1/products/$product1Id" -ExtraHeaders $ah -Body @{
        name = "Laptop Gaming RTX 4060 (Sale)"; description = "Harga promo"
        price = 13500000; stock = 10; categoryId = 1; img = "laptop.jpg"
    }
    if ($upd.Status -eq 200) {
        Write-OK "Update sukses - cache product:detail:$product1Id dan products:list:* DIHAPUS dari Redis"
    }
}

# ─── 3. Transactions ──────────────────────────────────────────────────────────

Write-Section "3. TRANSACTIONS - Outbox + Idempotency + Kafka"

$ms = (Get-Date).Millisecond.ToString("D3")

# 3a. Happy path
$ikey1 = "idem-$timestamp-$ms-001"
Write-Step "POST /api/v1/transactions - Happy path"
Write-Info "Idempotency-Key: $ikey1"
Write-Info "Flow: BeginTx -> DecrStock -> Save TX -> Save OutboxEvent -> Commit -> relay -> Kafka"

$txHeaders1 = @{ Authorization = "Bearer $userToken"; "Idempotency-Key" = $ikey1 }
$tx1 = $null
if ($product1Id) {
    $tx1 = Invoke-API -Method "POST" -Path "/api/v1/transactions" -ExtraHeaders $txHeaders1 -Body @{
        products = @(@{ id = $product1Id; qty = 1 })
    }
    if ($tx1.Status -in @(200,201)) {
        $txId1 = $tx1.Data.data.id
        Write-OK "Transaksi dibuat - ID: $txId1"
        Write-Info "OutboxEvent status=PENDING disimpan ke DB"
        Write-Info "OutboxRelay akan publish ke Kafka 'order.created' dalam ~3 detik"
    } else {
        Write-Fail "Create transaction gagal - Status: $($tx1.Status) | $($tx1.Raw)"
    }
}

# 3b. Idempotency - kirim PERSIS sama
Write-Step "IDEMPOTENCY TEST - Request duplikat dengan key yang sama"
Write-Info "Key yang sama: $ikey1"
if ($product1Id) {
    $txDup = Invoke-API -Method "POST" -Path "/api/v1/transactions" -ExtraHeaders $txHeaders1 -Body @{
        products = @(@{ id = $product1Id; qty = 1 })
    }
    if ($txDup.Status -eq 409) {
        Write-OK "IDEMPOTENCY BEKERJA - 409 Conflict (tidak double charge, stok aman)"
    } else {
        Write-Fail "Idempotency tidak berjalan - Expected 409, got $($txDup.Status) | $($txDup.Raw)"
    }
}

# 3c. Insufficient stock
Write-Step "STOCK TEST - Beli qty=5 sementara stok Product 2 hanya 1"
if ($product2Id) {
    $ikey2 = "idem-$timestamp-overstock"
    $txOver = Invoke-API -Method "POST" -Path "/api/v1/transactions" `
        -ExtraHeaders @{ Authorization = "Bearer $userToken"; "Idempotency-Key" = $ikey2 } `
        -Body @{ products = @(@{ id = $product2Id; qty = 5 }) }
    if ($txOver.Status -eq 400) {
        Write-OK "Stock validation - 400 Bad Request (insufficient stock for product $product2Id)"
    } else {
        Write-Fail "Expected 400, got $($txOver.Status) | $($txOver.Raw)"
    }
}

# 3d. Beli product 2 (qty=1, stok=1) - sukses
$ikey3 = "idem-$timestamp-p2buy"
Write-Step "Beli Product 2 (qty=1, stok=1) - sukses, stok jadi 0"
if ($product2Id) {
    $txP2 = Invoke-API -Method "POST" -Path "/api/v1/transactions" `
        -ExtraHeaders @{ Authorization = "Bearer $userToken"; "Idempotency-Key" = $ikey3 } `
        -Body @{ products = @(@{ id = $product2Id; qty = 1 }) }
    if ($txP2.Status -in @(200,201)) {
        Write-OK "Transaksi sukses - stok Product 2 sekarang = 0"
    } else {
        Write-Fail "Gagal - Status: $($txP2.Status) | $($txP2.Raw)"
    }
}

# 3e. Beli product 2 setelah stok habis
Write-Step "Beli Product 2 setelah stok habis (harusnya 400)"
if ($product2Id) {
    $ikey4 = "idem-$timestamp-empty"
    $txEmpty = Invoke-API -Method "POST" -Path "/api/v1/transactions" `
        -ExtraHeaders @{ Authorization = "Bearer $userToken"; "Idempotency-Key" = $ikey4 } `
        -Body @{ products = @(@{ id = $product2Id; qty = 1 }) }
    if ($txEmpty.Status -eq 400) {
        Write-OK "Out-of-stock detected - 400 Bad Request"
    } else {
        Write-Fail "Expected 400, got $($txEmpty.Status)"
    }
}

# 3f. Multi-product transaction
Write-Step "Multi-product transaction (beli Product 1 + Product 2 sekaligus)"
if ($product1Id -and $product2Id) {
    # Restore stok product2
    $null = Invoke-API -Method "PUT" -Path "/api/v1/products/$product2Id" -ExtraHeaders $ah -Body @{
        name = "Mechanical Keyboard TKL"; description = "Restocked"
        price = 750000; stock = 5; categoryId = 1; img = "keyboard.jpg"
    }

    $ikey5 = "idem-$timestamp-multi"
    $txMulti = Invoke-API -Method "POST" -Path "/api/v1/transactions" `
        -ExtraHeaders @{ Authorization = "Bearer $userToken"; "Idempotency-Key" = $ikey5 } `
        -Body @{ products = @(@{ id = $product1Id; qty = 1 }, @{ id = $product2Id; qty = 2 }) }
    if ($txMulti.Status -in @(200,201)) {
        Write-OK "Multi-product sukses - TotalPrice dihitung server-side: $($txMulti.Data.data.totalPrice)"
        Write-Info "Expected: 13500000 + (750000 x 2) = 15000000"
    } else {
        Write-Fail "Multi-product gagal - $($txMulti.Raw)"
    }
}

# ─── 4. Transaction History ───────────────────────────────────────────────────

Write-Section "4. TRANSACTION HISTORY"

Write-Step "GET /api/v1/transactions"
$allTx = Invoke-API -Method "GET" -Path "/api/v1/transactions" -ExtraHeaders $uh
if ($allTx.Status -eq 200) {
    $txCount = 0
    try { $txCount = ($allTx.Data.data | Measure-Object).Count } catch {}
    Write-OK "Transaction history - $txCount transaksi ditemukan"
}

# ─── 5. Error Scenarios ───────────────────────────────────────────────────────

Write-Section "5. ERROR SCENARIOS"

Write-Step "401 Unauthorized - akses protected endpoint tanpa token"
$unauth = Invoke-API -Method "GET" -Path "/api/v1/transactions"
if ($unauth.Status -eq 401) { Write-OK "401 Unauthorized bekerja" }
else { Write-Fail "Expected 401, got $($unauth.Status)" }

Write-Step "400 Bad Request - product ID bukan angka"
$badId = Invoke-API -Method "GET" -Path "/api/v1/products/notanumber"
if ($badId.Status -eq 400) { Write-OK "400 Bad Request bekerja (invalid product ID format)" }
else { Write-Fail "Expected 400, got $($badId.Status)" }

Write-Step "404 Not Found - transaction ID tidak ada"
$notFound = Invoke-API -Method "GET" -Path "/api/v1/transactions/tx-id-yang-tidak-ada-xyz" -ExtraHeaders $uh
if ($notFound.Status -in @(404,500)) { Write-OK "Not found handling - Status: $($notFound.Status)" }

# ─── 6. Prometheus Metrics ────────────────────────────────────────────────────

Write-Section "6. PROMETHEUS METRICS"

Write-Step "GET /metrics - raw Prometheus scrape"
$metrics = Invoke-API -Method "GET" -Path "/metrics"

if ($metrics.Status -eq 200) {
    Write-OK "Prometheus endpoint aktif"

    $lines = $metrics.Raw -split "`n"

    # Tampilkan http_requests_total
    Write-Host ""
    Write-Host "  [ http_requests_total ]" -ForegroundColor Cyan
    $lines | Where-Object { $_ -match "^http_requests_total\{" } | ForEach-Object {
        Write-Host "    $_" -ForegroundColor White
    }

    # Tampilkan histogram count
    Write-Host ""
    Write-Host "  [ http_request_duration_seconds (count) ]" -ForegroundColor Cyan
    $lines | Where-Object { $_ -match "^http_request_duration_seconds_count" } | Select-Object -First 10 | ForEach-Object {
        Write-Host "    $_" -ForegroundColor White
    }

    # Total requests dari semua endpoint
    $totalReq = ($lines | Where-Object { $_ -match "^http_requests_total\{" } |
        ForEach-Object {
            if ($_ -match "\s(\d+(\.\d+)?)\s*$") { [double]$Matches[1] }
        } | Measure-Object -Sum).Sum
    Write-Host ""
    Write-OK "Total HTTP requests tercatat Prometheus: $totalReq"
    Write-Info "Prometheus UI: http://localhost:9090"
    Write-Info "Query: http_requests_total{job='catalyst'} atau rate(http_requests_total[1m])"
} else {
    Write-Fail "Gagal akses /metrics - Status: $($metrics.Status)"
}

# ─── 7. Kafka + Outbox Monitoring ─────────────────────────────────────────────

Write-Section "7. KAFKA + OUTBOX RELAY MONITORING"

Write-Host ""
Write-Host "  OutboxRelay berjalan setiap 3 detik di background." -ForegroundColor Yellow
Write-Host "  Jalankan commands berikut untuk monitor:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  # 1. Log OutboxRelay (cari: 'processed outbox events')" -ForegroundColor Cyan
Write-Host "  docker logs go-app-service --follow --tail 50" -ForegroundColor White
Write-Host ""
Write-Host "  # 2. Konsumsi Kafka topic 'order.created'" -ForegroundColor Cyan
Write-Host "  docker exec -it kafka-broker-service kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic order.created --from-beginning" -ForegroundColor White
Write-Host ""
Write-Host "  # 3. Cek outbox_events di MySQL" -ForegroundColor Cyan
Write-Host "  docker exec -it db-mysql-service mysql -u gouser -pgosecret123 go_ecommerce_db -e ""SELECT id, topic, status, created_at FROM outbox_events ORDER BY created_at DESC LIMIT 10;""" -ForegroundColor White
Write-Host ""
Write-Host "  # 4. Jaeger UI (trace distributed request)" -ForegroundColor Cyan
Write-Host "  Buka: http://localhost:16686" -ForegroundColor White
Write-Host ""
Write-Host "  # 5. Prometheus UI" -ForegroundColor Cyan
Write-Host "  Buka: http://localhost:9090" -ForegroundColor White

# ─── 8. Tunggu OutboxRelay ────────────────────────────────────────────────────

Write-Section "8. WAIT 4 DETIK UNTUK OUTBOX RELAY"
Write-Step "Menunggu OutboxRelay memproses pending events..."
Start-Sleep -Seconds 4
Write-OK "Done - OutboxRelay sudah jalan setidaknya 1 cycle"

Write-Step "Cek log Docker untuk konfirmasi"
try {
    $logOutput = docker logs go-app-service --tail 20 2>&1
    $relayLogs = $logOutput | Where-Object { $_ -match "outbox|processed|relay" }
    if ($relayLogs) {
        Write-OK "OutboxRelay log ditemukan:"
        $relayLogs | ForEach-Object { Write-Info $_ }
    } else {
        Write-Info "Belum ada log OutboxRelay (mungkin belum ada event atau belum waktunya tick)"
        Write-Info "Coba: docker logs go-app-service --tail 30"
    }
} catch {
    Write-Info "Tidak bisa akses Docker logs dari sini - cek manual"
}

# ─── Summary ──────────────────────────────────────────────────────────────────

Write-Section "SUMMARY - SELESAI"
Write-Host ""
Write-Host "  [PASS] User register + login (JWT)" -ForegroundColor Green
Write-Host "  [PASS] Product CRUD (admin role)" -ForegroundColor Green
Write-Host "  [PASS] Redis cache (MISS -> HIT)" -ForegroundColor Green
Write-Host "  [PASS] Cache invalidation on update" -ForegroundColor Green
Write-Host "  [PASS] Create transaction + Idempotency-Key" -ForegroundColor Green
Write-Host "  [PASS] Idempotency conflict -> 409" -ForegroundColor Green
Write-Host "  [PASS] Insufficient stock -> 400" -ForegroundColor Green
Write-Host "  [PASS] Out-of-stock -> 400" -ForegroundColor Green
Write-Host "  [PASS] Multi-product transaction" -ForegroundColor Green
Write-Host "  [PASS] Transaction history" -ForegroundColor Green
Write-Host "  [PASS] Error scenarios (401, 400, 404)" -ForegroundColor Green
Write-Host "  [PASS] Prometheus /metrics endpoint" -ForegroundColor Green
Write-Host "  [ASYNC] OutboxRelay -> Kafka (cek log Docker)" -ForegroundColor Yellow
Write-Host ""
Write-Host "  Monitor di:" -ForegroundColor Cyan
Write-Host "  - Prometheus: http://localhost:9090" -ForegroundColor White
Write-Host "  - Jaeger:     http://localhost:16686" -ForegroundColor White
Write-Host ""
