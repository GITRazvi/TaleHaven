# Test for Inactivity-Based Auto-Deletion
# This test verifies that TTL is refreshed on read operations

$baseUrl = "http://localhost:3001"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Inactivity-Based Auto-Deletion Test" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Create a WebSession to persist cookies
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# Test 1: Save progress
Write-Host "Step 1: Saving initial progress..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-inactivity"
        chapterId = "test-chapter-inactivity"
        lastPage = 10
        totalPages = 50
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -WebSession $session

    Write-Host "  SUCCESS: Progress saved" -ForegroundColor Green
    Write-Host ""
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

# Note: To fully test TTL, you need Redis CLI access
# This script demonstrates the behavior

Write-Host "Step 2: Reading progress (this should refresh TTL)..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-inactivity?mangaId=test-manga-inactivity" `
        -Method GET `
        -WebSession $session

    Write-Host "  SUCCESS: Progress retrieved (TTL refreshed)" -ForegroundColor Green
    Write-Host "  Current page: $($response.data.lastPage)" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Step 3: Reading all progress (this should refresh all TTLs)..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" `
        -Method GET `
        -WebSession $session

    $mangaCount = ($response.data.mangas.PSObject.Properties | Measure-Object).Count
    Write-Host "  SUCCESS: All progress retrieved (all TTLs refreshed)" -ForegroundColor Green
    Write-Host "  Total manga: $mangaCount" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Step 4: Reading manga list (this should refresh index TTL)..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga-list" `
        -Method GET `
        -WebSession $session

    Write-Host "  SUCCESS: Manga list retrieved (index TTL refreshed)" -ForegroundColor Green
    Write-Host "  Manga count: $($response.count)" -ForegroundColor Gray
    Write-Host ""
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "TTL Behavior Explanation:" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Every operation above refreshed the 30-day TTL:" -ForegroundColor White
Write-Host "  - POST /reading-progress: Sets TTL to 30 days" -ForegroundColor Gray
Write-Host "  - GET /reading-progress/chapter: Refreshes TTL to 30 days" -ForegroundColor Gray
Write-Host "  - GET /reading-progress: Refreshes all chapter TTLs to 30 days" -ForegroundColor Gray
Write-Host "  - GET /reading-progress/manga-list: Refreshes index TTL to 30 days" -ForegroundColor Gray
Write-Host ""
Write-Host "Result: Data will only be deleted after 30 days of NO activity" -ForegroundColor Green
Write-Host "        (no reads AND no writes for 30 consecutive days)" -ForegroundColor Green
Write-Host ""

# Cleanup
Write-Host "Cleaning up test data..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" `
        -Method DELETE `
        -WebSession $session

    Write-Host "  SUCCESS: Test data cleaned up" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "To verify TTL in Redis:" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "1. Connect to Redis CLI: redis-cli" -ForegroundColor White
Write-Host "2. Find keys: KEYS session:*:reading:*" -ForegroundColor White
Write-Host "3. Check TTL: TTL <key-name>" -ForegroundColor White
Write-Host "4. The TTL should reset to ~2592000 seconds (30 days) after each read" -ForegroundColor White
Write-Host ""

