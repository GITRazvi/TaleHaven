# Simple Reading Progress API Test
$baseUrl = "http://localhost:3001"
$sessionId = "test-session-$(Get-Random)"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Reading Progress API Test" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Session ID: $sessionId" -ForegroundColor Yellow
Write-Host ""

$headers = @{
    "Content-Type" = "application/json"
    "Cookie" = "session_id=$sessionId"
}

# Test 1: Save reading progress
Write-Host "[Test 1] Saving reading progress..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-1"
        lastPage = 15
        totalPages = 50
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST -Headers $headers -Body $body
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
    $response | ConvertTo-Json | Write-Host -ForegroundColor Gray
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Get chapter progress
Write-Host "[Test 2] Getting chapter progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-1?mangaId=test-manga-1" -Method GET -Headers $headers
    Write-Host "  SUCCESS: Found progress at page $($response.data.lastPage)" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5 | Write-Host -ForegroundColor Gray
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 3: Save progress for another chapter
Write-Host "[Test 3] Saving progress for another chapter..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-2"
        lastPage = 8
        totalPages = 40
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST -Headers $headers -Body $body
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 4: Get manga progress (all chapters)
Write-Host "[Test 4] Getting all progress for manga..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga/test-manga-1" -Method GET -Headers $headers
    $chapterCount = $response.data.chapters.PSObject.Properties.Count
    Write-Host "  SUCCESS: Found $chapterCount chapters" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5 | Write-Host -ForegroundColor Gray
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 5: Save progress for another manga
Write-Host "[Test 5] Saving progress for a different manga..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-2"
        chapterId = "test-chapter-1"
        lastPage = 25
        totalPages = 60
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST -Headers $headers -Body $body
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Get all reading progress
Write-Host "[Test 6] Getting all reading progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method GET -Headers $headers
    $mangaCount = $response.data.mangas.PSObject.Properties.Count
    Write-Host "  SUCCESS: Found progress for $mangaCount manga(s)" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5 | Write-Host -ForegroundColor Gray
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 7: Get manga list
Write-Host "[Test 7] Getting read manga list..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga-list" -Method GET -Headers $headers
    Write-Host "  SUCCESS: Found $($response.count) manga(s) in list" -ForegroundColor Green
    $response | ConvertTo-Json | Write-Host -ForegroundColor Gray
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 8: Update existing progress
Write-Host "[Test 8] Updating existing progress..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-1"
        lastPage = 30
        totalPages = 50
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST -Headers $headers -Body $body
    Write-Host "  SUCCESS: Progress updated to page 30" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 9: Verify update
Write-Host "[Test 9] Verifying updated progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-1?mangaId=test-manga-1" -Method GET -Headers $headers
    Write-Host "  SUCCESS: Progress is now at page $($response.data.lastPage)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 10: Delete chapter progress
Write-Host "[Test 10] Deleting chapter progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-2?mangaId=test-manga-1" -Method DELETE -Headers $headers
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 11: Delete manga progress
Write-Host "[Test 11] Deleting all progress for manga-2..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga/test-manga-2" -Method DELETE -Headers $headers
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 12: Delete all progress
Write-Host "[Test 12] Deleting all progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method DELETE -Headers $headers
    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Test Completed!" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

