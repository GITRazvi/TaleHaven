# Reading Progress API Test Script
# This PowerShell script tests all reading progress endpoints

$baseUrl = "http://localhost:3001"
$sessionId = "test-session-$(Get-Random)"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Reading Progress API Test Script" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Session ID: $sessionId" -ForegroundColor Yellow
Write-Host ""

# Test data
$mangaId1 = "a77742b1-befd-49a4-bff5-1ad4e6b0ef7b"
$chapterId1 = "73af4d8d-1532-4a72-b1b9-8f4e5cd295c9"
$mangaId2 = "test-manga-2"
$chapterId2 = "test-chapter-2"

function Invoke-ApiTest {
    param(
        [string]$TestName,
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null
    )

    Write-Host "Testing: $TestName" -ForegroundColor Green
    Write-Host "  Method: $Method" -ForegroundColor Gray
    Write-Host "  Endpoint: $Endpoint" -ForegroundColor Gray

    try {
        $headers = @{
            "Content-Type" = "application/json"
            "Cookie" = "session_id=$sessionId"
        }

        $params = @{
            Uri = "$baseUrl$Endpoint"
            Method = $Method
            Headers = $headers
        }

        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json)
            Write-Host "  Body: $($params.Body)" -ForegroundColor Gray
        }

        $response = Invoke-RestMethod @params
        Write-Host "  ✓ SUCCESS" -ForegroundColor Green
        Write-Host "  Response:" -ForegroundColor Gray
        $response | ConvertTo-Json -Depth 5 | Write-Host -ForegroundColor DarkGray
        Write-Host ""
        return $true
    }
    catch {
        Write-Host "  ✗ FAILED" -ForegroundColor Red
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
        Write-Host ""
        return $false
    }
}

Write-Host "=== Phase 1: Save Reading Progress ===" -ForegroundColor Cyan
Write-Host ""

# Test 1: Save reading progress for manga 1, chapter 1
Invoke-ApiTest `
    -TestName "Save progress - Manga 1, Chapter 1, Page 15" `
    -Method "POST" `
    -Endpoint "/reading-progress" `
    -Body @{
        mangaId = $mangaId1
        chapterId = $chapterId1
        lastPage = 15
        totalPages = 53
    }

# Test 2: Save reading progress for manga 1, chapter 2
Invoke-ApiTest `
    -TestName "Save progress - Manga 1, Chapter 2, Page 8" `
    -Method "POST" `
    -Endpoint "/reading-progress" `
    -Body @{
        mangaId = $mangaId1
        chapterId = $chapterId2
        lastPage = 8
        totalPages = 40
    }

# Test 3: Save reading progress for manga 2, chapter 1
Invoke-ApiTest `
    -TestName "Save progress - Manga 2, Chapter 1, Page 25" `
    -Method "POST" `
    -Endpoint "/reading-progress" `
    -Body @{
        mangaId = $mangaId2
        chapterId = $chapterId1
        lastPage = 25
        totalPages = 60
    }

Write-Host "=== Phase 2: Retrieve Reading Progress ===" -ForegroundColor Cyan
Write-Host ""

# Test 4: Get chapter progress
Invoke-ApiTest `
    -TestName "Get chapter progress - Manga 1, Chapter 1" `
    -Method "GET" `
    -Endpoint "/reading-progress/chapter/${chapterId1}?mangaId=$mangaId1"

# Test 5: Get manga progress
Invoke-ApiTest `
    -TestName "Get all progress for Manga 1" `
    -Method "GET" `
    -Endpoint "/reading-progress/manga/$mangaId1"

# Test 6: Get all reading progress
Invoke-ApiTest `
    -TestName "Get all reading progress for session" `
    -Method "GET" `
    -Endpoint "/reading-progress"

# Test 7: Get read manga list
Invoke-ApiTest `
    -TestName "Get list of read manga IDs" `
    -Method "GET" `
    -Endpoint "/reading-progress/manga-list"

Write-Host "=== Phase 3: Update Progress ===" -ForegroundColor Cyan
Write-Host ""

# Test 8: Update existing progress
Invoke-ApiTest `
    -TestName "Update progress - Manga 1, Chapter 1, Page 30" `
    -Method "POST" `
    -Endpoint "/reading-progress" `
    -Body @{
        mangaId = $mangaId1
        chapterId = $chapterId1
        lastPage = 30
        totalPages = 53
    }

# Test 9: Verify update
Invoke-ApiTest `
    -TestName "Verify updated progress" `
    -Method "GET" `
    -Endpoint "/reading-progress/chapter/${chapterId1}?mangaId=$mangaId1"

Write-Host "=== Phase 4: Delete Progress ===" -ForegroundColor Cyan
Write-Host ""

# Test 10: Delete chapter progress
Invoke-ApiTest `
    -TestName "Delete chapter progress - Manga 1, Chapter 2" `
    -Method "DELETE" `
    -Endpoint "/reading-progress/chapter/${chapterId2}?mangaId=$mangaId1"

# Test 11: Verify chapter deleted
Invoke-ApiTest `
    -TestName "Verify chapter deleted (should return 404)" `
    -Method "GET" `
    -Endpoint "/reading-progress/chapter/${chapterId2}?mangaId=$mangaId1"

# Test 12: Get manga progress after deletion
Invoke-ApiTest `
    -TestName "Get manga progress after deleting one chapter" `
    -Method "GET" `
    -Endpoint "/reading-progress/manga/$mangaId1"

# Test 13: Delete manga progress
Invoke-ApiTest `
    -TestName "Delete all progress for Manga 2" `
    -Method "DELETE" `
    -Endpoint "/reading-progress/manga/$mangaId2"

# Test 14: Get all progress after manga deletion
Invoke-ApiTest `
    -TestName "Get all progress after deleting Manga 2" `
    -Method "GET" `
    -Endpoint "/reading-progress"

# Test 15: Delete all progress
Invoke-ApiTest `
    -TestName "Delete all reading progress" `
    -Method "DELETE" `
    -Endpoint "/reading-progress"

# Test 16: Verify all deleted
Invoke-ApiTest `
    -TestName "Verify all progress deleted (should return empty)" `
    -Method "GET" `
    -Endpoint "/reading-progress"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Test Script Completed!" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: Some tests are expected to fail (like 404 responses) - this is normal!" -ForegroundColor Yellow
Write-Host "The important thing is that the API responds correctly to each request." -ForegroundColor Yellow

