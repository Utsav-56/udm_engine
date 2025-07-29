# Changelog

## Overview

This version introduces significant improvements to the server extraction logic in `ServerHEader.go`. The main focus is on enhancing reliability and robustness, especially in real-world network conditions.

---

## Key Changes

### 1. Retry Mechanism

The `getServerData` function now includes a retry mechanism to handle transient network or transport errors.

- **Retry Details:**
    - **Total Attempts:** 3 (initial attempt + 2 retries)
    - **Delay Between Retries:** Brief pause using `time.Sleep`
    - **Selective Retries:** Retries are only performed for recoverable network/transport errors (not for HTTP 404s or logical issues).

### 2. Additional Enhancements

- **HEAD then GET:**  
   The function attempts a `HEAD` request first and falls back to `GET` if necessary.
- **RFC 6266 Filename Handling:**  
   Properly parses and handles filenames from the `Content-Disposition` header.
- **Redirect Tracking:**  
   Follows and tracks HTTP redirects to ensure the final resource is reached.
- **Range Support:**  
   Checks if the server accepts range requests.
- **Final URL Reporting:**  
   Includes the final URL after all redirects.
- **MIME Type Fallback:**  
   Uses a basic map for MIME type fallback if the server does not provide one.

---

## Summary of Improvements

| Feature           | Status       |
| ----------------- | ------------ |
| Retry logic       | ✅ 3 tries   |
| HEAD then GET     | ✅ fallback  |
| RFC 6266 filename | ✅ handled   |
| Redirects         | ✅ tracked   |
| Accepts ranges    | ✅ checked   |
| Final URL         | ✅ included  |
| MIME fallback     | ✅ basic map |

---

These enhancements make the server extraction logic more robust, reliable, and compliant with best practices.
