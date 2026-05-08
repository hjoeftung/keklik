# Minimal GDPR Technical Security Requirements

## 1. Transport security

-   Use HTTPS everywhere
-   Enforce TLS (no HTTP fallback)
-   Validate certificates properly

## 2. Authentication & access control

-   Users can only access their own data
-   Caregivers only access explicitly shared data
-   Verify identity on every request
-   Perform authorization checks on every endpoint

## 3. Data storage protection

-   Encryption at rest (DB/disk level is OK)
-   No public database access
-   Secrets not stored in code repository
-   Use environment variables or a secrets manager

## 4. Basic key handling

-   Do not hardcode encryption keys
-   Store keys separately from data
-   Rotate keys if compromised (optional)

## 5. Data minimization

-   Only store necessary fields
-   Avoid logging sensitive data (names, birth dates, sleep data)

## 6. Secure API design

-   Validate all inputs
-   Prevent IDOR vulnerabilities
-   Use UUIDs instead of incremental IDs

## 7. User data deletion

-   Provide "Delete account" functionality
-   Ensure real deletion (not just soft delete)
-   Apply deletion to backups eventually

## 8. Backup security

-   Encrypt backups
-   Restrict access to backups
-   Avoid public storage

## 9. Breach readiness

-   Have basic logging and monitoring
-   Be able to detect and investigate breaches

## 10. Third-party services

-   Use trusted providers
-   Configure them securely
-   Disclose them in privacy policy

## 11. Device-side security

-   Avoid storing sensitive data in plain text
-   Use secure storage for tokens
-   Prevent data leaks via logs

## 12. No unnecessary tracking

-   Avoid analytics or ad SDKs that collect personal data
-   Disable unnecessary data collection
