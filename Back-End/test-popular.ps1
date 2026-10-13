# Test script for popular manga endpoint
Write-Host "Testing /manga/popular endpoint" -ForegroundColor Cyan

# Test 1: Get popular manga with default parameters
Write-Host "`n[Test 1] GET /manga/popular (default parameters)" -ForegroundColor Yellow
$response1 = Invoke-RestMethod -Uri "http://localhost:3001/manga/popular" -Method GET
Write-Host "Result: $($response1.result)" -ForegroundColor Green
Write-Host "Total: $($response1.total)" -ForegroundColor Green
Write-Host "Returned: $($response1.data.Count) manga" -ForegroundColor Green
if ($response1.data.Count -gt 0) {
    Write-Host "`nFirst manga:" -ForegroundColor Magenta
    Write-Host "  ID: $($response1.data[0].id)"
    Write-Host "  Title: $($response1.data[0].title)"
    Write-Host "  Cover URL: $($response1.data[0].coverUrl)"
    Write-Host "  Status: $($response1.data[0].status)"
}

# Test 2: Get popular manga with custom limit
Write-Host "`n[Test 2] GET /manga/popular?limit=5" -ForegroundColor Yellow
$response2 = Invoke-RestMethod -Uri "http://localhost:3001/manga/popular?limit=5" -Method GET
Write-Host "Result: $($response2.result)" -ForegroundColor Green
Write-Host "Returned: $($response2.data.Count) manga (limit=5)" -ForegroundColor Green

# Test 3: Get popular manga ordered by rating
Write-Host "`n[Test 3] GET /manga/popular?orderBy=rating&limit=10" -ForegroundColor Yellow
$response3 = Invoke-RestMethod -Uri "http://localhost:3001/manga/popular?orderBy=rating&limit=10" -Method GET
Write-Host "Result: $($response3.result)" -ForegroundColor Green
Write-Host "Returned: $($response3.data.Count) manga (order by rating)" -ForegroundColor Green

# Test 4: Get popular manga with pagination
Write-Host "`n[Test 4] GET /manga/popular?limit=10&offset=10" -ForegroundColor Yellow
$response4 = Invoke-RestMethod -Uri "http://localhost:3001/manga/popular?limit=10&offset=10" -Method GET
Write-Host "Result: $($response4.result)" -ForegroundColor Green
Write-Host "Offset: $($response4.offset), Returned: $($response4.data.Count) manga" -ForegroundColor Green

Write-Host "`n✅ All tests completed!" -ForegroundColor Green
Write-Host "`nNote: Cover URLs are ready to use in frontend:" -ForegroundColor Cyan
Write-Host "  - All manga include coverUrl field" -ForegroundColor White
Write-Host "  - URLs point to MangaDex CDN (512px quality)" -ForegroundColor White
Write-Host "  - Direct display in <img> tags: <img src=`"{{coverUrl}}`" />" -ForegroundColor White

