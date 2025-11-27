# Asset Caching Guide

Learn how Tracks-generated applications handle HTTP caching for optimal performance.

## Overview

Tracks uses **content-addressed asset serving** via [hashfs](https://github.com/benbjohnson/hashfs) to enable aggressive browser caching while ensuring users always get the latest assets when content changes.

## How It Works

### Content-Addressed URLs

When you reference assets in your templates:

```templ
<link rel="stylesheet" href={ "/assets/" + assets.CSSURL() }/>
<script src={ "/assets/" + assets.JSURL() } defer></script>
```

The `assets.CSSURL()` and `assets.JSURL()` functions return hashed filenames:

- `app.css` becomes `app.abc123def.css`
- `app.js` becomes `app.xyz789abc.js`

The hash is derived from the file content. When you change the CSS or JS, the hash changes, and browsers fetch the new version.

### Cache Headers

Tracks sets appropriate `Cache-Control` headers based on the URL pattern:

**Hashed assets** (URLs containing content hash):

```text
Cache-Control: public, max-age=31536000, immutable
```

This tells browsers and CDNs:

- `public` - Can be cached by shared caches (CDNs, proxies)
- `max-age=31536000` - Cache for 1 year (maximum practical duration)
- `immutable` - Never revalidate; the content will never change at this URL

**Why immutable?** Since the hash is derived from content, if the content changes, the URL changes. The browser will never request the old URL again because templates reference the new URL.

### ETag Support

hashfs automatically handles ETags:

1. Sets `ETag` header using the content hash
2. Handles `If-None-Match` conditional requests
3. Returns `304 Not Modified` when appropriate

This means even non-hashed assets benefit from conditional request support.

## Cache Middleware

The cache middleware (`internal/http/middleware/cache.go`) detects hashed assets using a regex pattern:

```go
var hashPattern = regexp.MustCompile(`\.[a-f0-9]{8,}\.[^.]+$`)
```

This matches filenames like:

- `app.abc12345.css`
- `htmx.def67890abc.js`
- `vendor.1234abcd5678.min.js`

## Compression

Tracks automatically compresses assets using gzip. The compression middleware handles:

- `text/html`
- `text/css`
- `text/javascript`
- `application/javascript`
- `application/json`
- `image/svg+xml`

Combined with caching, this means:

1. First request: Compressed response, cached for 1 year
2. Subsequent requests: Served from browser cache (no network request)

## Browser Behavior

### Normal Navigation

When users click links or navigate normally, browsers use cached assets without revalidation (thanks to `immutable`).

### Hard Refresh (Ctrl+F5)

Hard refresh bypasses cache and fetches fresh. The `immutable` directive doesn't prevent this - it's an explicit user action.

### Back/Forward Navigation

Browser history navigation uses cached content for fast page transitions.

## CDN Compatibility

The cache headers work with all major CDNs:

| CDN | Support |
|-----|---------|
| Cloudflare | Full - respects `immutable` |
| Fastly | Full |
| CloudFront | Full |
| Vercel Edge | Full |

For CDN deployment, no additional configuration is needed.

## Customization

### Adding Custom Assets

To add custom assets with caching:

1. Place files in `internal/assets/web/`
2. Use `assets.AssetURL("path/to/file.ext")` in templates

```templ
<img src={ "/assets/" + assets.AssetURL("images/logo.png") } alt="Logo"/>
```

The hash is automatically computed and included in the URL.

### Non-Hashed Assets

Some assets shouldn't use content hashing (e.g., `favicon.ico`, `robots.txt`). These:

- Don't get immutable caching
- Still benefit from ETag-based conditional requests
- Should be placed directly in the assets directory

## Verifying Cache Behavior

Test cache headers with curl:

```bash
# Check headers for a hashed asset
curl -I http://localhost:8080/assets/css/app.abc123.css

# Expected output includes:
# Cache-Control: public, max-age=31536000, immutable
```

Or use browser dev tools:

1. Open Network tab
2. Load a page
3. Check the `Cache-Control` header on asset requests
4. Refresh and verify assets show "from disk cache"

## Related Topics

- [Architecture Overview](/docs/guides/architecture-overview) - Overall application structure
- [Routing Guide](/docs/guides/routing-guide) - How routes are defined and handled
