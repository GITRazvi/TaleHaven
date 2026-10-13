# JWT Authentication Test Script
# Tests the new JWT-based authentication system

$baseUrl = "http://localhost:3001"
$testEmail = "jwt-test-$(Get-Random)@example.com"
$testPassword = "TestPassword123"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "JWT Authentication Test" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Register
Write-Host "Test 1: Register new user" -ForegroundColor Yellow
$registerBody = @{
    email = $testEmail
    password = $testPassword
} | ConvertTo-Json

try {
    $registerResponse = Invoke-WebRequest -Uri "$baseUrl/auth/register" `
        -Method POST `
        -Body $registerBody `
        -ContentType "application/json" `
        -SessionVariable session `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[OK] Registration successful" -ForegroundColor Green
    Write-Host "Response: $($registerResponse.Content)" -ForegroundColor Gray

    # Check for JWT cookie
    $jwtCookie = $session.Cookies.GetCookies($baseUrl) | Where-Object { $_.Name -eq "talehaven_auth_token" }
    if ($jwtCookie) {
        Write-Host "[OK] JWT token set in cookie" -ForegroundColor Green
        Write-Host "Token (first 50 chars): $($jwtCookie.Value.Substring(0, [Math]::Min(50, $jwtCookie.Value.Length)))..." -ForegroundColor Gray
    } else {
        Write-Host "[FAIL] JWT token NOT found in cookies" -ForegroundColor Red
    }
} catch {
    Write-Host "[FAIL] Registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Test 2: Get current user (authenticated)
Write-Host "Test 2: Get current user (should work with JWT)" -ForegroundColor Yellow
try {
    $meResponse = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[OK] Successfully retrieved current user" -ForegroundColor Green
    Write-Host "Response: $($meResponse.Content)" -ForegroundColor Gray
} catch {
    Write-Host "[FAIL] Failed to get current user: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 3: Logout
Write-Host "Test 3: Logout" -ForegroundColor Yellow
try {
    $logoutResponse = Invoke-WebRequest -Uri "$baseUrl/auth/logout" `
        -Method POST `
        -WebSession $session `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[OK] Logout successful" -ForegroundColor Green
    Write-Host "Response: $($logoutResponse.Content)" -ForegroundColor Gray
} catch {
    Write-Host "[FAIL] Logout failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 4: Try accessing protected resource after logout (should fail)
Write-Host "Test 4: Try accessing protected resource after logout (should fail)" -ForegroundColor Yellow
try {
    $meResponse2 = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[FAIL] Unexpectedly succeeded - should have failed!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "[OK] Correctly rejected (401 Unauthorized)" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] Wrong error code: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

Write-Host ""

# Test 5: Login
Write-Host "Test 5: Login with same credentials" -ForegroundColor Yellow
$loginBody = @{
    email = $testEmail
    password = $testPassword
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json" `
        -SessionVariable session2 `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[OK] Login successful" -ForegroundColor Green
    Write-Host "Response: $($loginResponse.Content)" -ForegroundColor Gray

    # Check for JWT cookie
    $jwtCookie2 = $session2.Cookies.GetCookies($baseUrl) | Where-Object { $_.Name -eq "talehaven_auth_token" }
    if ($jwtCookie2) {
        Write-Host "[OK] JWT token set in cookie" -ForegroundColor Green
    }
} catch {
    Write-Host "[FAIL] Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 6: Access protected resource after login (should work)
Write-Host "Test 6: Access protected resource after login (should work)" -ForegroundColor Yellow
try {
    $meResponse3 = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session2 `
        -UseBasicParsing `
        -ErrorAction Stop

    Write-Host "[OK] Successfully accessed protected resource" -ForegroundColor Green
    Write-Host "Response: $($meResponse3.Content)" -ForegroundColor Gray
} catch {
    Write-Host "[FAIL] Failed to access protected resource: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "JWT Authentication Test Complete" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan


