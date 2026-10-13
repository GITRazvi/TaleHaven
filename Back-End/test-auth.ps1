# TaleHaven Authentication Test Script
# Tests all authentication endpoints

$baseUrl = "http://localhost:3001"
$testEmail = "test-$(Get-Random)@example.com"
$testPassword = "SecurePassword123!"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TaleHaven Authentication API Test" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# Test 1: Register a new user
Write-Host "[TEST 1] Registering new user: $testEmail" -ForegroundColor Yellow
try {
    $registerBody = @{
        email = $testEmail
        password = $testPassword
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$baseUrl/auth/register" `
        -Method POST `
        -Body $registerBody `
        -ContentType "application/json" `
        -SessionVariable session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Registration successful!" -ForegroundColor Green
    Write-Host "  User ID: $($result.user.id)" -ForegroundColor Gray
    Write-Host "  Email: $($result.user.email)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Registration failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Get current user (should be logged in after registration)
Write-Host "`n[TEST 2] Getting current user info" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Successfully retrieved user info!" -ForegroundColor Green
    Write-Host "  User ID: $($result.user.id)" -ForegroundColor Gray
    Write-Host "  Email: $($result.user.email)" -ForegroundColor Gray
    Write-Host "  Created At: $($result.user.created_at)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Get user info failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Logout
Write-Host "`n[TEST 3] Logging out" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/logout" `
        -Method POST `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Logout successful!" -ForegroundColor Green
    Write-Host "  Message: $($result.message)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Logout failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Try to access protected route (should fail - not logged in)
Write-Host "`n[TEST 4] Attempting to access protected route (should fail)" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -ErrorAction Stop

    Write-Host "✗ This should have failed!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ Correctly blocked unauthenticated access!" -ForegroundColor Green
        Write-Host "  Status: 401 Unauthorized" -ForegroundColor Gray
    } else {
        Write-Host "✗ Unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 5: Login
Write-Host "`n[TEST 5] Logging in with credentials" -ForegroundColor Yellow
try {
    $loginBody = @{
        email = $testEmail
        password = $testPassword
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$baseUrl/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json" `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Login successful!" -ForegroundColor Green
    Write-Host "  Message: $($result.message)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 6: Get current user again (should work now)
Write-Host "`n[TEST 6] Getting current user info after login" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Successfully retrieved user info!" -ForegroundColor Green
    Write-Host "  User ID: $($result.user.id)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Get user info failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Change password
Write-Host "`n[TEST 7] Changing password" -ForegroundColor Yellow
$newPassword = "NewSecurePassword456!"
try {
    $passwordBody = @{
        old_password = $testPassword
        new_password = $newPassword
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$baseUrl/auth/password" `
        -Method PUT `
        -Body $passwordBody `
        -ContentType "application/json" `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Password changed successfully!" -ForegroundColor Green
    Write-Host "  Message: $($result.message)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Password change failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 8: Try to login with old password (should fail)
Write-Host "`n[TEST 8] Attempting login with old password (should fail)" -ForegroundColor Yellow
try {
    $loginBody = @{
        email = $testEmail
        password = $testPassword
    } | ConvertTo-Json

    # Create a new session for this test
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json" `
        -SessionVariable newSession `
        -ErrorAction Stop

    Write-Host "✗ Login with old password should have failed!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ Correctly rejected old password!" -ForegroundColor Green
        Write-Host "  Status: 401 Unauthorized" -ForegroundColor Gray
    } else {
        Write-Host "✗ Unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 9: Login with new password
Write-Host "`n[TEST 9] Logging in with new password" -ForegroundColor Yellow
try {
    $loginBody = @{
        email = $testEmail
        password = $newPassword
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$baseUrl/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json" `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Login with new password successful!" -ForegroundColor Green
} catch {
    Write-Host "✗ Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 10: Delete account
Write-Host "`n[TEST 10] Deleting user account" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/account" `
        -Method DELETE `
        -WebSession $session `
        -ErrorAction Stop

    $result = $response.Content | ConvertFrom-Json
    Write-Host "✓ Account deleted successfully!" -ForegroundColor Green
    Write-Host "  Message: $($result.message)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Account deletion failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 11: Try to access user info after deletion (should fail)
Write-Host "`n[TEST 11] Attempting to access deleted account (should fail)" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/auth/me" `
        -Method GET `
        -WebSession $session `
        -ErrorAction Stop

    Write-Host "✗ This should have failed!" -ForegroundColor Red
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ Correctly blocked access to deleted account!" -ForegroundColor Green
        Write-Host "  Status: 401 Unauthorized" -ForegroundColor Gray
    } else {
        Write-Host "✗ Unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 12: Test duplicate registration (should fail)
Write-Host "`n[TEST 12] Testing duplicate email registration (should fail)" -ForegroundColor Yellow
try {
    # First register a user
    $uniqueEmail = "duplicate-test-$(Get-Random)@example.com"
    $registerBody = @{
        email = $uniqueEmail
        password = "password123"
    } | ConvertTo-Json

    $response1 = Invoke-WebRequest -Uri "$baseUrl/auth/register" `
        -Method POST `
        -Body $registerBody `
        -ContentType "application/json" `
        -SessionVariable tempSession `
        -ErrorAction Stop

    # Try to register again with same email
    $response2 = Invoke-WebRequest -Uri "$baseUrl/auth/register" `
        -Method POST `
        -Body $registerBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "✗ Duplicate registration should have failed!" -ForegroundColor Red

    # Cleanup: delete the test user
    $loginBody = @{
        email = $uniqueEmail
        password = "password123"
    } | ConvertTo-Json

    Invoke-WebRequest -Uri "$baseUrl/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json" `
        -WebSession $tempSession `
        -ErrorAction SilentlyContinue | Out-Null

    Invoke-WebRequest -Uri "$baseUrl/auth/account" `
        -Method DELETE `
        -WebSession $tempSession `
        -ErrorAction SilentlyContinue | Out-Null

} catch {
    if ($_.Exception.Response.StatusCode -eq 409) {
        Write-Host "✓ Correctly rejected duplicate email!" -ForegroundColor Green
        Write-Host "  Status: 409 Conflict" -ForegroundColor Gray

        # Cleanup: delete the test user
        $loginBody = @{
            email = $uniqueEmail
            password = "password123"
        } | ConvertTo-Json

        Invoke-WebRequest -Uri "$baseUrl/auth/login" `
            -Method POST `
            -Body $loginBody `
            -ContentType "application/json" `
            -WebSession $tempSession `
            -ErrorAction SilentlyContinue | Out-Null

        Invoke-WebRequest -Uri "$baseUrl/auth/account" `
            -Method DELETE `
            -WebSession $tempSession `
            -ErrorAction SilentlyContinue | Out-Null
    } else {
        Write-Host "✗ Unexpected error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

