# Reading Progress API Test with proper session handling
$baseUrl = "http://localhost:3001"

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Reading Progress API Test" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Create a WebSession to persist cookies
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# Test 1
Write-Host "Test 1: Saving reading progress (creates session)..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-1"
        lastPage = 15
        totalPages = 50
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green

    $cookie = $session.Cookies.GetCookies($baseUrl) | Where-Object {$_.Name -eq "session_id"}
    if ($cookie) {
        Write-Host "  Session ID: $($cookie.Value.Substring(0, 20))..." -ForegroundColor Gray
    }
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2
Write-Host "Test 2: Getting chapter progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-1?mangaId=test-manga-1" `
        -Method GET `
        -WebSession $session

    Write-Host "  SUCCESS: Found progress at page $($response.data.lastPage)/$($response.data.totalPages)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 3
Write-Host "Test 3: Saving progress for chapter 2..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-2"
        lastPage = 8
        totalPages = 40
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 4
Write-Host "Test 4: Getting all progress for manga-1..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga/test-manga-1" `
        -Method GET `
        -WebSession $session

    $chapterCount = ($response.data.chapters.PSObject.Properties | Measure-Object).Count
    Write-Host "  SUCCESS: Found $chapterCount chapter(s)" -ForegroundColor Green

    foreach ($chapterProp in $response.data.chapters.PSObject.Properties) {
        $chapter = $chapterProp.Value
        Write-Host "    - Chapter $($chapter.chapterId): Page $($chapter.lastPage)/$($chapter.totalPages)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 5
Write-Host "Test 5: Saving progress for manga-2..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-2"
        chapterId = "test-chapter-1"
        lastPage = 25
        totalPages = 60
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6
Write-Host "Test 6: Getting all reading progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" `
        -Method GET `
        -WebSession $session

    $mangaCount = ($response.data.mangas.PSObject.Properties | Measure-Object).Count
    Write-Host "  SUCCESS: Found progress for $mangaCount manga(s)" -ForegroundColor Green

    foreach ($mangaProp in $response.data.mangas.PSObject.Properties) {
        $manga = $mangaProp.Value
        $chCount = ($manga.chapters.PSObject.Properties | Measure-Object).Count
        Write-Host "    - Manga $($manga.mangaId): $chCount chapter(s)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 7
Write-Host "Test 7: Getting read manga list..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga-list" `
        -Method GET `
        -WebSession $session

    Write-Host "  SUCCESS: Found $($response.count) manga(s) in list" -ForegroundColor Green
    $response.data | ForEach-Object {
        Write-Host "    - $_" -ForegroundColor Gray
    }
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 8
Write-Host "Test 8: Updating chapter-1 progress to page 30..." -ForegroundColor Green
try {
    $body = @{
        mangaId = "test-manga-1"
        chapterId = "test-chapter-1"
        lastPage = 30
        totalPages = 50
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" -Method POST `
        -Headers @{"Content-Type" = "application/json"} `
        -Body $body `
        -WebSession $session

    Write-Host "  SUCCESS: Progress updated" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 9
Write-Host "Test 9: Verifying updated progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-1?mangaId=test-manga-1" `
        -Method GET `
        -WebSession $session

    Write-Host "  SUCCESS: Progress is at page $($response.data.lastPage)" -ForegroundColor Green
    if ($response.data.lastPage -eq 30) {
        Write-Host "  Update verified correctly!" -ForegroundColor Green
    }
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 10
Write-Host "Test 10: Deleting chapter-2 progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/chapter/test-chapter-2?mangaId=test-manga-1" `
        -Method DELETE `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 11
Write-Host "Test 11: Deleting all manga-2 progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress/manga/test-manga-2" `
        -Method DELETE `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 12
Write-Host "Test 12: Final state check..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" `
        -Method GET `
        -WebSession $session

    $mangaCount = ($response.data.mangas.PSObject.Properties | Measure-Object).Count
    Write-Host "  SUCCESS: $mangaCount manga(s) remaining" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 13
Write-Host "Test 13: Cleaning up - deleting all progress..." -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/reading-progress" `
        -Method DELETE `
        -WebSession $session

    Write-Host "  SUCCESS: $($response.message)" -ForegroundColor Green
}
catch {
    Write-Host "  FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "All Tests Completed!" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

